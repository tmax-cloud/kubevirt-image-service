package virtualmachineimage

import (
	"context"
	goerrors "errors"
	snapshotv1beta1 "github.com/kubernetes-csi/external-snapshotter/v2/pkg/apis/volumesnapshot/v1beta1"
	hc "github.com/tmax-cloud/kubevirt-image-service/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func (r *ReconcileVirtualMachineImage) syncSnapshot() error {
	imported, found, err := r.isPvcImported()
	if err != nil {
		return err
	} else if !found {
		return nil
	}

	snapshot := &snapshotv1beta1.VolumeSnapshot{}
	err = r.Client.Get(context.TODO(), types.NamespacedName{Namespace: r.vmi.Namespace, Name: GetSnapshotNameFromVmiName(r.vmi.Name)}, snapshot)
	if err != nil && !errors.IsNotFound(err) {
		return err
	}
	existsSnapshot := err == nil

	if imported && !existsSnapshot {
		// 임포트 되어 있는데 스냅샷이 없으므로 만든다
		klog.Infof("Create a new snapshot for vmi %s", r.vmi.Name)
		newSnapshot, err := newSnapshot(r.vmi, r.Scheme)
		if err != nil {
			return err
		}
		if err := r.Client.Create(context.TODO(), newSnapshot); err != nil && !errors.IsAlreadyExists(err) {
			return err
		}
	} else if imported && existsSnapshot && snapshot.Status != nil {
		if snapshot.Status.Error != nil {
			return goerrors.New("Snapshot is error for vmi " + r.vmi.Name)
		} else if *snapshot.Status.ReadyToUse {
			// 임포트 되어 있고 스냅샷도 있다면 스냅샷의 readyToUse에 따라 상태를 변경한다.
			if err := r.updateStateWithReadyToUse(hc.VirtualMachineImageStateAvailable, corev1.ConditionTrue, "VmiIsReady", "Vmi is ready to use"); err != nil {
				return err
			}
		}
	} else if !imported && existsSnapshot {
		// 임포트 안되어 있는데 스냅샷이 있으므로 삭제한다
		if err := r.Client.Delete(context.TODO(), snapshot); err != nil && !errors.IsNotFound(err) {
			return err
		}
	}

	return nil
}

// GetSnapshotNameFromVmiName returns snapshot name of vminame
func GetSnapshotNameFromVmiName(vmiName string) string {
	return vmiName + "-image-snapshot"
}

func newSnapshot(vmi *hc.VirtualMachineImage, scheme *runtime.Scheme) (*snapshotv1beta1.VolumeSnapshot, error) {
	pvcName := GetPvcNameFromVmiName(vmi.Name)
	snapshot := &snapshotv1beta1.VolumeSnapshot{
		ObjectMeta: metav1.ObjectMeta{
			Name:      GetSnapshotNameFromVmiName(vmi.Name),
			Namespace: vmi.Namespace,
		},
		Spec: snapshotv1beta1.VolumeSnapshotSpec{
			Source: snapshotv1beta1.VolumeSnapshotSource{
				PersistentVolumeClaimName: &pvcName,
			},
			VolumeSnapshotClassName: &vmi.Spec.SnapshotClassName,
		},
	}
	if err := controllerutil.SetControllerReference(vmi, snapshot, scheme); err != nil {
		return nil, err
	}
	return snapshot, nil
}

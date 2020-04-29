package virtualmachineimage

import (
	"context"
	snapshotv1alpha1 "github.com/kubernetes-csi/external-snapshotter/pkg/apis/volumesnapshot/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func (r *ReconcileVirtualMachineImage) getSnapshot() (*snapshotv1alpha1.VolumeSnapshot, error) {
	snapshotName := GetSnapshotName(r.vmi.Name)
	snapshotNamespace := r.getNamespace()
	snapshot := &snapshotv1alpha1.VolumeSnapshot{}
	if err := r.client.Get(context.Background(), types.NamespacedName{Namespace: snapshotNamespace, Name: snapshotName}, snapshot); err != nil {
		return nil, err
	}
	r.log.Info("Get Snapshot", "snapshot", snapshot)
	return snapshot, nil
}

func (r *ReconcileVirtualMachineImage) createSnapshot() (*snapshotv1alpha1.VolumeSnapshot, error) {
	r.log.Info("Create new snapshot", "name", GetSnapshotName(r.vmi.Name), "namespace", r.getNamespace())
	snapshot, err := r.newSnapshot()
	if err != nil {
		return nil, err
	}
	if err := r.client.Create(context.Background(), snapshot); err != nil {
		return nil, err
	}
	return snapshot, nil
}

func (r *ReconcileVirtualMachineImage) deleteSnapshot() error {
	snapshotName := GetSnapshotName(r.vmi.Name)
	snapshotNamespace := r.getNamespace()
	r.log.Info("Delete snapshot", "name", snapshotName, "namespace", snapshotNamespace)
	snapshot := &snapshotv1alpha1.VolumeSnapshot{
		ObjectMeta: metav1.ObjectMeta{
			Name:      snapshotName,
			Namespace: snapshotNamespace,
		},
	}
	return r.client.Delete(context.Background(), snapshot)
}

// GetSnapshotName is called to create snapshot name
func GetSnapshotName(name string) string {
	return name + "-snapshot"
}

func (r *ReconcileVirtualMachineImage) newSnapshot() (*snapshotv1alpha1.VolumeSnapshot, error) {
	snapshotName := GetSnapshotName(r.vmi.Name)
	snapshotNamespace := r.getNamespace()
	pvcName := GetPvcName(r.vmi.Name, false)
	snapshot := &snapshotv1alpha1.VolumeSnapshot{
		// TODO: without typemeta
		TypeMeta: metav1.TypeMeta{
			Kind:       "VolumeSnapshot",
			APIVersion: "snapshot.storage.k8s.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      snapshotName,
			Namespace: snapshotNamespace,
		},
		Spec: snapshotv1alpha1.VolumeSnapshotSpec{
			Source: &corev1.TypedLocalObjectReference{
				Kind: "PersistentVolumeClaim",
				Name: pvcName,
			},
			VolumeSnapshotClassName: &r.vmi.Spec.SnapshotClassName,
		},
	}
	if err := controllerutil.SetControllerReference(r.vmi, snapshot, r.scheme); err != nil {
		return nil, err
	}
	return snapshot, nil
}

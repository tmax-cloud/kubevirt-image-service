package virtualmachineimage

import (
	"context"
	hc "github.com/tmax-cloud/kubevirt-image-service/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func (r *ReconcileVirtualMachineImage) syncScratchPvc() error {
	imported, found, err := r.isPvcImported()
	if err != nil {
		return err
	} else if !found {
		return nil
	}

	scratchPvc := &corev1.PersistentVolumeClaim{}
	err = r.Client.Get(context.TODO(), types.NamespacedName{Name: getScratchPvcNameFromVmiName(r.vmi.Name), Namespace: r.vmi.Namespace}, scratchPvc)
	if err != nil && !errors.IsNotFound(err) {
		return err
	}
	existsScratchPvc := err == nil

	if imported && existsScratchPvc {
		// 임포팅이 완료됐으니 삭제한다
		klog.Infof("Delete scratchPvc because importing completed vmi: %s", r.vmi.Name)
		if err := r.Client.Delete(context.TODO(), scratchPvc); err != nil && !errors.IsNotFound(err) {
			return err
		}
	} else if !imported && !existsScratchPvc {
		// 임포팅을 해야하므로 scratchPvc를 만든다
		klog.Infof("Create scratchPvc for importing vmi: %s", r.vmi.Name)
		newScratchPvc, err := newScratchPvc(r.vmi, r.Scheme)
		if err != nil {
			return err
		}
		if err := r.Client.Create(context.TODO(), newScratchPvc); err != nil && !errors.IsAlreadyExists(err) {
			return err
		}
	}
	return nil
}

func getScratchPvcNameFromVmiName(vmiName string) string {
	return vmiName + "-scratch-image-pvc"
}

func newScratchPvc(vmi *hc.VirtualMachineImage, scheme *runtime.Scheme) (*corev1.PersistentVolumeClaim, error) {
	volumeMode := corev1.PersistentVolumeFilesystem
	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      getScratchPvcNameFromVmiName(vmi.Name),
			Namespace: vmi.Namespace,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			VolumeMode:       &volumeMode,
			AccessModes:      []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
			Resources:        vmi.Spec.PVC.Resources,
			StorageClassName: vmi.Spec.PVC.StorageClassName,
		},
	}
	if err := controllerutil.SetControllerReference(vmi, pvc, scheme); err != nil {
		return nil, err
	}
	return pvc, nil
}

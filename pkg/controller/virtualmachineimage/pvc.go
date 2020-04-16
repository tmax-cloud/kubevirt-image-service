package virtualmachineimage

import (
	"context"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func (r *ReconcileVirtualMachineImage) getPvc(isScratch bool) error {
	pvcName := r.getPvcName(isScratch)
	pvcNamespace := r.getNamespace()
	pvc := &corev1.PersistentVolumeClaim{}
	if err := r.client.Get(context.Background(), types.NamespacedName{Name: pvcName, Namespace: pvcNamespace}, pvc); err != nil {
		return err
	}
	r.log.Info("Get pvc", "pvc", pvc)
	return nil
}

func (r *ReconcileVirtualMachineImage) createPvc(isScratch bool) error {
	r.log.Info("Create new pvc", "name", r.getPvcName(isScratch), "namespace", r.getNamespace())
	pvc, err := r.newPvc(isScratch)
	if err != nil {
		return err
	}
	if err := r.client.Create(context.Background(), pvc); err != nil {
		return err
	}
	return nil
}

func (r *ReconcileVirtualMachineImage) deletePvc(isScratch bool) error {
	pvcName := r.getPvcName(isScratch)
	pvcNamespace := r.getNamespace()
	r.log.Info("Delete pvc", "name", pvcName, "namespace", pvcNamespace)
	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      pvcName,
			Namespace: pvcNamespace,
		},
	}
	return r.client.Delete(context.Background(), pvc)
}

func (r *ReconcileVirtualMachineImage) getPvcName(isScratch bool) string {
	if isScratch {
		return r.vmi.Name + "-scratch-pvc"
	}
	return r.vmi.Name + "-pvc"
}

func (r *ReconcileVirtualMachineImage) getNamespace() string {
	return r.vmi.Namespace
}

func (r *ReconcileVirtualMachineImage) newPvc(isScratch bool) (*corev1.PersistentVolumeClaim, error) {
	pvcName := r.getPvcName(isScratch)
	pvcNamespace := r.getNamespace()
	pvc := &corev1.PersistentVolumeClaim{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PersistentVolumeClaim",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      pvcName,
			Namespace: pvcNamespace,
		},
		Spec: r.vmi.Spec.PVC,
	}
	if isScratch {
		scratchPvc := r.vmi.Spec.PVC.DeepCopy()
		volumeMode := corev1.PersistentVolumeFilesystem
		scratchPvc.VolumeMode = &volumeMode
		pvc.Spec = *scratchPvc
	} else
		if pvc.Spec.VolumeMode == nil {
			volumeMode := corev1.PersistentVolumeBlock
			pvc.Spec.VolumeMode = &volumeMode
		}

	if err := controllerutil.SetControllerReference(r.vmi, pvc, r.scheme); err != nil {
		return nil, err
	}
	return pvc, nil
}

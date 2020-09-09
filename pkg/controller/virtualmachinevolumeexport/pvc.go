package virtualmachinevolumeexport

import (
	"context"
	goerrors "errors"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog"
	hc "kubevirt-image-service/pkg/apis/hypercloud/v1alpha1"
	vmv "kubevirt-image-service/pkg/controller/virtualmachinevolume"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func (r *ReconcileVirtualMachineVolumeExport) syncExportPvc() error {
	if _, err := r.getPvc(GetExportPvcName(r.vmvExport.Name)); err == nil {
		return nil
	} else if !errors.IsNotFound(err) {
		return err
	}

	// get virtual machine volume pvc
	sourcePvc, err := r.getPvc(vmv.GetVolumePvcName(r.vmvExport.Spec.VirtualMachineVolume.Name))
	if err != nil {
		if err2 := r.updateStateWithReadyToUse(hc.VirtualMachineVolumeExportStatePending, corev1.ConditionFalse, "VmvExportIsPending", "VmvExport is in pending"); err2 != nil {
			return err2
		}
		//this is not exactly corret...
		return nil
	}

	klog.Infof("Create a new pvc for vmvExport %s", r.vmvExport.Name)
	if err2 := r.updateStateWithReadyToUse(hc.VirtualMachineVolumeExportStateCreating, corev1.ConditionFalse, "VmvExportIsCreating", "VmvExport is in creating"); err2 != nil {
		return err2
	}

	newPvc, err := newPvc(sourcePvc, r.vmvExport, r.scheme)
	if err != nil {
		return err
	}
	if err := r.client.Create(context.TODO(), newPvc); err != nil && !errors.IsAlreadyExists(err) {
		return err
	}
	return nil
}

func (r *ReconcileVirtualMachineVolumeExport) getPvc(pvcName string) (*corev1.PersistentVolumeClaim, error) {
	pvc := &corev1.PersistentVolumeClaim{}
	if err := r.client.Get(context.TODO(), types.NamespacedName{Namespace: r.vmvExport.Namespace, Name: pvcName}, pvc); err != nil {
		return nil, err
	}
	return pvc, nil
}

// GetExportPvcName is return pvcName for vmvExport name
func GetExportPvcName(vmvExportName string) string {
	return vmvExportName + "-export-pvc"
}

func newPvc(sourcePvc *corev1.PersistentVolumeClaim, vmvExport *hc.VirtualMachineVolumeExport, scheme *runtime.Scheme) (*corev1.PersistentVolumeClaim, error) {
	volumeMode := corev1.PersistentVolumeFilesystem
	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:        GetExportPvcName(vmvExport.Name),
			Namespace:   vmvExport.Namespace,
			Annotations: map[string]string{"completed": "no"},
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes:      []v1.PersistentVolumeAccessMode{"ReadWriteOnce"},
			VolumeMode:       &volumeMode,
			StorageClassName: sourcePvc.Spec.StorageClassName,
			Resources:        sourcePvc.Spec.Resources,
		},
	}
	if err := controllerutil.SetControllerReference(vmvExport, pvc, scheme); err != nil {
		return nil, err
	}
	return pvc, nil
}

func (r *ReconcileVirtualMachineVolumeExport) isPvcExportCompleted() (completed, found bool, err error) {
	pvc, err := r.getPvc(GetExportPvcName(r.vmvExport.Name))
	if err != nil {
		if errors.IsNotFound(err) {
			return false, false, nil
		}
		return false, false, err
	}
	i, found := pvc.Annotations["completed"]
	if !found {
		return false, true, goerrors.New("invalid pvc annotation: Must need 'completed'")
	}
	return i == "yes", true, nil
}

func (r *ReconcileVirtualMachineVolumeExport) updatePvcCompleted(completed bool) error {
	pvc, err := r.getPvc(GetExportPvcName(r.vmvExport.Name))
	if err != nil {
		return err
	}
	if pvc.Annotations == nil {
		pvc.Annotations = map[string]string{}
	}
	if completed {
		pvc.Annotations["completed"] = "yes"
	} else {
		pvc.Annotations["completed"] = "no"
	}
	return r.client.Update(context.TODO(), pvc)
}

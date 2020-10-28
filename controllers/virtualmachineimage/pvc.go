package virtualmachineimage

import (
	"context"
	goerrors "errors"
	hc "github.com/tmax-cloud/kubevirt-image-service/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func (r *ReconcileVirtualMachineImage) syncPvc() error {
	if _, err := r.getPvc(r.vmi); err == nil {
		return nil
	} else if !errors.IsNotFound(err) {
		return err
	}

	klog.Infof("Create a new pvc for vmi %s", r.vmi.Name)
	if err := r.updateStateWithReadyToUse(hc.VirtualMachineImageStateCreating, corev1.ConditionFalse, "VmiIsCreating", "VMI is in creating"); err != nil {
		return err
	}

	newPvc, err := newPvc(r.vmi, r.Scheme)
	if err != nil {
		return err
	}
	if err := r.Client.Create(context.TODO(), newPvc); err != nil && !errors.IsAlreadyExists(err) {
		return err
	}
	return nil
}

func (r *ReconcileVirtualMachineImage) getPvc(vmi *hc.VirtualMachineImage) (*corev1.PersistentVolumeClaim, error) {
	pvc := &corev1.PersistentVolumeClaim{}
	err := r.Client.Get(context.TODO(), types.NamespacedName{Name: GetPvcNameFromVmiName(vmi.Name), Namespace: vmi.Namespace}, pvc)
	if err != nil {
		return nil, err
	}
	return pvc, err
}

// GetPvcNameFromVmiName is return pvcName for vmiName
func GetPvcNameFromVmiName(vmiName string) string {
	return vmiName + "-image-pvc"
}

func newPvc(vmi *hc.VirtualMachineImage, scheme *runtime.Scheme) (*corev1.PersistentVolumeClaim, error) {
	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:        GetPvcNameFromVmiName(vmi.Name),
			Namespace:   vmi.Namespace,
			Annotations: map[string]string{"imported": "no"},
		},
		Spec: vmi.Spec.PVC,
	}
	if err := controllerutil.SetControllerReference(vmi, pvc, scheme); err != nil {
		return nil, err
	}
	return pvc, nil
}

func (r *ReconcileVirtualMachineImage) isPvcImported() (imported, found bool, err error) {
	pvc, err := r.getPvc(r.vmi)
	if err != nil {
		if errors.IsNotFound(err) {
			return false, false, nil
		}
		return false, false, err
	}
	i, found := pvc.Annotations["imported"]
	if !found {
		return false, true, goerrors.New("invalid pvc annotation: Must need 'imported'")
	}
	return i == "yes", true, nil
}

func (r *ReconcileVirtualMachineImage) updatePvcImported(imported bool) error {
	pvc, err := r.getPvc(r.vmi)
	if err != nil {
		return err
	}
	if pvc.Annotations == nil {
		pvc.Annotations = map[string]string{}
	}
	if imported {
		pvc.Annotations["imported"] = "yes"
	} else {
		pvc.Annotations["imported"] = "no"
	}
	return r.Client.Update(context.TODO(), pvc)
}

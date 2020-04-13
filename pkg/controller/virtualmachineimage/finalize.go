package virtualmachineimage

import (
	"context"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	hypercloudv1alpha1 "kubevirt-image-service/pkg/apis/hypercloud/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const virtualMachineImageFinalizer = "virtualmachineimages.hypercloud.tmaxanc.com"

func (r *ReconcileVirtualMachineImage) addFinalizer(vmi *hypercloudv1alpha1.VirtualMachineImage) error {
	if !contains(vmi.GetFinalizers(), virtualMachineImageFinalizer) {
		vmi.SetFinalizers(append(vmi.GetFinalizers(), virtualMachineImageFinalizer))
		return r.client.Update(context.Background(), vmi)
	}
	return nil
}

func contains(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}

func isVmiToBeDeleted(vmi *hypercloudv1alpha1.VirtualMachineImage) bool {
	return vmi.GetDeletionTimestamp() != nil
}

func (r *ReconcileVirtualMachineImage) finalize(reqLogger logr.Logger, vmi *hypercloudv1alpha1.VirtualMachineImage) (reconcile.Result, error) {
	// Delete PVC
	pvc := &corev1.PersistentVolumeClaim{}
	if err := r.deleteIfExists(vmi.Name+"-pvc", vmi.Namespace, pvc); err != nil {
		return reconcile.Result{}, err
	}

	// Delete finalizer
	if contains(vmi.GetFinalizers(), virtualMachineImageFinalizer) {
		vmi.SetFinalizers(remove(vmi.GetFinalizers(), virtualMachineImageFinalizer))
		if err := r.client.Update(context.TODO(), vmi); err != nil {
			return reconcile.Result{}, err
		}
	}

	reqLogger.Info("Successfully finalized VirtualMachineImage")
	return reconcile.Result{}, nil
}

func (r *ReconcileVirtualMachineImage) deleteIfExists(name, namespace string, obj runtime.Object) error {
	if err := r.client.Get(context.Background(), types.NamespacedName{Name: name, Namespace: namespace}, obj); err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		return err
	}
	return r.client.Delete(context.Background(), obj)
}

func remove(list []string, s string) []string {
	for i, v := range list {
		if v == s {
			list = append(list[:i], list[i+1:]...)
		}
	}
	return list
}

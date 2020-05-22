package virtualmachinevolume

import (
	"context"
	goerrors "errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog"
	hc "kubevirt-image-service/pkg/apis/hypercloud/v1alpha1"
	img "kubevirt-image-service/pkg/controller/virtualmachineimage"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func (r *ReconcileVirtualMachineVolume) syncVolumePvc() error {
	pvc := &corev1.PersistentVolumeClaim{}
	err := r.client.Get(context.Background(), types.NamespacedName{Name: GetVolumePvcName(r.volume.Name),
		Namespace: r.volume.Namespace}, pvc)
	if err != nil && !errors.IsNotFound(err) {
		return nil
	}
	pvcExists := err == nil

	if pvcExists {
		if pvc.Status.Phase == corev1.ClaimBound {
			if err := r.updateStateWithReadyToUse(hc.VirtualMachineVolumeStateAvailable, corev1.ConditionTrue, "SuccessfulCreate", "VirtualMachineVolume is available"); err != nil {
				return err
			}
		} else if pvc.Status.Phase == corev1.ClaimLost {
			return goerrors.New("PVC is lost")
		}
	} else {
		klog.Infof("Create a new pvc for volume %s", r.volume.Name)
		if err := r.updateStateWithReadyToUse(hc.VirtualMachineVolumeStateCreating, corev1.ConditionFalse, "CreatingPVC", "VirtualMachineVolume is creating PVC"); err != nil {
			return err
		}

		_, err := r.createVolumePvc()
		if err != nil {
			return err
		}
	}
	return nil
}

// createVolumePvc creates pvc from volumeSnapShot created by virtualMachineImage
func (r *ReconcileVirtualMachineVolume) createVolumePvc() (*corev1.PersistentVolumeClaim, error) {
	image := &hc.VirtualMachineImage{}
	if err := r.client.Get(context.Background(), types.NamespacedName{Name: r.volume.Spec.VirtualMachineImage.Name, Namespace: r.volume.Namespace}, image); err != nil {
		return nil, err
	}

	apiGroup := "snapshot.storage.k8s.io"
	pvc := &corev1.PersistentVolumeClaim{
		TypeMeta: v1.TypeMeta{
			Kind:       "PersistentVolumeClaim",
			APIVersion: "v1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      GetVolumePvcName(r.volume.Name),
			Namespace: r.volume.Namespace,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			StorageClassName: image.Spec.PVC.StorageClassName,
			AccessModes:      image.Spec.PVC.AccessModes,
			VolumeMode:       image.Spec.PVC.VolumeMode,
			DataSource: &corev1.TypedLocalObjectReference{
				APIGroup: &apiGroup,
				Kind:     "VolumeSnapshot",
				Name:     img.GetSnapshotNameFromVmiName(image.Name),
			},
			Resources: corev1.ResourceRequirements{
				Requests: r.volume.Spec.Capacity,
			},
		},
	}
	if err := controllerutil.SetControllerReference(r.volume, pvc, r.scheme); err != nil {
		return nil, err
	}
	if err := r.client.Create(context.Background(), pvc); err != nil {
		return nil, err
	}

	return pvc, nil
}

// GetVolumePvcName gets the name of the pvc created by virtualMachineVolume
func GetVolumePvcName(volumeName string) string {
	return volumeName + "-vmv-pvc"
}

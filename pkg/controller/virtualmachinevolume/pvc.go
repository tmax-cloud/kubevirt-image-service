package virtualmachinevolume

import (
	"context"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	hc "kubevirt-image-service/pkg/apis/hypercloud/v1alpha1"
	img "kubevirt-image-service/pkg/controller/virtualmachineimage"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func (r *ReconcileVirtualMachineVolume) getRestoredPvc() (*corev1.PersistentVolumeClaim, error) {
	pvc := &corev1.PersistentVolumeClaim{}
	if err := r.client.Get(context.Background(), types.NamespacedName{Name: GetRestoredPvcName(r.volume.Name),
		Namespace: r.volume.Namespace}, pvc); err != nil {
		return nil, err
	}
	return pvc, nil
}

// restorePvc restores pvc from volumeSnapShot created by virtualMachineImage
func (r *ReconcileVirtualMachineVolume) restorePvc(image *hc.VirtualMachineImage) (*corev1.PersistentVolumeClaim, error) {
	apiGroup := "snapshot.storage.k8s.io"
	pvc := &corev1.PersistentVolumeClaim{
		TypeMeta: v1.TypeMeta{
			Kind:       "PersistentVolumeClaim",
			APIVersion: "v1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      GetRestoredPvcName(r.volume.Name),
			Namespace: r.volume.Namespace,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			StorageClassName: image.Spec.PVC.StorageClassName,
			AccessModes:      image.Spec.PVC.AccessModes,
			VolumeMode:       image.Spec.PVC.VolumeMode,
			DataSource: &corev1.TypedLocalObjectReference{
				APIGroup: &apiGroup,
				Kind:     "VolumeSnapshot",
				Name:     img.GetSnapshotName(image.Name),
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

// GetRestoredPvcName gets the name of the restored pvc from volumeSnapShot created by virtualMachineVolume
func GetRestoredPvcName(volumeName string) string {
	return volumeName + "-vmv-pvc"
}

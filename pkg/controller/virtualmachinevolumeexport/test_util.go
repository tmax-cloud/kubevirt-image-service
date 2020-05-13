package virtualmachinevolumeexport

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	hc "kubevirt-image-service/pkg/apis/hypercloud/v1alpha1"
	img "kubevirt-image-service/pkg/controller/virtualmachineimage"
	vmv "kubevirt-image-service/pkg/controller/virtualmachinevolume"
	"kubevirt-image-service/pkg/util"
)

const (
	vmiName          = "testvmi"
	vmvName          = "testvmv"
	vmvExportName    = "testvmvexport"
	defaultNamespace = "default"
)

func createFakeReconcileVmvExport(objects ...runtime.Object) *ReconcileVirtualMachineVolumeExport {
	vmvExport := newTestVmvExport()
	client, scheme, err := util.CreateFakeClientAndScheme(append(objects, vmvExport)...)
	if err != nil {
		panic(err)
	}
	return &ReconcileVirtualMachineVolumeExport{client: client, scheme: scheme, vmvExport: vmvExport}
}

func newTestVmvExport() *hc.VirtualMachineVolumeExport {
	local := hc.VirtualMachineVolumeExportDestinationLocal{}
	return &hc.VirtualMachineVolumeExport{
		ObjectMeta: metav1.ObjectMeta{
			Name:      vmvExportName,
			Namespace: defaultNamespace,
		},
		Spec: hc.VirtualMachineVolumeExportSpec{
			VirtualMachineVolume: hc.VirtualMachineVolumeSource{
				Name: vmvName,
			},
			Destination: hc.VirtualMachineVolumeExportDestination{
				Local: &local,
			},
		},
	}
}

func newVmvPvc() *corev1.PersistentVolumeClaim {
	apiGroup := "snapshot.storage.k8s.io"
	storageClass := "rook-ceph-block"
	volumeMode := corev1.PersistentVolumeBlock

	return &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      vmv.GetVolumePvcName(vmvName),
			Namespace: defaultNamespace,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			StorageClassName: &storageClass,
			AccessModes: []corev1.PersistentVolumeAccessMode{
				corev1.ReadWriteOnce,
			},
			VolumeMode: &volumeMode,
			DataSource: &corev1.TypedLocalObjectReference{
				APIGroup: &apiGroup,
				Kind:     "VolumeSnapshot",
				Name:     img.GetSnapshotNameFromVmiName(vmiName),
			},
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse("3Gi"),
				},
			},
		},
	}
}

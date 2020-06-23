package virtualmachinevolume

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	hc "kubevirt-image-service/pkg/apis/hypercloud/v1alpha1"
	img "kubevirt-image-service/pkg/controller/virtualmachineimage"
	"kubevirt-image-service/pkg/util"
)

const (
	testVolumeName = "myvmv"
	testImageName  = "myvmi"
	testNameSpace  = "mynamespace"
)

var (
	testStorageClassName = "mystorageclass"
	testVolumeNamespacedName = types.NamespacedName{Name: testVolumeName, Namespace: testNameSpace}
)

func createFakeReconcileVmv(objects ...runtime.Object) *ReconcileVirtualMachineVolume {
	v := newTestVolume()
	client, scheme, err := util.CreateFakeClientAndScheme(append(objects, v)...)
	if err != nil {
		panic(err)
	}
	return &ReconcileVirtualMachineVolume{client: client, scheme: scheme, volume: v}
}

func createFakeReconcileVolumeWithImage(objects ...runtime.Object) *ReconcileVirtualMachineVolume {
	v := newTestVolume()
	i := newTestImage()
	i.Status.Conditions = util.SetConditionByType(i.Status.Conditions, hc.ConditionReadyToUse, corev1.ConditionTrue, "VmiIsReady", "Vmi is ready to use")
	client, scheme, err := util.CreateFakeClientAndScheme(append(objects, v, i)...)
	if err != nil {
		panic(err)
	}
	return &ReconcileVirtualMachineVolume{client: client, scheme: scheme, volume: v}
}

func newTestVolume() *hc.VirtualMachineVolume {
	return &hc.VirtualMachineVolume{
		ObjectMeta: v1.ObjectMeta{
			Name:      testVolumeName,
			Namespace: testNameSpace,
		},
		Spec: hc.VirtualMachineVolumeSpec{
			VirtualMachineImage: hc.VirtualMachineImageName{
				Name: testImageName,
			},
			Capacity: corev1.ResourceList{
				corev1.ResourceStorage: resource.MustParse("3Gi"),
			},
		},
	}
}

func newTestPvc() *corev1.PersistentVolumeClaim {
	apiGroup := "snapshot.storage.k8s.io"
	volumeMode := corev1.PersistentVolumeBlock

	return &corev1.PersistentVolumeClaim{
		ObjectMeta: v1.ObjectMeta{
			Name:      GetVolumePvcName(testVolumeName),
			Namespace: testNameSpace,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			StorageClassName: &testStorageClassName,
			AccessModes: []corev1.PersistentVolumeAccessMode{
				corev1.ReadWriteOnce,
			},
			VolumeMode: &volumeMode,
			DataSource: &corev1.TypedLocalObjectReference{
				APIGroup: &apiGroup,
				Kind:     "VolumeSnapshot",
				Name:     img.GetSnapshotNameFromVmiName(testImageName),
			},
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse("3Gi"),
				},
			},
		},
	}
}

func newTestImage() *hc.VirtualMachineImage {
	return &hc.VirtualMachineImage{
		ObjectMeta: v1.ObjectMeta{
			Name:      testImageName,
			Namespace: testNameSpace,
		},
		Spec: hc.VirtualMachineImageSpec{
			Source: hc.VirtualMachineImageSource{
				HTTP: "https://kr.tmaxsoft.com/main.do",
			},
			PVC: corev1.PersistentVolumeClaimSpec{
				AccessModes: []corev1.PersistentVolumeAccessMode{
					corev1.ReadWriteOnce,
				},
				Resources: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceStorage: resource.MustParse("3Gi"),
					},
				},
			},
		},
	}
}

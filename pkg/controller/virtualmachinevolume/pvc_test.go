package virtualmachinevolume

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	hc "kubevirt-image-service/pkg/apis/hypercloud/v1alpha1"
	img "kubevirt-image-service/pkg/controller/virtualmachineimage"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("getRestoredPvc", func() {
	It("Should get a restored pvc, if volume has it", func() {
		pvc := getPvcStruct("myvmv")
		r := createFakeReconcileVolume(pvc)
		pvcFound, err := r.getRestoredPvc()
		Expect(err).ToNot(HaveOccurred())
		Expect(pvc.Spec).To(Equal(pvcFound.Spec))
	})

	It("Should not get a restored pvc, if volume does not have it", func() {
		r := createFakeReconcileVolume()
		_, err := r.getRestoredPvc()
		Expect(errors.IsNotFound(err)).To(BeTrue())
	})
})

var _ = Describe("restorePvc", func() {
	It("Should restore pvc from volumeSnapShot", func() {
		image := getImageStruct()
		r := createFakeReconcileVolume()
		pvcCreated, err := r.restorePvc(image)
		expectedAccessModes := []corev1.PersistentVolumeAccessMode{
			corev1.ReadWriteOnce,
		}
		expectedResources :=  corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceStorage: resource.MustParse("3Gi"),
			},
		}
		Expect(err).ToNot(HaveOccurred())
		Expect(expectedAccessModes).To(Equal(pvcCreated.Spec.AccessModes))
		Expect(expectedResources).To(Equal(pvcCreated.Spec.Resources))
	})
})

func getPvcStruct(volumeName string) *corev1.PersistentVolumeClaim {
	apiGroup := "snapshot.storage.k8s.io"
	storageClass := "rook-ceph-block"
	volumeMode := corev1.PersistentVolumeBlock

	return &corev1.PersistentVolumeClaim{
		TypeMeta: v1.TypeMeta{
			Kind:       "PersistentVolumeClaim",
			APIVersion: "v1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      GetRestoredPvcName(volumeName),
			Namespace: "mynamespace",
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
				Name:     img.GetSnapshotName("myvmi"),
			},
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse("3Gi"),
				},
			},
		},
	}
}

func getImageStruct() *hc.VirtualMachineImage {
	return &hc.VirtualMachineImage{
		TypeMeta: v1.TypeMeta{
			Kind:       "VirtualMachineImage",
			APIVersion: "hypercloud.tmaxanc.com/v1alpha1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      "myvmi",
			Namespace: "mynamespace",
		},
		Spec: hc.VirtualMachineImageSpec{
			Source: hc.VirtualMachineImageSource{
				HTTP: "https://kr.tmaxsoft.com/main.do",
			},
			PVC: corev1.PersistentVolumeClaimSpec{
				AccessModes:  []corev1.PersistentVolumeAccessMode{
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

func createFakeReconcileVolume(objects ...runtime.Object) *ReconcileVirtualMachineVolume {
	v := &hc.VirtualMachineVolume{
		TypeMeta: v1.TypeMeta{
			Kind:       "VirtualMachineVolume",
			APIVersion: "hypercloud.tmaxanc.com/v1alpha1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      "myvmv",
			Namespace: "mynamespace",
		},
		Spec: hc.VirtualMachineVolumeSpec{
			VirtualMachineImage: hc.VirtualMachineImageName{
				Name: "myvmi",
			},
			Capacity: corev1.ResourceList{
				corev1.ResourceStorage: resource.MustParse("3Gi"),
			},
		},
	}

	s := scheme.Scheme
	err := hc.SchemeBuilder.AddToScheme(s)
	Expect(err).ToNot(HaveOccurred())

	var objs []runtime.Object
	objs = append(objs, objects...)
	cl := fake.NewFakeClientWithScheme(s, objs...)

	return &ReconcileVirtualMachineVolume{
		client: cl,
		scheme: s,
		volume: v,
		log:    log.WithName("unit-test"),
	}
}
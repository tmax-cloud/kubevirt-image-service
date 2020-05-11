package virtualmachinevolume

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

var _ = Describe("Create volume pvc", func() {
	It("Should create pvc from volumeSnapShot", func() {
		image := newImage()
		r := newReconcileVolume()

		pvcCreated, err := r.createVolumePvc(image)
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

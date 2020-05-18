package virtualmachinevolume

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
)

var _ = Describe("getRestoredPvc", func() {
	It("Should get a restored pvc, if volume has it", func() {
		pvc := newPvc("myvmv")
		r := newReconcileVolume(pvc)
		pvcFound, err := r.getRestoredPvc()
		Expect(err).ToNot(HaveOccurred())
		Expect(pvc.Spec).To(Equal(pvcFound.Spec))
	})

	It("Should not get a restored pvc, if volume does not have it", func() {
		r := newReconcileVolume()
		_, err := r.getRestoredPvc()
		Expect(errors.IsNotFound(err)).To(BeTrue())
	})
})

var _ = Describe("restorePvc", func() {
	It("Should restore pvc from volumeSnapShot", func() {
		image := newImage()
		r := newReconcileVolume()
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

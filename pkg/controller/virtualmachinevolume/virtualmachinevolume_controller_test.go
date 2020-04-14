package virtualmachinevolume

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/errors"
)

var _ = Describe("getImage", func() {
	It("Should get a virtualMachineVolume, if it exists", func() {
		image := getImageStruct()
		r := createFakeReconcileVolume(image)
		imageFound, err := r.getImage()
		Expect(err).ToNot(HaveOccurred())
		Expect(image.Spec).To(Equal(imageFound.Spec))
	})

	It("Should not get a virtualMachineVolume, if it does not exists", func() {
		r := createFakeReconcileVolume()
		_, err := r.getImage()
		Expect(errors.IsNotFound(err)).To(BeTrue())
	})
})


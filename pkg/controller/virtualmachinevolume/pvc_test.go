package virtualmachinevolume

import (
	"context"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	hc "kubevirt-image-service/pkg/apis/hypercloud/v1alpha1"
	"kubevirt-image-service/pkg/util"
)

// #    pvc    pvc phase    volume condition    volume state
// 1	X
// 2	O	   bound		true       		    available
// 3	O	   lost
// 4	O	   pending

var _ = Describe("syncVolumePvc", func() {
	Context("1. with no pvc", func() {
		r := createFakeReconcileVolumeWithImage()
		err := r.syncVolumePvc()

		It("Should not return error", func() {
			Expect(err).Should(BeNil())
		})
		It("Should create pvc", func() {
			pvc := &corev1.PersistentVolumeClaim{}
			err := r.client.Get(context.TODO(), types.NamespacedName{Name: GetVolumePvcName(r.volume.Name),
				Namespace: r.volume.Namespace}, pvc)
			Expect(err).Should(BeNil())
		})
	})

	Context("2. with bound pvc", func() {
		pvc := newTestPvc()
		pvc.Status.Phase = corev1.ClaimBound
		r := createFakeReconcileVolumeWithImage(pvc)
		err := r.syncVolumePvc()

		It("Should not return error", func() {
			Expect(err).Should(BeNil())
		})
		It("Should update state to available", func() {
			volume := &hc.VirtualMachineVolume{}
			err = r.client.Get(context.TODO(), testVolumeNamespacedName, volume)
			Expect(err).Should(BeNil())
			Expect(volume.Status.State).Should(Equal(hc.VirtualMachineVolumeStateAvailable))
		})
		It("Should update condition readyToUse to true", func() {
			volume := &hc.VirtualMachineVolume{}
			err = r.client.Get(context.TODO(), testVolumeNamespacedName, volume)
			Expect(err).Should(BeNil())
			found, cond := util.GetConditionByType(volume.Status.Conditions, hc.VirtualMachineVolumeConditionReadyToUse)
			Expect(found).Should(BeTrue())
			Expect(cond.Status).Should(Equal(corev1.ConditionTrue))
		})
	})

	Context("3. with lost pvc", func() {
		pvc := newTestPvc()
		pvc.Status.Phase = corev1.ClaimLost
		r := createFakeReconcileVolumeWithImage(pvc)
		err := r.syncVolumePvc()

		It("Should return error", func() {
			Expect(err).ShouldNot(BeNil())
		})
	})

	Context("4. with pending pvc", func() {
		pvc := newTestPvc()
		pvc.Status.Phase = corev1.ClaimPending
		r := createFakeReconcileVolumeWithImage(pvc)
		err := r.syncVolumePvc()

		It("Should not return error", func() {
			Expect(err).Should(BeNil())
		})
	})
})

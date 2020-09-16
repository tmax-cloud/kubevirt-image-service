package virtualmachinevolume

import (
	"context"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	hc "kubevirt-image-service/pkg/apis/hypercloud/v1alpha1"
	"kubevirt-image-service/pkg/util"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var _ = Describe("Reconcile", func() {
	Context("1. with no image", func() {
		r := createFakeReconcileVmv()
		_, err := r.Reconcile(reconcile.Request{NamespacedName: testVolumeNamespacedName})

		It("Should be nil", func() {
			Expect(err).Should(BeNil())
		})
		It("Should not create pvc", func() {
			pvc := &corev1.PersistentVolumeClaim{}
			err = r.client.Get(context.TODO(), types.NamespacedName{Name: GetVolumePvcName(r.volume.Name),
				Namespace: r.volume.Namespace}, pvc)
			Expect(errors.IsNotFound(err)).Should(BeTrue())
		})
		It("Should update state to pending", func() {
			volume := &hc.VirtualMachineVolume{}
			err = r.client.Get(context.TODO(), testVolumeNamespacedName, volume)
			Expect(err).Should(BeNil())
			Expect(volume.Status.State).Should(Equal(hc.VirtualMachineVolumeStatePending))
		})
		It("Should update condition readyToUse to false", func() {
			volume := &hc.VirtualMachineVolume{}
			err = r.client.Get(context.TODO(), testVolumeNamespacedName, volume)
			Expect(err).Should(BeNil())
			// clear lastTransitionTime for easy comparison
			for i := range volume.Status.Conditions {
				volume.Status.Conditions[i].LastTransitionTime = v1.Time{}
			}
			found, cond := util.GetConditionByType(volume.Status.Conditions, hc.VirtualMachineVolumeConditionReadyToUse)
			Expect(found).Should(BeTrue())
			Expect(cond.Status).Should(Equal(corev1.ConditionFalse))
		})
	})

	Context("2. with false status image", func() {
		image := newTestImage()
		r := createFakeReconcileVmv(image)
		image.Status.Conditions = util.SetConditionByType(image.Status.Conditions, hc.ConditionReadyToUse, corev1.ConditionFalse, "VmiIsReady", "Vmi is ready to use")
		_, err := r.Reconcile(reconcile.Request{NamespacedName: testVolumeNamespacedName})

		It("Should be nil", func() {
			Expect(err).Should(BeNil())
		})
		It("Should not create pvc", func() {
			pvc := &corev1.PersistentVolumeClaim{}
			err = r.client.Get(context.TODO(), types.NamespacedName{Name: GetVolumePvcName(r.volume.Name),
				Namespace: r.volume.Namespace}, pvc)
			Expect(errors.IsNotFound(err)).Should(BeTrue())
		})
		It("Should update state to pending", func() {
			volume := &hc.VirtualMachineVolume{}
			err = r.client.Get(context.TODO(), testVolumeNamespacedName, volume)
			Expect(err).Should(BeNil())
			Expect(volume.Status.State).Should(Equal(hc.VirtualMachineVolumeStatePending))
		})
		It("Should update condition readyToUse to false", func() {
			volume := &hc.VirtualMachineVolume{}
			err = r.client.Get(context.TODO(), testVolumeNamespacedName, volume)
			Expect(err).Should(BeNil())
			// clear lastTransitionTime for easy comparison
			for i := range volume.Status.Conditions {
				volume.Status.Conditions[i].LastTransitionTime = v1.Time{}
			}
			found, cond := util.GetConditionByType(volume.Status.Conditions, hc.VirtualMachineVolumeConditionReadyToUse)
			Expect(found).Should(BeTrue())
			Expect(cond.Status).Should(Equal(corev1.ConditionFalse))
		})
	})

	Context("3. with nil status image", func() {
		image := newTestImage()
		r := createFakeReconcileVmv(image)
		_, err := r.Reconcile(reconcile.Request{NamespacedName: testVolumeNamespacedName})

		It("Should be nil", func() {
			Expect(err).Should(BeNil())
		})
		It("Should not create pvc", func() {
			pvc := &corev1.PersistentVolumeClaim{}
			err = r.client.Get(context.TODO(), types.NamespacedName{Name: GetVolumePvcName(r.volume.Name),
				Namespace: r.volume.Namespace}, pvc)
			Expect(errors.IsNotFound(err)).Should(BeTrue())
		})
		It("Should update state to pending", func() {
			volume := &hc.VirtualMachineVolume{}
			err = r.client.Get(context.TODO(), testVolumeNamespacedName, volume)
			Expect(err).Should(BeNil())
			Expect(volume.Status.State).Should(Equal(hc.VirtualMachineVolumeStatePending))
		})
		It("Should update condition readyToUse to false", func() {
			volume := &hc.VirtualMachineVolume{}
			err = r.client.Get(context.TODO(), testVolumeNamespacedName, volume)
			Expect(err).Should(BeNil())
			// clear lastTransitionTime for easy comparison
			for i := range volume.Status.Conditions {
				volume.Status.Conditions[i].LastTransitionTime = v1.Time{}
			}
			found, cond := util.GetConditionByType(volume.Status.Conditions, hc.VirtualMachineVolumeConditionReadyToUse)
			Expect(found).Should(BeTrue())
			Expect(cond.Status).Should(Equal(corev1.ConditionFalse))
		})
	})

	Context("4. with true status, invalid size image", func() {
		image := newTestImage()
		image.Spec.PVC.Resources.Requests[corev1.ResourceStorage] = resource.MustParse("7Gi")
		image.Status.Conditions = util.SetConditionByType(image.Status.Conditions, hc.ConditionReadyToUse, corev1.ConditionTrue, "VmiIsReady", "Vmi is ready to use")
		r := createFakeReconcileVmv(image)
		_, err := r.Reconcile(reconcile.Request{NamespacedName: testVolumeNamespacedName})

		It("Should return error", func() {
			Expect(err).ShouldNot(BeNil())
		})
		It("Should not create pvc", func() {
			pvc := &corev1.PersistentVolumeClaim{}
			err = r.client.Get(context.TODO(), types.NamespacedName{Name: GetVolumePvcName(r.volume.Name),
				Namespace: r.volume.Namespace}, pvc)
			Expect(errors.IsNotFound(err)).Should(BeTrue())
		})
		It("Should update state to error", func() {
			volume := &hc.VirtualMachineVolume{}
			err = r.client.Get(context.TODO(), testVolumeNamespacedName, volume)
			Expect(err).Should(BeNil())
			Expect(volume.Status.State).Should(Equal(hc.VirtualMachineVolumeStateError))
		})
		It("Should update condition readyToUse to false", func() {
			volume := &hc.VirtualMachineVolume{}
			err = r.client.Get(context.TODO(), testVolumeNamespacedName, volume)
			Expect(err).Should(BeNil())
			for i := range volume.Status.Conditions {
				volume.Status.Conditions[i].LastTransitionTime = v1.Time{}
			}
			found, cond := util.GetConditionByType(volume.Status.Conditions, hc.VirtualMachineVolumeConditionReadyToUse)
			Expect(found).Should(BeTrue())
			Expect(cond.Status).Should(Equal(corev1.ConditionFalse))
		})
	})

	Context("5. with true status, valid size image", func() {
		r := createFakeReconcileVolumeWithImage()
		_, err := r.Reconcile(reconcile.Request{NamespacedName: testVolumeNamespacedName})

		It("Should not return error", func() {
			Expect(err).Should(BeNil())
		})
		It("Should create pvc", func() {
			pvc := &corev1.PersistentVolumeClaim{}
			err = r.client.Get(context.TODO(), types.NamespacedName{Name: GetVolumePvcName(r.volume.Name),
				Namespace: r.volume.Namespace}, pvc)
			Expect(err).Should(BeNil())
		})
		It("Should update state to creating", func() {
			volume := &hc.VirtualMachineVolume{}
			err = r.client.Get(context.TODO(), testVolumeNamespacedName, volume)
			Expect(err).Should(BeNil())
			Expect(volume.Status.State).Should(Equal(hc.VirtualMachineVolumeStateCreating))
		})
		It("Should update condition readyToUse to false", func() {
			volume := &hc.VirtualMachineVolume{}
			err = r.client.Get(context.TODO(), testVolumeNamespacedName, volume)
			Expect(err).Should(BeNil())
			// clear lastTransitionTime for easy comparison
			for i := range volume.Status.Conditions {
				volume.Status.Conditions[i].LastTransitionTime = v1.Time{}
			}
			found, cond := util.GetConditionByType(volume.Status.Conditions, hc.VirtualMachineVolumeConditionReadyToUse)
			Expect(found).Should(BeTrue())
			Expect(cond.Status).Should(Equal(corev1.ConditionFalse))
		})
	})

	Context("6. with valid image, lost phase pvc", func() {
		pvc := newTestPvc()
		pvc.Status.Phase = corev1.ClaimLost
		r := createFakeReconcileVolumeWithImage(pvc)
		_, err := r.Reconcile(reconcile.Request{NamespacedName: testVolumeNamespacedName})

		It("Should return error", func() {
			Expect(err).ShouldNot(BeNil())
		})
		It("Should update state to error", func() {
			vmv := &hc.VirtualMachineVolume{}
			err = r.client.Get(context.TODO(), testVolumeNamespacedName, vmv)
			Expect(err).Should(BeNil())
			Expect(vmv.Status.State).Should(Equal(hc.VirtualMachineVolumeStateError))
		})
		It("Should update condition readyToUse to false", func() {
			vmv := &hc.VirtualMachineVolume{}
			err = r.client.Get(context.TODO(), testVolumeNamespacedName, vmv)
			Expect(err).Should(BeNil())
			// clear lastTransitionTime for easy comparison
			for i := range vmv.Status.Conditions {
				vmv.Status.Conditions[i].LastTransitionTime = v1.Time{}
			}
			found, cond := util.GetConditionByType(vmv.Status.Conditions, hc.VirtualMachineVolumeConditionReadyToUse)
			Expect(found).Should(BeTrue())
			Expect(cond.Status).Should(Equal(corev1.ConditionFalse))
		})
	})
})

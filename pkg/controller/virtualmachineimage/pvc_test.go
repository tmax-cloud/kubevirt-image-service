package virtualmachineimage

import (
	"context"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	hc "kubevirt-image-service/pkg/apis/hypercloud/v1alpha1"
)

// 번호		pvc
// 1		X
// 2		O
var _ = Describe("syncPvc", func() {
	Context("1. with no pvc", func() {
		r := createFakeReconcileVmi()
		err := r.syncPvc()

		It("Should return no error", func() {
			Expect(err).Should(BeNil())
		})
		It("Should create a pvc", func() {
			pvc := &corev1.PersistentVolumeClaim{}
			err = r.client.Get(context.TODO(), types.NamespacedName{Namespace: r.vmi.Namespace, Name: GetPvcNameFromVmiName(r.vmi.Name)}, pvc)
			Expect(err).Should(BeNil())
		})
		It("Should update state to creating", func() {
			vmi := &hc.VirtualMachineImage{}
			err = r.client.Get(context.TODO(), types.NamespacedName{Namespace: r.vmi.Namespace, Name: r.vmi.Name}, vmi)
			Expect(err).Should(BeNil())
			Expect(vmi.Status.State).Should(Equal(hc.VirtualMachineImageStateCreating))
		})
		It("Should update readyToUse to false", func() {
			vmi := &hc.VirtualMachineImage{}
			err = r.client.Get(context.TODO(), types.NamespacedName{Namespace: r.vmi.Namespace, Name: r.vmi.Name}, vmi)
			Expect(err).Should(BeNil())
			Expect(*vmi.Status.ReadyToUse).Should(BeFalse())
		})
	})

	Context("2. with pvc", func() {
		r := createFakeReconcileVmi()
		err := r.syncPvc()

		It("Should return no error", func() {
			Expect(err).Should(BeNil())
		})
		It("Should not delete the pvc", func() {
			pvc := &corev1.PersistentVolumeClaim{}
			err = r.client.Get(context.TODO(), types.NamespacedName{Namespace: r.vmi.Namespace, Name: GetPvcNameFromVmiName(r.vmi.Name)}, pvc)
			Expect(err).Should(BeNil())
		})
	})
})

// pvc가 없는 경우, pvc가 있는데 애노테이션이 없는 경우, pvc가 있고 애노테이션이 no인 경우, pvc가 있고 애노테이션이 yes인 경우
// 번호		pvc		애노테이션
// 1		X
// 2		O		X
// 3		O		no
// 4		O		yes
var _ = Describe("isPvcImported", func() {
	Context("1. with no pvc", func() {
		r := createFakeReconcileVmi()
		_, found, err := r.isPvcImported()

		It("Should return no error", func() {
			Expect(err).Should(BeNil())
		})
		It("Should return not found", func() {
			Expect(found).Should(BeFalse())
		})
	})

	Context("2. with pvc without annotation", func() {
		pvcWithoutAnnotation := &corev1.PersistentVolumeClaim{
			ObjectMeta: v1.ObjectMeta{
				Name:        GetPvcNameFromVmiName(testVmiName),
				Namespace:   testVmiNs,
				Annotations: nil,
			},
		}
		r := createFakeReconcileVmi(pvcWithoutAnnotation)
		_, _, err := r.isPvcImported()

		It("Should return error", func() {
			Expect(err).ShouldNot(BeNil())
		})
	})

	Context("3. with pvc with annotation(imported=no)", func() {
		notImportedPvc := &corev1.PersistentVolumeClaim{
			ObjectMeta: v1.ObjectMeta{
				Name:      GetPvcNameFromVmiName(testVmiName),
				Namespace: testVmiNs,
				Annotations: map[string]string{
					"imported": "no",
				},
			},
		}
		r := createFakeReconcileVmi(notImportedPvc)
		imported, found, err := r.isPvcImported()

		It("Should return no error", func() {
			Expect(err).Should(BeNil())
		})
		It("Should return found", func() {
			Expect(found).Should(BeTrue())
		})
		It("Should return imported false", func() {
			Expect(imported).Should(BeFalse())
		})
	})

	Context("4. with pvc with annotation(imported=yes)", func() {
		importedPvc := &corev1.PersistentVolumeClaim{
			ObjectMeta: v1.ObjectMeta{
				Name:      GetPvcNameFromVmiName(testVmiName),
				Namespace: testVmiNs,
				Annotations: map[string]string{
					"imported": "yes",
				},
			},
		}
		r := createFakeReconcileVmi(importedPvc)
		imported, found, err := r.isPvcImported()

		It("Should return no error", func() {
			Expect(err).Should(BeNil())
		})
		It("Should return found", func() {
			Expect(found).Should(BeTrue())
		})
		It("Should return imported true", func() {
			Expect(imported).Should(BeTrue())
		})
	})
})

// 번호		pvc		updateTo
// 1		X
// 2		O		yes
// 3		O		no
var _ = Describe("updatePvcImported", func() {
	Context("1. with no pvc", func() {
		r := createFakeReconcileVmi()
		err := r.updatePvcImported(false)

		It("Should return error", func() {
			Expect(errors.IsNotFound(err)).Should(BeTrue())
		})
	})

	Context("2. with pvc and update to yes", func() {
		pvc := &corev1.PersistentVolumeClaim{
			ObjectMeta: v1.ObjectMeta{
				Name:      GetPvcNameFromVmiName(testVmiName),
				Namespace: testVmiNs,
			},
		}
		r := createFakeReconcileVmi(pvc)
		err := r.updatePvcImported(true)

		It("Should not return error", func() {
			Expect(err).Should(BeNil())
		})
		It("Should pvc has annotation with imported yes", func() {
			pvc := &corev1.PersistentVolumeClaim{}
			err = r.client.Get(context.TODO(), types.NamespacedName{Namespace: r.vmi.Namespace, Name: GetPvcNameFromVmiName(r.vmi.Name)}, pvc)
			Expect(err).Should(BeNil())
			Expect(pvc.Annotations["imported"]).Should(Equal("yes"))
		})
	})

	Context("3. with pvc and update to no", func() {
		pvc := &corev1.PersistentVolumeClaim{
			ObjectMeta: v1.ObjectMeta{
				Name:      GetPvcNameFromVmiName(testVmiName),
				Namespace: testVmiNs,
			},
		}
		r := createFakeReconcileVmi(pvc)
		err := r.updatePvcImported(false)

		It("Should not return error", func() {
			Expect(err).Should(BeNil())
		})
		It("Should pvc has annotation with imported no", func() {
			pvc := &corev1.PersistentVolumeClaim{}
			err = r.client.Get(context.TODO(), types.NamespacedName{Namespace: r.vmi.Namespace, Name: GetPvcNameFromVmiName(r.vmi.Name)}, pvc)
			Expect(err).Should(BeNil())
			Expect(pvc.Annotations["imported"]).Should(Equal("no"))
		})
	})
})

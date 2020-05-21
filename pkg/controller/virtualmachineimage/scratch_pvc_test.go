package virtualmachineimage

import (
	"context"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// 번호		pvc		annotation		scratchPvc
// 1		X
// 2		O		X
// 3		O		no				X
// 4		O		no				O
// 5		O		yes				X
// 6		O		yes				O
var _ = Describe("syncScratchPvc", func() {
	Context("1. with no pvc", func() {
		r := createFakeReconcileVmi()
		err := r.syncScratchPvc()

		It("Should return no error", func() {
			Expect(err).Should(BeNil())
		})
	})

	Context("2. with pvc(without annotation)", func() {
		pvcWithoutAnnotation := &corev1.PersistentVolumeClaim{
			ObjectMeta: v1.ObjectMeta{
				Name:      GetPvcNameFromVmiName(testVmiName),
				Namespace: testVmiNs,
			},
		}
		r := createFakeReconcileVmi(pvcWithoutAnnotation)
		err := r.syncScratchPvc()

		It("Should return error", func() {
			Expect(err).ShouldNot(BeNil())
		})
	})

	Context("3. with pvc(imported: no) and no scratchPvc", func() {
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
		err := r.syncScratchPvc()

		It("Should return no error", func() {
			Expect(err).Should(BeNil())
		})
		It("Should create scratch pvc", func() {
			scratchPvc := &corev1.PersistentVolumeClaim{}
			err = r.client.Get(context.TODO(), types.NamespacedName{Namespace: r.vmi.Namespace, Name: getScratchPvcNameFromVmiName(r.vmi.Name)}, scratchPvc)
			Expect(err).Should(BeNil())
		})
	})

	Context("4. with pvc(imported: no) and scratchPvc", func() {
		notImportedPvc := &corev1.PersistentVolumeClaim{
			ObjectMeta: v1.ObjectMeta{
				Name:      GetPvcNameFromVmiName(testVmiName),
				Namespace: testVmiNs,
				Annotations: map[string]string{
					"imported": "no",
				},
			},
		}
		scratchPvc := &corev1.PersistentVolumeClaim{
			ObjectMeta: v1.ObjectMeta{
				Name:      getScratchPvcNameFromVmiName(testVmiName),
				Namespace: testVmiNs,
			},
		}
		r := createFakeReconcileVmi(notImportedPvc, scratchPvc)
		err := r.syncScratchPvc()

		It("Should return no error", func() {
			Expect(err).Should(BeNil())
		})
		It("Should not delete scratch pvc", func() {
			scratchPvc := &corev1.PersistentVolumeClaim{}
			err = r.client.Get(context.TODO(), types.NamespacedName{Namespace: r.vmi.Namespace, Name: getScratchPvcNameFromVmiName(r.vmi.Name)}, scratchPvc)
			Expect(err).Should(BeNil())
		})
	})

	Context("5. with pvc(imported: yes) and no scratchPvc", func() {
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
		err := r.syncScratchPvc()

		It("Should return no error", func() {
			Expect(err).Should(BeNil())
		})
		It("Should not create scratch pvc", func() {
			scratchPvc := &corev1.PersistentVolumeClaim{}
			err = r.client.Get(context.TODO(), types.NamespacedName{Namespace: r.vmi.Namespace, Name: getScratchPvcNameFromVmiName(r.vmi.Name)}, scratchPvc)
			Expect(errors.IsNotFound(err)).Should(BeTrue())
		})
	})

	Context("6. with pvc(imported: yes) and scratchPvc", func() {
		importedPvc := &corev1.PersistentVolumeClaim{
			ObjectMeta: v1.ObjectMeta{
				Name:      GetPvcNameFromVmiName(testVmiName),
				Namespace: testVmiNs,
				Annotations: map[string]string{
					"imported": "yes",
				},
			},
		}
		scratchPvc := &corev1.PersistentVolumeClaim{
			ObjectMeta: v1.ObjectMeta{
				Name:      getScratchPvcNameFromVmiName(testVmiName),
				Namespace: testVmiNs,
			},
		}
		r := createFakeReconcileVmi(importedPvc, scratchPvc)
		err := r.syncScratchPvc()

		It("Should return no error", func() {
			Expect(err).Should(BeNil())
		})
		It("Should delete scratch pvc", func() {
			scratchPvc := &corev1.PersistentVolumeClaim{}
			err = r.client.Get(context.TODO(), types.NamespacedName{Namespace: r.vmi.Namespace, Name: getScratchPvcNameFromVmiName(r.vmi.Name)}, scratchPvc)
			Expect(errors.IsNotFound(err)).Should(BeTrue())
		})
	})
})

package virtualmachineimage

import (
	"context"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// 번호		pvc		imported		importPod		importPodState
// 1		X
// 2		O		yes				X
// 3		O		no				X
// 4		O		no				O				Running
// 5		O		no				O				Complete
var _ = Describe("syncImporterPod", func() {
	Context("1. with no pvc", func() {
		r := createFakeReconcileVmi()
		err := r.syncImporterPod()

		It("Should return no error", func() {
			Expect(err).Should(BeNil())
		})
		It("Should not create importerPod", func() {
			importerPod := &corev1.Pod{}
			err = r.client.Get(context.TODO(), types.NamespacedName{Namespace: r.vmi.Namespace, Name: GetImporterPodNameFromVmiName(r.vmi.Name)}, importerPod)
			Expect(errors.IsNotFound(err)).Should(BeTrue())
		})
	})

	Context("2. with pvc, imported=yes, no importerPod", func() {
		importedPvc := &corev1.PersistentVolumeClaim{
			ObjectMeta: metav1.ObjectMeta{
				Name:      GetPvcNameFromVmiName(testVmiName),
				Namespace: testVmiNs,
				Annotations: map[string]string{
					"imported": "yes",
				},
			},
		}
		r := createFakeReconcileVmi(importedPvc)
		err := r.syncImporterPod()

		It("Should return no error", func() {
			Expect(err).Should(BeNil())
		})
		It("Should not create importerPod", func() {
			importerPod := &corev1.Pod{}
			err = r.client.Get(context.TODO(), types.NamespacedName{Namespace: r.vmi.Namespace, Name: GetImporterPodNameFromVmiName(r.vmi.Name)}, importerPod)
			Expect(errors.IsNotFound(err)).Should(BeTrue())
		})
	})

	Context("3. with pvc, imported=no, no importerPod", func() {
		notImportedPvc := &corev1.PersistentVolumeClaim{
			ObjectMeta: metav1.ObjectMeta{
				Name:      GetPvcNameFromVmiName(testVmiName),
				Namespace: testVmiNs,
				Annotations: map[string]string{
					"imported": "no",
				},
			},
		}
		r := createFakeReconcileVmi(notImportedPvc)
		err := r.syncImporterPod()

		It("Should return no error", func() {
			Expect(err).Should(BeNil())
		})
		It("Should create importerPod", func() {
			importerPod := &corev1.Pod{}
			err = r.client.Get(context.TODO(), types.NamespacedName{Namespace: r.vmi.Namespace, Name: GetImporterPodNameFromVmiName(r.vmi.Name)}, importerPod)
			Expect(err).Should(BeNil())
		})
	})

	Context("4. with pvc, imported=no, importerPod with running", func() {
		notImportedPvc := &corev1.PersistentVolumeClaim{
			ObjectMeta: metav1.ObjectMeta{
				Name:      GetPvcNameFromVmiName(testVmiName),
				Namespace: testVmiNs,
				Annotations: map[string]string{
					"imported": "no",
				},
			},
		}
		importerPod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      GetImporterPodNameFromVmiName(testVmiName),
				Namespace: testVmiNs,
			},
		}
		r := createFakeReconcileVmi(notImportedPvc, importerPod)
		err := r.syncImporterPod()

		It("Should return no error", func() {
			Expect(err).Should(BeNil())
		})
		It("Should not delete importerPod", func() {
			importerPod := &corev1.Pod{}
			err = r.client.Get(context.TODO(), types.NamespacedName{Namespace: r.vmi.Namespace, Name: GetImporterPodNameFromVmiName(r.vmi.Name)}, importerPod)
			Expect(err).Should(BeNil())
		})
	})

	Context("5. with pvc, imported=no, importerPod with complete", func() {
		notImportedPvc := &corev1.PersistentVolumeClaim{
			ObjectMeta: metav1.ObjectMeta{
				Name:      GetPvcNameFromVmiName(testVmiName),
				Namespace: testVmiNs,
				Annotations: map[string]string{
					"imported": "no",
				},
			},
		}
		importerPod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      GetImporterPodNameFromVmiName(testVmiName),
				Namespace: testVmiNs,
			},
			Status: corev1.PodStatus{
				ContainerStatuses: []corev1.ContainerStatus{
					{
						State: corev1.ContainerState{
							Terminated: &corev1.ContainerStateTerminated{
								Reason: "Completed",
							},
						},
					},
				},
			},
		}
		r := createFakeReconcileVmi(notImportedPvc, importerPod)
		err := r.syncImporterPod()

		It("Should return no error", func() {
			Expect(err).Should(BeNil())
		})
		It("Should delete importerPod", func() {
			importerPod := &corev1.Pod{}
			err = r.client.Get(context.TODO(), types.NamespacedName{Namespace: r.vmi.Namespace, Name: GetImporterPodNameFromVmiName(r.vmi.Name)}, importerPod)
			Expect(errors.IsNotFound(err)).Should(BeTrue())
		})
		It("Should update pvc annotation(imported=true)", func() {
			pvc := &corev1.PersistentVolumeClaim{}
			err = r.client.Get(context.TODO(), types.NamespacedName{Namespace: r.vmi.Namespace, Name: GetPvcNameFromVmiName(r.vmi.Name)}, pvc)
			Expect(err).Should(BeNil())
			Expect(pvc.Annotations["imported"]).Should(Equal("yes"))
		})
	})
})

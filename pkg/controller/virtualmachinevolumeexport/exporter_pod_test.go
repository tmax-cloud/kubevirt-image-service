package virtualmachinevolumeexport

import (
	"context"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// no.		pvc		completed		exporterPod		exporterPodState
// 1		X
// 2		O		yes				X
// 3		O		no				X
// 4		O		no				O				Running
// 5		O		no				O				Complete
var _ = Describe("syncExporterPod", func() {
	Context("1. with no pvc", func() {
		vmvPvc := newVmvPvc()

		r := createFakeReconcileVmvExport(vmvPvc)
		err := r.syncExporterPod()

		It("Should return no error", func() {
			Expect(err).Should(BeNil())
		})
		It("Should not create exporterPod", func() {
			exporterPod := &corev1.Pod{}
			err = r.client.Get(context.TODO(), types.NamespacedName{Namespace: r.vmvExport.Namespace, Name: GetExporterPodName(r.vmvExport.Name)}, exporterPod)
			Expect(errors.IsNotFound(err)).Should(BeTrue())
		})
	})

	Context("2. with pvc, completed=yes, no exporterPod", func() {
		completedPvc := &corev1.PersistentVolumeClaim{
			ObjectMeta: metav1.ObjectMeta{
				Name:      GetExportPvcName(vmvExportName),
				Namespace: defaultNamespace,
				Annotations: map[string]string{
					"completed": "yes",
				},
			},
		}
		r := createFakeReconcileVmvExport(completedPvc)
		err := r.syncExporterPod()

		It("Should return no error", func() {
			Expect(err).Should(BeNil())
		})
		It("Should not create exporterPod", func() {
			exporterPod := &corev1.Pod{}
			err = r.client.Get(context.TODO(), types.NamespacedName{Namespace: r.vmvExport.Namespace, Name: GetExporterPodName(r.vmvExport.Name)}, exporterPod)
			Expect(errors.IsNotFound(err)).Should(BeTrue())
		})
	})

	Context("3. with pvc, completed=no, no exporterPod", func() {
		notCompletedPvc := &corev1.PersistentVolumeClaim{
			ObjectMeta: metav1.ObjectMeta{
				Name:      GetExportPvcName(vmvExportName),
				Namespace: defaultNamespace,
				Annotations: map[string]string{
					"completed": "no",
				},
			},
		}
		r := createFakeReconcileVmvExport(notCompletedPvc)
		err := r.syncExporterPod()

		It("Should return no error", func() {
			Expect(err).Should(BeNil())
		})
		It("Should create exporterPod", func() {
			exporterPod := &corev1.Pod{}
			err = r.client.Get(context.TODO(), types.NamespacedName{Namespace: r.vmvExport.Namespace, Name: GetExporterPodName(r.vmvExport.Name)}, exporterPod)
			Expect(err).Should(BeNil())
		})
	})

	Context("4. with pvc, completed=no, exporterPod with running", func() {
		notCompletedPvc := &corev1.PersistentVolumeClaim{
			ObjectMeta: metav1.ObjectMeta{
				Name:      GetExportPvcName(vmvExportName),
				Namespace: defaultNamespace,
				Annotations: map[string]string{
					"completed": "no",
				},
			},
		}
		exporterPod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      GetExporterPodName(vmvExportName),
				Namespace: defaultNamespace,
			},
		}
		r := createFakeReconcileVmvExport(notCompletedPvc, exporterPod)
		err := r.syncExporterPod()

		It("Should return no error", func() {
			Expect(err).Should(BeNil())
		})
		It("Should not delete exporterPod", func() {
			exporterPod := &corev1.Pod{}
			err = r.client.Get(context.TODO(), types.NamespacedName{Namespace: r.vmvExport.Namespace, Name: GetExporterPodName(r.vmvExport.Name)}, exporterPod)
			Expect(err).Should(BeNil())
		})
	})

	Context("5. with pvc, completed=no, exporterPod with complete", func() {
		notCompletedPvc := &corev1.PersistentVolumeClaim{
			ObjectMeta: metav1.ObjectMeta{
				Name:      GetExportPvcName(vmvExportName),
				Namespace: defaultNamespace,
				Annotations: map[string]string{
					"completed": "no",
				},
			},
		}
		exporterPod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      GetExporterPodName(vmvExportName),
				Namespace: defaultNamespace,
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
		r := createFakeReconcileVmvExport(notCompletedPvc, exporterPod)
		err := r.syncExporterPod()

		It("Should return no error", func() {
			Expect(err).Should(BeNil())
		})
		It("Should delete exporterPod", func() {
			exporterPod := &corev1.Pod{}
			err = r.client.Get(context.TODO(), types.NamespacedName{Namespace: r.vmvExport.Namespace, Name: GetExporterPodName(r.vmvExport.Name)}, exporterPod)
			Expect(errors.IsNotFound(err)).Should(BeTrue())
		})
		It("Should update pvc annotation(completed=yes)", func() {
			pvc := &corev1.PersistentVolumeClaim{}
			err = r.client.Get(context.TODO(), types.NamespacedName{Namespace: r.vmvExport.Namespace, Name: GetExportPvcName(r.vmvExport.Name)}, pvc)
			Expect(err).Should(BeNil())
			Expect(pvc.Annotations["completed"]).Should(Equal("yes"))
		})
	})
})

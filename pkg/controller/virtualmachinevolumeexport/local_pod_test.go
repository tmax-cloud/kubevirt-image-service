package virtualmachinevolumeexport

import (
	"context"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	hc "kubevirt-image-service/pkg/apis/hypercloud/v1alpha1"
	"kubevirt-image-service/pkg/util"
)

// no.		pvc		completed		localPod
// 1		X
// 2		O		X
// 3		O		no				X
// 4		O		no				O
// 5		O		yes				X
// 6		O		yes				O
var _ = Describe("syncLocalPod", func() {
	Context("1. with no pvc", func() {
		r := createFakeReconcileVmvExport()
		err := r.syncLocalPod()

		It("Should return no error", func() {
			Expect(err).Should(BeNil())
		})
	})

	Context("2. with pvc(without annotation)", func() {
		pvcWithoutAnnotation := &corev1.PersistentVolumeClaim{
			ObjectMeta: metav1.ObjectMeta{
				Name:      GetExportPvcName(vmvExportName),
				Namespace: defaultNamespace,
			},
		}
		r := createFakeReconcileVmvExport(pvcWithoutAnnotation)
		err := r.syncLocalPod()

		It("Should return error", func() {
			Expect(err).ShouldNot(BeNil())
		})
	})

	Context("3. with pvc(completed: no) and no localPod", func() {
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
		err := r.syncLocalPod()

		It("Should not return error", func() {
			Expect(err).Should(BeNil())
		})
		It("Should not create localPod", func() {
			localPod := &corev1.Pod{}
			err = r.client.Get(context.TODO(), types.NamespacedName{Namespace: r.vmvExport.Namespace, Name: getLocalPodName(r.vmvExport.Name)}, localPod)
			Expect(errors.IsNotFound(err)).Should(BeTrue())
		})
	})

	Context("4. with pvc(completed: no) and localPod", func() {
		notCompletedPvc := &corev1.PersistentVolumeClaim{
			ObjectMeta: metav1.ObjectMeta{
				Name:      GetExportPvcName(vmvExportName),
				Namespace: defaultNamespace,
				Annotations: map[string]string{
					"completed": "no",
				},
			},
		}
		localPod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      getLocalPodName(vmvExportName),
				Namespace: defaultNamespace,
			},
		}
		r := createFakeReconcileVmvExport(notCompletedPvc, localPod)
		err := r.syncLocalPod()

		It("Should not return error", func() {
			Expect(err).Should(BeNil())
		})
		It("Should Delete snapshot", func() {
			localPod := &corev1.Pod{}
			err = r.client.Get(context.TODO(), types.NamespacedName{Namespace: r.vmvExport.Namespace, Name: getLocalPodName(r.vmvExport.Name)}, localPod)
			Expect(errors.IsNotFound(err)).Should(BeTrue())
		})
	})

	Context("5. with pvc(completed: yes) and no localPod", func() {
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
		err := r.syncLocalPod()

		It("Should not return error", func() {
			Expect(err).Should(BeNil())
		})
		It("Should Create localPod", func() {
			localPod := &corev1.Pod{}
			err = r.client.Get(context.TODO(), types.NamespacedName{Namespace: r.vmvExport.Namespace, Name: getLocalPodName(r.vmvExport.Name)}, localPod)
			Expect(errors.IsNotFound(err)).Should(BeFalse())
		})
		It("Should update state to available", func() {
			vmvExport := &hc.VirtualMachineVolumeExport{}
			err = r.client.Get(context.TODO(), types.NamespacedName{Namespace: r.vmvExport.Namespace, Name: r.vmvExport.Name}, vmvExport)
			Expect(err).Should(BeNil())
			Expect(vmvExport.Status.State).Should(Equal(hc.VirtualMachineVolumeExportStateCompleted))
		})
		It("Should update readyToUse to true", func() {
			vmvExport := &hc.VirtualMachineVolumeExport{}
			err = r.client.Get(context.TODO(), types.NamespacedName{Namespace: r.vmvExport.Namespace, Name: r.vmvExport.Name}, vmvExport)
			Expect(err).Should(BeNil())
			found, cond := util.GetConditionByType(vmvExport.Status.Conditions, hc.VirtualMachineVolumeExportConditionReadyToUse)
			Expect(found).Should(BeTrue())
			Expect(cond.Status).Should(Equal(corev1.ConditionTrue))
		})
	})

	Context("6. with pvc(completed: yes) and localPod", func() {
		completedPvc := &corev1.PersistentVolumeClaim{
			ObjectMeta: metav1.ObjectMeta{
				Name:      GetExportPvcName(vmvExportName),
				Namespace: defaultNamespace,
				Annotations: map[string]string{
					"completed": "yes",
				},
			},
		}
		localPod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      getLocalPodName(vmvExportName),
				Namespace: defaultNamespace,
			},
		}
		r := createFakeReconcileVmvExport(completedPvc, localPod)
		err := r.syncLocalPod()

		It("Should not return error", func() {
			Expect(err).Should(BeNil())
		})
		It("Should not Delete localPod", func() {
			localPod := &corev1.Pod{}
			err = r.client.Get(context.TODO(), types.NamespacedName{Namespace: r.vmvExport.Namespace, Name: getLocalPodName(r.vmvExport.Name)}, localPod)
			Expect(err).Should(BeNil())
			Expect(errors.IsNotFound(err)).Should(BeFalse())
		})
	})
})

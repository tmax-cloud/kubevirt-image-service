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
)

// no.	pvc
// 1	X
// 2	O
var _ = Describe("syncExportPvc", func() {
	Context("1. with no pvc", func() {
		vmvPvc := newVmvPvc()

		r := createFakeReconcileVmvExport(vmvPvc)
		err := r.syncExportPvc()

		It("Should return no error", func() {
			Expect(err).Should(BeNil())
		})
		It("Should create a pvc", func() {
			pvc := &corev1.PersistentVolumeClaim{}
			err = r.client.Get(context.TODO(), types.NamespacedName{Namespace: r.vmvExport.Namespace, Name: GetExportPvcName(r.vmvExport.Name)}, pvc)
			Expect(err).Should(BeNil())
		})
		It("Should update state to creating", func() {
			vmvExport := &hc.VirtualMachineVolumeExport{}
			err = r.client.Get(context.TODO(), types.NamespacedName{Namespace: r.vmvExport.Namespace, Name: r.vmvExport.Name}, vmvExport)
			Expect(err).Should(BeNil())
			Expect(vmvExport.Status.State).Should(Equal(hc.VirtualMachineVolumeExportStateCreating))
		})
		It("Should update readyToUse to false", func() {
			vmvExport := &hc.VirtualMachineVolumeExport{}
			err = r.client.Get(context.TODO(), types.NamespacedName{Namespace: r.vmvExport.Namespace, Name: r.vmvExport.Name}, vmvExport)
			Expect(err).Should(BeNil())
			Expect(*vmvExport.Status.ReadyToUse).Should(BeFalse())
		})
	})

	Context("2. with pvc", func() {
		vmvPvc := newVmvPvc()

		r := createFakeReconcileVmvExport(vmvPvc)
		err := r.syncExportPvc()

		It("Should return no error", func() {
			Expect(err).Should(BeNil())
		})
		It("Should not delete the pvc", func() {
			pvc := &corev1.PersistentVolumeClaim{}
			err = r.client.Get(context.TODO(), types.NamespacedName{Namespace: r.vmvExport.Namespace, Name: GetExportPvcName(r.vmvExport.Name)}, pvc)
			Expect(err).Should(BeNil())
		})
	})
})

// no.		pvc		completed
// 1		X
// 2		O		X
// 3		O		no
// 4		O		yes
var _ = Describe("isPvcCompleted", func() {
	Context("1. with no pvc", func() {
		r := createFakeReconcileVmvExport()
		_, found, err := r.isPvcExportCompleted()

		It("Should return no error", func() {
			Expect(err).Should(BeNil())
		})
		It("Should return not found", func() {
			Expect(found).Should(BeFalse())
		})
	})

	Context("2. with pvc without annotation", func() {
		pvcWithoutAnnotation := &corev1.PersistentVolumeClaim{
			ObjectMeta: metav1.ObjectMeta{
				Name:        GetExportPvcName(vmvExportName),
				Namespace:   defaultNamespace,
				Annotations: nil,
			},
		}
		r := createFakeReconcileVmvExport(pvcWithoutAnnotation)
		_, _, err := r.isPvcExportCompleted()

		It("Should return error", func() {
			Expect(err).ShouldNot(BeNil())
		})
	})

	Context("3. with pvc with annotation(completed=no)", func() {
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
		completed, found, err := r.isPvcExportCompleted()

		It("Should return no error", func() {
			Expect(err).Should(BeNil())
		})
		It("Should return found", func() {
			Expect(found).Should(BeTrue())
		})
		It("Should return completed false", func() {
			Expect(completed).Should(BeFalse())
		})
	})

	Context("4. with pvc with annotation(completed=yes)", func() {
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
		completed, found, err := r.isPvcExportCompleted()

		It("Should return no error", func() {
			Expect(err).Should(BeNil())
		})
		It("Should return found", func() {
			Expect(found).Should(BeTrue())
		})
		It("Should return completedPvc true", func() {
			Expect(completed).Should(BeTrue())
		})
	})
})

// no.		pvc		updateCompletedTo
// 1		X
// 2		O		yes
// 3		O		no
var _ = Describe("updatePvcCompleted", func() {
	Context("1. with no pvc", func() {
		r := createFakeReconcileVmvExport()
		err := r.updatePvcCompleted(false)

		It("Should return error", func() {
			Expect(errors.IsNotFound(err)).Should(BeTrue())
		})
	})

	Context("2. with pvc and update to yes", func() {
		pvc := &corev1.PersistentVolumeClaim{
			ObjectMeta: metav1.ObjectMeta{
				Name:      GetExportPvcName(vmvExportName),
				Namespace: defaultNamespace,
			},
		}
		r := createFakeReconcileVmvExport(pvc)
		err := r.updatePvcCompleted(true)

		It("Should not return error", func() {
			Expect(err).Should(BeNil())
		})
		It("Should pvc has annotation with completed yes", func() {
			pvc := &corev1.PersistentVolumeClaim{}
			err = r.client.Get(context.TODO(), types.NamespacedName{Namespace: r.vmvExport.Namespace, Name: GetExportPvcName(r.vmvExport.Name)}, pvc)
			Expect(err).Should(BeNil())
			Expect(pvc.Annotations["completed"]).Should(Equal("yes"))
		})
	})

	Context("3. with pvc and update to no", func() {
		pvc := &corev1.PersistentVolumeClaim{
			ObjectMeta: metav1.ObjectMeta{
				Name:      GetExportPvcName(vmvExportName),
				Namespace: defaultNamespace,
			},
		}
		r := createFakeReconcileVmvExport(pvc)
		err := r.updatePvcCompleted(false)

		It("Should not return error", func() {
			Expect(err).Should(BeNil())
		})
		It("Should pvc has annotation with completed no", func() {
			pvc := &corev1.PersistentVolumeClaim{}
			err = r.client.Get(context.TODO(), types.NamespacedName{Namespace: r.vmvExport.Namespace, Name: GetExportPvcName(r.vmvExport.Name)}, pvc)
			Expect(err).Should(BeNil())
			Expect(pvc.Annotations["completed"]).Should(Equal("no"))
		})
	})
})

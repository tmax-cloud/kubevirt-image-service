package virtualmachineimage

import (
	"context"
	snapshotv1beta1 "github.com/kubernetes-csi/external-snapshotter/v2/pkg/apis/volumesnapshot/v1beta1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	hc "github.com/tmax-cloud/kubevirt-image-service/api/v1alpha1"
	"github.com/tmax-cloud/kubevirt-image-service/controllers/util"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// 번호		pvc		annotation		snapshot
// 1		X
// 2		O		X
// 3		O		no				X
// 4		O		no				O
// 5		O		yes				readyToUse
// 6		O		yes				Not ReadyToUse
// 7		O		yes				error
var _ = Describe("syncSnapshot", func() {
	Context("1. with no pvc", func() {
		r := createFakeReconcileVmi()
		err := r.syncSnapshot()

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
		err := r.syncSnapshot()

		It("Should return error", func() {
			Expect(err).ShouldNot(BeNil())
		})
	})

	Context("3. with pvc(imported: no) and no snapshot", func() {
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
		err := r.syncSnapshot()

		It("Should not return error", func() {
			Expect(err).Should(BeNil())
		})
		It("Should not create snapshot", func() {
			snapshot := &snapshotv1beta1.VolumeSnapshot{}
			err = r.Client.Get(context.TODO(), types.NamespacedName{Namespace: r.vmi.Namespace, Name: GetSnapshotNameFromVmiName(r.vmi.Name)}, snapshot)
			Expect(errors.IsNotFound(err)).Should(BeTrue())
		})
	})

	Context("4. with pvc(imported: no) and snapshot", func() {
		notImportedPvc := &corev1.PersistentVolumeClaim{
			ObjectMeta: v1.ObjectMeta{
				Name:      GetPvcNameFromVmiName(testVmiName),
				Namespace: testVmiNs,
				Annotations: map[string]string{
					"imported": "no",
				},
			},
		}
		snapshot := &snapshotv1beta1.VolumeSnapshot{
			ObjectMeta: v1.ObjectMeta{
				Name:      GetSnapshotNameFromVmiName(testVmiName),
				Namespace: testVmiNs,
			},
		}
		r := createFakeReconcileVmi(notImportedPvc, snapshot)
		err := r.syncSnapshot()

		It("Should not return error", func() {
			Expect(err).Should(BeNil())
		})
		It("Should Delete snapshot", func() {
			snapshot := &snapshotv1beta1.VolumeSnapshot{}
			err = r.Client.Get(context.TODO(), types.NamespacedName{Namespace: r.vmi.Namespace, Name: GetSnapshotNameFromVmiName(r.vmi.Name)}, snapshot)
			Expect(errors.IsNotFound(err)).Should(BeTrue())
		})
	})

	Context("5. with pvc(imported: yes) and snapshot(readyToUse: true)", func() {
		importedPod := &corev1.PersistentVolumeClaim{
			ObjectMeta: v1.ObjectMeta{
				Name:      GetPvcNameFromVmiName(testVmiName),
				Namespace: testVmiNs,
				Annotations: map[string]string{
					"imported": "yes",
				},
			},
		}
		readyToUseTrue := true
		snapshot := &snapshotv1beta1.VolumeSnapshot{
			ObjectMeta: v1.ObjectMeta{
				Name:      GetSnapshotNameFromVmiName(testVmiName),
				Namespace: testVmiNs,
			},
			Status: &snapshotv1beta1.VolumeSnapshotStatus{
				ReadyToUse: &readyToUseTrue,
			},
		}
		r := createFakeReconcileVmi(importedPod, snapshot)
		err := r.syncSnapshot()

		It("Should not return error", func() {
			Expect(err).Should(BeNil())
		})
		It("Should not Delete snapshot", func() {
			snapshot := &snapshotv1beta1.VolumeSnapshot{}
			err = r.Client.Get(context.TODO(), types.NamespacedName{Namespace: r.vmi.Namespace, Name: GetSnapshotNameFromVmiName(r.vmi.Name)}, snapshot)
			Expect(err).Should(BeNil())
		})
		It("Should update state to available", func() {
			vmi := &hc.VirtualMachineImage{}
			err = r.Client.Get(context.TODO(), types.NamespacedName{Namespace: r.vmi.Namespace, Name: r.vmi.Name}, vmi)
			Expect(err).Should(BeNil())
			Expect(vmi.Status.State).Should(Equal(hc.VirtualMachineImageStateAvailable))
		})
		It("Should update readyToUse to true", func() {
			vmi := &hc.VirtualMachineImage{}
			err = r.Client.Get(context.TODO(), types.NamespacedName{Namespace: r.vmi.Namespace, Name: r.vmi.Name}, vmi)
			Expect(err).Should(BeNil())
			found, cond := util.GetConditionByType(vmi.Status.Conditions, hc.ConditionReadyToUse)
			Expect(found).Should(BeTrue())
			Expect(cond.Status).Should(Equal(corev1.ConditionTrue))
		})
	})

	Context("6. with pvc(imported: yes) and snapshot(readyToUse: false)", func() {
		importedPod := &corev1.PersistentVolumeClaim{
			ObjectMeta: v1.ObjectMeta{
				Name:      GetPvcNameFromVmiName(testVmiName),
				Namespace: testVmiNs,
				Annotations: map[string]string{
					"imported": "yes",
				},
			},
		}
		readyToUserFalse := false
		snapshot := &snapshotv1beta1.VolumeSnapshot{
			ObjectMeta: v1.ObjectMeta{
				Name:      GetSnapshotNameFromVmiName(testVmiName),
				Namespace: testVmiNs,
			},
			Status: &snapshotv1beta1.VolumeSnapshotStatus{
				ReadyToUse: &readyToUserFalse,
			},
		}
		r := createFakeReconcileVmi(importedPod, snapshot)
		err := r.syncSnapshot()

		It("Should not return error", func() {
			Expect(err).Should(BeNil())
		})
		It("Should not Delete snapshot", func() {
			snapshot := &snapshotv1beta1.VolumeSnapshot{}
			err = r.Client.Get(context.TODO(), types.NamespacedName{Namespace: r.vmi.Namespace, Name: GetSnapshotNameFromVmiName(r.vmi.Name)}, snapshot)
			Expect(err).Should(BeNil())
		})
	})

	Context("7. with pvc(imported: yes) and snapshot(error)", func() {
		importedPod := &corev1.PersistentVolumeClaim{
			ObjectMeta: v1.ObjectMeta{
				Name:      GetPvcNameFromVmiName(testVmiName),
				Namespace: testVmiNs,
				Annotations: map[string]string{
					"imported": "yes",
				},
			},
		}
		errorMessage := "errors"
		snapshot := &snapshotv1beta1.VolumeSnapshot{
			ObjectMeta: v1.ObjectMeta{
				Name:      GetSnapshotNameFromVmiName(testVmiName),
				Namespace: testVmiNs,
			},
			Status: &snapshotv1beta1.VolumeSnapshotStatus{
				Error: &snapshotv1beta1.VolumeSnapshotError{
					Message: &errorMessage,
				},
			},
		}
		r := createFakeReconcileVmi(importedPod, snapshot)
		err := r.syncSnapshot()

		It("Should return error", func() {
			Expect(err).ShouldNot(BeNil())
		})
	})
})

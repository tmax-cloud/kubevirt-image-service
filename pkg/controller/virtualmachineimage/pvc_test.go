package virtualmachineimage

import (
	"context"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

const blockStorageClassName = "rook-ceph-block"

var _ = Describe("getPvc", func() {
	It("Should get the pvc, if vmi has a pvc", func() {
		r, pvc := createFakeReconcileVmiWithPvc()
		err := r.getPvc(false)
		Expect(err).ToNot(HaveOccurred())

		pvcFound := &corev1.PersistentVolumeClaim{}
		err = r.client.Get(context.Background(), types.NamespacedName{Name: "testvmi-pvc", Namespace: "default"}, pvcFound)
		Expect(errors.IsNotFound(err)).To(BeFalse())

		Expect(pvcFound.Spec).To(Equal(pvc.Spec))
	})

	It("Should not get the pvc, if vmi has no pvc", func() {
		r := createFakeReconcileVmi()
		err := r.getPvc(false)

		Expect(errors.IsNotFound(err)).To(BeTrue())
	})
})

var _ = Describe("createPvc", func() {
	It("Should success to create the pvc", func() {
		volumeMode := corev1.PersistentVolumeBlock
		storageClassName := blockStorageClassName
		expectedPvcSpec := corev1.PersistentVolumeClaimSpec{
			VolumeMode:  &volumeMode,
			AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
			Resources: corev1.ResourceRequirements{
				Requests: map[corev1.ResourceName]resource.Quantity{
					corev1.ResourceStorage: resource.MustParse("3Gi"),
				},
			},
			StorageClassName: &storageClassName,
		}

		r := createFakeReconcileVmi()
		err := r.createPvc(false)
		Expect(err).ToNot(HaveOccurred())

		pvcFound := &corev1.PersistentVolumeClaim{}
		err = r.client.Get(context.Background(), types.NamespacedName{Name: "testvmi-pvc", Namespace: "default"}, pvcFound)
		Expect(errors.IsNotFound(err)).To(BeFalse())

		Expect(pvcFound.Spec).To(Equal(expectedPvcSpec))
	})
})

var _ = Describe("deletePvc", func() {
	It("Should delete the pvc, if vmi has a pvc", func() {
		r, _ := createFakeReconcileVmiWithPvc()

		err := r.deletePvc(false)
		Expect(err).ToNot(HaveOccurred())

		pvcFound := &corev1.PersistentVolumeClaim{}
		err = r.client.Get(context.Background(), types.NamespacedName{Name: "testvmi-pvc", Namespace: "default"}, pvcFound)
		Expect(errors.IsNotFound(err)).To(BeTrue())
	})

	It("Should not delete the pvc, if vmi has no pvc", func() {
		r := createFakeReconcileVmi()
		err := r.deletePvc(false)

		Expect(errors.IsNotFound(err)).To(BeTrue())
	})
})

var _ = Describe("GetPvcName", func() {
	It("Should get the pvcName", func() {
		expectedPvcName := "testvmi-pvc"

		r := createFakeReconcileVmi()
		pvcName := GetPvcName(r.vmi.Name, false)

		Expect(pvcName).To(Equal(expectedPvcName))
	})
})

var _ = Describe("newPvc", func() {
	It("Should success to create the pvc", func() {
		r := createFakeReconcileVmi()
		pvc, err := r.newPvc(false)

		isController := true
		blockOwnerDeletion := true
		volumeMode := corev1.PersistentVolumeBlock
		storageClassName := blockStorageClassName
		expectedPvc := &corev1.PersistentVolumeClaim{
			TypeMeta: v1.TypeMeta{
				Kind:       "PersistentVolumeClaim",
				APIVersion: "v1",
			},
			ObjectMeta: v1.ObjectMeta{
				Name:      "testvmi-pvc",
				Namespace: "default",
				OwnerReferences: []v1.OwnerReference{
					{
						APIVersion:         "hypercloud.tmaxanc.com/v1alpha1",
						Kind:               "VirtualMachineImage",
						Name:               "testvmi",
						UID:                "",
						Controller:         &isController,
						BlockOwnerDeletion: &blockOwnerDeletion,
					},
				},
			},
			Spec: corev1.PersistentVolumeClaimSpec{
				VolumeMode:  &volumeMode,
				AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
				Resources: corev1.ResourceRequirements{
					Requests: map[corev1.ResourceName]resource.Quantity{
						corev1.ResourceStorage: resource.MustParse("3Gi"),
					},
				},
				StorageClassName: &storageClassName,
			},
		}

		Expect(err).ToNot(HaveOccurred())
		Expect(pvc).To(Equal(expectedPvc))
	})
})

func createFakeReconcileVmiWithPvc() (*ReconcileVirtualMachineImage, *corev1.PersistentVolumeClaim) {
	storageClassName := blockStorageClassName
	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: v1.ObjectMeta{
			Name:      "testvmi-pvc",
			Namespace: "default",
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
			Resources: corev1.ResourceRequirements{
				Requests: map[corev1.ResourceName]resource.Quantity{
					corev1.ResourceStorage: resource.MustParse("3Gi"),
				},
			},
			StorageClassName: &storageClassName,
		},
	}
	return createFakeReconcileVmi(pvc), pvc
}

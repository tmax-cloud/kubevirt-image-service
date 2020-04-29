package virtualmachineimage

import (
	"context"
	snapshotv1alpha1 "github.com/kubernetes-csi/external-snapshotter/pkg/apis/volumesnapshot/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	hc "kubevirt-image-service/pkg/apis/hypercloud/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

const snapshotClassName = "csi-rbdplugin-snapclass"

var _ = Describe("getSnapshot", func() {
	It("Should get the snapshot, if vmi has a snapshot", func() {
		r, snapshot := createFakeReconcileVmiWithSnapshot()
		snapshotFound, err := r.getSnapshot()

		Expect(err).ToNot(HaveOccurred())
		Expect(snapshotFound.Spec).To(Equal(snapshot.Spec))
	})

	It("Should not get the snapshot, if vmi has no snapshot", func() {
		r := createFakeReconcileVmi()
		_, err := r.getSnapshot()

		Expect(errors.IsNotFound(err)).To(BeTrue())
	})
})

var _ = Describe("createSnapshot", func() {
	It("Should success to create the snapshot", func() {
		snapshotClassName := snapshotClassName
		snapshotSpecExpected := snapshotv1alpha1.VolumeSnapshotSpec{
			Source: &corev1.TypedLocalObjectReference{
				Kind: "PersistentVolumeClaim",
				Name: "testvmi-pvc",
			},
			SnapshotContentName:     "",
			VolumeSnapshotClassName: &snapshotClassName,
		}

		r := createFakeReconcileVmi()
		snapCreated, err := r.createSnapshot()

		Expect(err).ToNot(HaveOccurred())
		Expect(snapCreated.Spec).To(Equal(snapshotSpecExpected))
	})
})

var _ = Describe("deleteSnapshot", func() {
	It("Should delete the snapshot, if vmi has a snapshot", func() {
		r, _ := createFakeReconcileVmiWithSnapshot()

		err := r.deleteSnapshot()
		Expect(err).ToNot(HaveOccurred())

		snapshotFound := &snapshotv1alpha1.VolumeSnapshot{}
		err = r.client.Get(context.Background(), types.NamespacedName{Name: "testvmi-snapshot", Namespace: "default"}, snapshotFound)
		Expect(errors.IsNotFound(err)).To(BeTrue())
	})

	It("Should not delete the snapshot, if vmi has no snapshot", func() {
		r := createFakeReconcileVmi()
		err := r.deleteSnapshot()

		Expect(errors.IsNotFound(err)).To(BeTrue())
	})
})

var _ = Describe("getSnapshotName", func() {
	It("Should get the snapshotName", func() {
		expectedSnapshotName := "testvmi-snapshot"

		r := createFakeReconcileVmi()
		snapshotName := GetSnapshotName(r.vmi.Name)

		Expect(snapshotName).To(Equal(expectedSnapshotName))
	})
})

var _ = Describe("newSnapshot", func() {
	It("Should get the newSnapshot", func() {
		r := createFakeReconcileVmi()
		snapshot, err := r.newSnapshot()

		isController := true
		blockOwnerDeletion := true
		expectedSnapshot := &snapshotv1alpha1.VolumeSnapshot{
			TypeMeta: v1.TypeMeta{
				Kind:       "VolumeSnapshot",
				APIVersion: "snapshot.storage.k8s.io/v1alpha1",
			},
			ObjectMeta: v1.ObjectMeta{
				Name:      "testvmi-snapshot",
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
			Spec: snapshotv1alpha1.VolumeSnapshotSpec{
				Source: &corev1.TypedLocalObjectReference{
					Kind: "PersistentVolumeClaim",
					Name: "testvmi-pvc",
				},
				VolumeSnapshotClassName: &r.vmi.Spec.SnapshotClassName,
			},
		}

		Expect(err).ToNot(HaveOccurred())
		Expect(snapshot).To(Equal(expectedSnapshot))
	})
})

func createFakeReconcileVmi(objects ...runtime.Object) *ReconcileVirtualMachineImage {
	storageClassName := blockStorageClassName
	vmi := &hc.VirtualMachineImage{
		TypeMeta: v1.TypeMeta{
			Kind:       "VirtualMachineImage",
			APIVersion: "hypercloud.tmaxanc.com/v1alpha1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      "testvmi",
			Namespace: "default",
		},
		Spec: hc.VirtualMachineImageSpec{
			Source: hc.VirtualMachineImageSource{
				HTTP: "https://download.cirros-cloud.net/contrib/0.3.0/cirros-0.3.0-i386-disk.img",
			},
			PVC: corev1.PersistentVolumeClaimSpec{
				AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
				Resources: corev1.ResourceRequirements{
					Requests: map[corev1.ResourceName]resource.Quantity{
						corev1.ResourceStorage: resource.MustParse("3Gi"),
					},
				},
				StorageClassName: &storageClassName,
			},
			SnapshotClassName: snapshotClassName,
		},
	}

	s := scheme.Scheme
	err := hc.SchemeBuilder.AddToScheme(s)
	Expect(err).ToNot(HaveOccurred())
	err = snapshotv1alpha1.AddToScheme(s)
	Expect(err).ToNot(HaveOccurred())

	var objs []runtime.Object
	objs = append(objs, objects...)
	cl := fake.NewFakeClientWithScheme(s, objs...)

	return &ReconcileVirtualMachineImage{
		client: cl,
		scheme: s,
		vmi:    vmi,
		log:    log.WithName("controller-test"),
	}
}

func createFakeReconcileVmiWithSnapshot() (*ReconcileVirtualMachineImage, *snapshotv1alpha1.VolumeSnapshot) {
	snapshotClassName := snapshotClassName
	snapshot := &snapshotv1alpha1.VolumeSnapshot{
		ObjectMeta: v1.ObjectMeta{
			Name:      "testvmi-snapshot",
			Namespace: "default",
		},
		Spec: snapshotv1alpha1.VolumeSnapshotSpec{
			Source: &corev1.TypedLocalObjectReference{
				Kind: "PersistentVolumeClaim",
				Name: "testvmi-pvc",
			},
			SnapshotContentName:     "",
			VolumeSnapshotClassName: &snapshotClassName,
		},
	}
	return createFakeReconcileVmi(snapshot), snapshot
}

package virtualmachinevolume

import (
	"context"
	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	hc "kubevirt-image-service/pkg/apis/hypercloud/v1alpha1"
	img "kubevirt-image-service/pkg/controller/virtualmachineimage"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var _ = Describe("Volume reconcile loop", func() {
	table.DescribeTable("Should reconcile volume status against volume pvc phases", func(expected hc.VirtualMachineVolumeState, pvcPhase corev1.PersistentVolumeClaimPhase) {
		image := newImage()
		image.Status.State = hc.VirtualMachineImageStateAvailable
		pvc := newPvc()
		pvc.Status.Phase = pvcPhase
		r := newReconcileVolume(newVolume(), image, pvc)

		_, _ = r.Reconcile(reconcile.Request{NamespacedName: types.NamespacedName{Name: "myvmv", Namespace: "mynamespace"}})
		vmv := &hc.VirtualMachineVolume{}
		_ = r.client.Get(context.TODO(), types.NamespacedName{Name: "myvmv", Namespace: "mynamespace"}, vmv)
		Expect(vmv.Status.State).To(Equal(expected))
	},
		table.Entry("should be available state when pvc is bound", hc.VirtualMachineVolumeStateAvailable, corev1.ClaimBound),
		table.Entry("should be creating state when pvc is pending", hc.VirtualMachineVolumeStateCreating, corev1.ClaimPending),
		table.Entry("should be error state when pvc is lost", hc.VirtualMachineVolumeStateError, corev1.ClaimLost),
	)
})

var _ = Describe("Get volume", func() {
	It("Should get a virtualMachineVolume with the NamespacedName", func() {
		r := newReconcileVolume(newVolume())

		err := r.client.Get(context.TODO(), types.NamespacedName{Name: "myvmv", Namespace: "mynamespace"}, &hc.VirtualMachineVolume{})
		Expect(err).ToNot(HaveOccurred())
	})
})

func newImage() *hc.VirtualMachineImage {
	return &hc.VirtualMachineImage{
		ObjectMeta: v1.ObjectMeta{
			Name:      "myvmi",
			Namespace: "mynamespace",
		},
		Spec: hc.VirtualMachineImageSpec{
			Source: hc.VirtualMachineImageSource{
				HTTP: "https://kr.tmaxsoft.com/main.do",
			},
			PVC: corev1.PersistentVolumeClaimSpec{
				AccessModes:  []corev1.PersistentVolumeAccessMode{
					corev1.ReadWriteOnce,
				},
				Resources: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceStorage: resource.MustParse("3Gi"),
					},
				},
			},
		},
	}
}

func newPvc() *corev1.PersistentVolumeClaim {
	apiGroup := "snapshot.storage.k8s.io"
	storageClass := "rook-ceph-block"
	volumeMode := corev1.PersistentVolumeBlock

	return &corev1.PersistentVolumeClaim{
		ObjectMeta: v1.ObjectMeta{
			Name:      GetVolumePvcName("myvmv"),
			Namespace: "mynamespace",
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			StorageClassName: &storageClass,
			AccessModes: []corev1.PersistentVolumeAccessMode{
				corev1.ReadWriteOnce,
			},
			VolumeMode: &volumeMode,
			DataSource: &corev1.TypedLocalObjectReference{
				APIGroup: &apiGroup,
				Kind:     "VolumeSnapshot",
				Name:     img.GetSnapshotName("myvmi"),
			},
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse("3Gi"),
				},
			},
		},
	}
}

func newVolume() *hc.VirtualMachineVolume {
	return &hc.VirtualMachineVolume{
		ObjectMeta: v1.ObjectMeta{
			Name:      "myvmv",
			Namespace: "mynamespace",
		},
		Spec: hc.VirtualMachineVolumeSpec{
			VirtualMachineImage: hc.VirtualMachineImageName{
				Name: "myvmi",
			},
			Capacity: corev1.ResourceList{
				corev1.ResourceStorage: resource.MustParse("3Gi"),
			},
		},
	}
}

func newReconcileVolume(objects ...runtime.Object) *ReconcileVirtualMachineVolume {
	v := newVolume()
	s := scheme.Scheme
	err := hc.SchemeBuilder.AddToScheme(s)
	Expect(err).ToNot(HaveOccurred())

	var objs []runtime.Object
	objs = append(objs, objects...)
	cl := fake.NewFakeClientWithScheme(s, objs...)

	return &ReconcileVirtualMachineVolume{
		client: cl,
		scheme: s,
		volume: v,
		log:    log.WithName("unit-test"),
	}
}
package virtualmachineimage

import (
	hc "github.com/tmax-cloud/kubevirt-image-service/api/v1alpha1"
	"github.com/tmax-cloud/kubevirt-image-service/controllers/util"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	testVmiName           = "testvmi"
	testVmiNs             = "default"
	testSnapshotClassName = "testSnapshotClassName"
)

var (
	testStorageClassName = "testStorageClassName"
)

func createFakeReconcileVmi(objects ...runtime.Object) *ReconcileVirtualMachineImage {
	vmi := newTestVmi()
	client, scheme, err := util.CreateFakeClientAndScheme(append(objects, vmi)...)
	if err != nil {
		panic(err)
	}
	return &ReconcileVirtualMachineImage{Client: client, Scheme: scheme, vmi: vmi}
}

func newTestVmi() *hc.VirtualMachineImage {
	return &hc.VirtualMachineImage{
		ObjectMeta: v1.ObjectMeta{
			Name:      testVmiName,
			Namespace: testVmiNs,
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
				StorageClassName: &testStorageClassName,
			},
			SnapshotClassName: testSnapshotClassName,
		},
	}
}

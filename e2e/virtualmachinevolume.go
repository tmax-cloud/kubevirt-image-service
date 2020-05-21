package e2e

import (
	"context"
	goerrors "errors"
	framework "github.com/operator-framework/operator-sdk/pkg/test"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"kubevirt-image-service/pkg/apis"
	hc "kubevirt-image-service/pkg/apis/hypercloud/v1alpha1"
	volume "kubevirt-image-service/pkg/controller/virtualmachinevolume"
	"testing"
)

func virtualMachineVolumeTest(t *testing.T, ctx *framework.Context) {
	ns, err := ctx.GetWatchNamespace()
	if err != nil {
		t.Fatal(err)
	}

	vmv := &hc.VirtualMachineVolume{}
	if err := framework.AddToFrameworkScheme(apis.AddToScheme, vmv); err != nil {
		t.Fatal(err)
	}

	cleanupOptions := newCleanupOptions()
	if err := virtualMachineVolumeStateAvailableTest(t, cleanupOptions, ns); err != nil {
		t.Fatal(err)
	}
}

func virtualMachineVolumeStateAvailableTest(t *testing.T, cleanupOptions *framework.CleanupOptions, ns string) error {
	vmv := newVmv(ns)
	if err := framework.Global.Client.Create(context.Background(), vmv, cleanupOptions); err != nil {
		return err
	}

	if err := waitForVmvState(t, ns, hc.VirtualMachineVolumeStateAvailable); err != nil {
		return err
	}
	t.Logf("Vmv available (%s/%s)\n", ns, "testvmv")

	pvc := &corev1.PersistentVolumeClaim{}
	if err := framework.Global.Client.Get(context.Background(), types.NamespacedName{Namespace: ns, Name: volume.GetVolumePvcName("testvmv")}, pvc); err != nil {
		return err
	}
	if pvc.Status.Phase != corev1.ClaimBound {
		return goerrors.New("Vmv pvc is not available (" + ns + "/" + volume.GetVolumePvcName("testvmv") + ")")
	}
	t.Logf("Vmv pvc available (%s/%s)\n", ns, volume.GetVolumePvcName("testvmv"))

	return nil
}

func waitForVmvState(t *testing.T, namespace string, state hc.VirtualMachineVolumeState) error {
	return wait.Poll(retryInterval, timeout, func() (done bool, err error) {
		vmv := &hc.VirtualMachineVolume{}
		if err = framework.Global.Client.Get(context.Background(), types.NamespacedName{Namespace: namespace, Name: "testvmv"}, vmv); err != nil {
			if errors.IsNotFound(err) {
				t.Logf("Waiting for creating vmv: %s in Namespace: %s \n", "testvmv", namespace)
				return false, nil
			}
			return false, err
		}
		t.Logf("Waiting for %s of %s vmv\n", state, "testvmv")
		return vmv.Status.State == state, nil
	})
}

func newVmv(ns string) *hc.VirtualMachineVolume {
	return &hc.VirtualMachineVolume{
		TypeMeta: v1.TypeMeta{
			Kind:       "VirtualMachineVolume",
			APIVersion: "hypercloud.tmaxanc.com/v1alpha1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      "testvmv",
			Namespace: ns,
		},
		Spec: hc.VirtualMachineVolumeSpec{
			VirtualMachineImage: hc.VirtualMachineImageName{
				Name: "testcr",
			},
			Capacity: corev1.ResourceList{
				corev1.ResourceStorage: resource.MustParse("3Gi"),
			},
		},
	}
}

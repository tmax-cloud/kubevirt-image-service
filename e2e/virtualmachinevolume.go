package e2e

import (
	"context"
	framework "github.com/operator-framework/operator-sdk/pkg/test"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"kubevirt-image-service/pkg/apis"
	"kubevirt-image-service/pkg/apis/hypercloud/v1alpha1"
	volume "kubevirt-image-service/pkg/controller/virtualmachinevolume"
	"testing"
	"time"
)

func virtualMachineVolumeTest(t *testing.T, ctx *framework.Context, cleanupOptions *framework.CleanupOptions, retryInterval, timeout time.Duration) {
	ns, err := ctx.GetWatchNamespace()
	if err != nil {
		t.Fatal(err)
	}
	vmv := &v1alpha1.VirtualMachineVolume{}
	if err := framework.AddToFrameworkScheme(apis.AddToScheme, vmv); err != nil {
		t.Fatal(err)
	}
	vmv = &v1alpha1.VirtualMachineVolume{
		TypeMeta: v1.TypeMeta{
			Kind:       "VirtualMachineVolume",
			APIVersion: "hypercloud.tmaxanc.com/v1alpha1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      "testvmv",
			Namespace: ns,
		},
		Spec: v1alpha1.VirtualMachineVolumeSpec{
			VirtualMachineImage: v1alpha1.VirtualMachineImageName{
				Name: "testcr",
			},
			Capacity: corev1.ResourceList{
				corev1.ResourceStorage: resource.MustParse("3Gi"),
			},
		},
	}

	f := framework.Global
	if err := f.Client.Create(context.Background(), vmv, cleanupOptions); err != nil {
		t.Fatal(err)
	}

	if err := waitForVmv(t, ns, "testvmv", retryInterval, timeout); err != nil {
		t.Fatal(err)
	}
	pvc := &corev1.PersistentVolumeClaim{}
	if err := f.Client.Get(context.Background(), types.NamespacedName{Namespace: ns, Name: volume.GetVolumePvcName("testvmv")}, pvc); err != nil {
		t.Fatal(err)
	}
	if pvc.Status.Phase == corev1.ClaimBound {
		t.Logf("Vmv pvc available (%s/%s)\n", ns, volume.GetVolumePvcName("testvmv"))
	} else {
		t.Fatalf("Vmv pvc is not available (%s/%s)\n", ns, volume.GetVolumePvcName("testvmv"))
	}
}

func waitForVmv(t *testing.T, namespace, name string, retryInterval, timeout time.Duration) error {
	f := framework.Global
	err := wait.Poll(retryInterval, timeout, func() (done bool, err error) {
		vmv := &v1alpha1.VirtualMachineVolume{}
		if err = f.Client.Get(context.Background(), types.NamespacedName{Namespace: namespace, Name: name}, vmv); err != nil {
			if errors.IsNotFound(err) {
				t.Logf("Waiting for creating vmv: %s in Namespace: %s \n", name, namespace)
				return false, nil
			}
			return false, err
		}
		if vmv.Status.State == v1alpha1.VirtualMachineVolumeStateAvailable {
			return true, nil
		}
		t.Logf("Waiting for full availability of %s vmv\n", name)
		return false, nil
	})
	if err != nil {
		return err
	}
	t.Logf("Vmv available (%s/%s)\n", namespace, name)
	return nil
}

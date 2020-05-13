package e2e

import (
	"context"
	"kubevirt-image-service/pkg/apis"
	"kubevirt-image-service/pkg/apis/hypercloud/v1alpha1"
	"testing"

	hc "kubevirt-image-service/pkg/apis/hypercloud/v1alpha1"

	framework "github.com/operator-framework/operator-sdk/pkg/test"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
)

func virtualMachineVolumeExportTest(t *testing.T, ctx *framework.Context) {
	ns, err := ctx.GetWatchNamespace()
	if err != nil {
		t.Fatal(err)
	}
	vmve := &v1alpha1.VirtualMachineVolumeExport{}
	if err = framework.AddToFrameworkScheme(apis.AddToScheme, vmve); err != nil {
		t.Fatal(err)
	}

	if err = virtualMachineVolumeExportStateAvailableTest(t, &cleanupOptions, ns); err != nil {
		t.Log(err)
		t.Fatal(err)
	}
}

func virtualMachineVolumeExportStateAvailableTest(t *testing.T, cleanupOptions *framework.CleanupOptions, ns string) error {
	vmveName := "testvmve"
	vmve := newVmve(ns, vmveName)
	if err := framework.Global.Client.Create(context.Background(), vmve, cleanupOptions); err != nil {
		return err
	}

	if err := waitForVmveState(t, ns, vmveName, hc.VirtualMachineVolumeExportStateExportCompleted); err != nil {
		return err
	}
	t.Logf("Vmve %s available\n", vmveName)

	localPodName := vmveName + "-exporter-local"
	if err := waitForPod(t, ns, localPodName); err != nil {
		t.Log(err)
		t.Fatal(err)
	}
	t.Logf("localPod %s is Running", localPodName)
	return nil
}

func waitForVmveState(t *testing.T, namespace, name string, state hc.VirtualMachineVolumeExportState) error {
	return wait.Poll(retryInterval, timeout, func() (done bool, err error) {
		t.Logf("Waiting for creating vmve: %s in Namespace: %s \n", name, namespace)
		vmve := &hc.VirtualMachineVolumeExport{}
		if err = framework.Global.Client.Get(context.Background(), types.NamespacedName{Namespace: namespace, Name: name}, vmve); err != nil {
			if errors.IsNotFound(err) {
				return false, nil
			}
			return false, err
		}
		t.Logf("Waiting for %s of %s vmve\n", state, name)
		return vmve.Status.State == state, nil
	})
}

func newVmve(ns, vmveName string) *hc.VirtualMachineVolumeExport {
	local := v1alpha1.VirtualMachineVolumeExportDestinationLocal{}
	return &hc.VirtualMachineVolumeExport{
		TypeMeta: v1.TypeMeta{
			Kind:       "VirtualMachineVolumeExport",
			APIVersion: "hypercloud.tmaxanc.com/v1alpha1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      vmveName,
			Namespace: ns,
		},
		Spec: v1alpha1.VirtualMachineVolumeExportSpec{
			VirtualMachineVolume: v1alpha1.VirtualMachineVolumeSource{
				Name: "testvmv",
			},
			Destination: v1alpha1.VirtualMachineVolumeExportDestination{
				Local: &local,
			},
		},
	}
}

func waitForPod(t *testing.T, namespace, name string) error {
	return wait.Poll(retryInterval, timeout, func() (done bool, err error) {
		t.Logf("Waiting for creating pod: %s in Namespace: %s \n", name, namespace)
		localPod := &corev1.Pod{}
		if err = framework.Global.Client.Get(context.Background(), types.NamespacedName{Namespace: namespace, Name: name}, localPod); err != nil {
			if errors.IsNotFound(err) {
				return false, nil
			}
			return false, err
		}
		t.Logf("Waiting for Running of %s pod\n", name)
		return localPod.Status.Phase == corev1.PodRunning, nil
	})
}

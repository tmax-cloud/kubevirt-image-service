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
	"testing"
)

func virtualMachineImageTest(t *testing.T, ctx *framework.Context) {
	ns, err := ctx.GetWatchNamespace()
	if err != nil {
		t.Fatal(err)
	}

	sc := &v1alpha1.VirtualMachineImage{}
	err = framework.AddToFrameworkScheme(apis.AddToScheme, sc)
	if err != nil {
		t.Fatal(err)
	}

	cleanupOptions := newCleanupOptions()
	if err = virtualMachineImageStateAvailableTest(t, cleanupOptions, ns, "testcr"); err != nil {
		t.Log(err)
		t.Fatal(err)
	}

	if err = virtualMachineImageStateSnapshottingErrorTest(t, cleanupOptions, ns, "testcr2"); err != nil {
		t.Log(err)
		t.Fatal(err)
	}
}

func virtualMachineImageStateAvailableTest(t *testing.T, cleanupOptions *framework.CleanupOptions, namespace, name string) error {
	vmi := newVmi(namespace, name)

	if err := framework.Global.Client.Create(context.Background(), vmi, cleanupOptions); err != nil {
		return err
	}

	if err := waitForVmiState(t, namespace, name, v1alpha1.VirtualMachineImageStateAvailable); err != nil {
		return err
	}
	return nil
}

func virtualMachineImageStateSnapshottingErrorTest(t *testing.T, cleanupOptions *framework.CleanupOptions, namespace, name string) error {
	vmi := newVmi(namespace, name)
	vmi.Spec.SnapshotClassName = "abc"

	if err := framework.Global.Client.Create(context.Background(), vmi, cleanupOptions); err != nil {
		return err
	}

	if err := waitForVmiState(t, namespace, name, v1alpha1.VirtualMachineImageStateSnapshottingError); err != nil {
		return err
	}
	return nil
}

func waitForVmiState(t *testing.T, namespace, name string, state v1alpha1.VirtualMachineImageState) error {
	return wait.Poll(retryInterval, timeout, func() (done bool, err error) {
		vmi := &v1alpha1.VirtualMachineImage{}
		if err := framework.Global.Client.Get(context.Background(), types.NamespacedName{Namespace: namespace, Name: name}, vmi); err != nil {
			if errors.IsNotFound(err) {
				t.Logf("Waiting for creating vmi: %s in Namespace: %s \n", name, namespace)
				return false, nil
			}
			return false, err
		}
		t.Logf("Waiting for %s of %s vmi\n", state, name)
		return vmi.Status.State == state, nil
	})
}

func newVmi(ns, name string) *v1alpha1.VirtualMachineImage {
	storageClassName := "rook-ceph-block"
	volumeMode := corev1.PersistentVolumeBlock
	return &v1alpha1.VirtualMachineImage{
		ObjectMeta: v1.ObjectMeta{
			Name:      name,
			Namespace: ns,
		},
		Spec: v1alpha1.VirtualMachineImageSpec{
			Source: v1alpha1.VirtualMachineImageSource{
				HTTP: "https://download.cirros-cloud.net/contrib/0.3.0/cirros-0.3.0-i386-disk.img",
			},
			PVC: corev1.PersistentVolumeClaimSpec{
				VolumeMode:  &volumeMode,
				AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
				Resources: corev1.ResourceRequirements{
					Requests: map[corev1.ResourceName]resource.Quantity{
						corev1.ResourceStorage: resource.MustParse("3Gi"),
					},
				},
				StorageClassName: &storageClassName,
			},
			SnapshotClassName: "csi-rbdplugin-snapclass",
		},
	}
}

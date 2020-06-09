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
	"kubevirt-image-service/pkg/util"
	"testing"
)

const (
	storageClassName  = "rook-ceph-block"
	snapshotClassName = "csi-rbdplugin-snapclass"
)

func virtualMachineImageTest(t *testing.T, ctx *framework.Context) error {
	ns, err := ctx.GetWatchNamespace()
	if err != nil {
		return err
	}
	sc := &v1alpha1.VirtualMachineImage{}
	err = framework.AddToFrameworkScheme(apis.AddToScheme, sc)
	if err != nil {
		return err
	}
	if err := testVmi(t, ns); err != nil {
		return err
	}
	if err := testVmiWithInvalidSnapshotClassName(t, ns); err != nil {
		return err
	}
	if err := testVmiWithPvcRwx(t, ns); err != nil {
		return err
	}
	return nil
}

func testVmi(t *testing.T, namespace string) error {
	vmiName := "availvmi"
	vmi := newVmi(namespace, vmiName)
	if err := framework.Global.Client.Create(context.Background(), vmi, &cleanupOptions); err != nil {
		return err
	}
	return waitForVmi(t, namespace, vmiName)
}

func testVmiWithInvalidSnapshotClassName(t *testing.T, namespace string) error {
	vmiName := "errorvmi"
	vmi := newVmi(namespace, vmiName)
	vmi.Spec.SnapshotClassName = "wrongSnapshotClassname"
	if err := framework.Global.Client.Create(context.Background(), vmi, &cleanupOptions); err != nil {
		return err
	}
	return waitForVmiStateError(t, namespace, vmiName)
}

func testVmiWithPvcRwx(t *testing.T, namespace string) error {
	vmiName := "rwxvmi"
	vmi := newVmi(namespace, vmiName)
	vmi.Spec.PVC.AccessModes = []corev1.PersistentVolumeAccessMode{corev1.ReadWriteMany}
	if err := framework.Global.Client.Create(context.Background(), vmi, &cleanupOptions); err != nil {
		return err
	}
	return waitForVmi(t, namespace, vmiName)
}

func waitForVmi(t *testing.T, namespace, name string) error {
	return wait.Poll(retryInterval, timeout, func() (done bool, err error) {
		t.Logf("Waiting for creating vmi: %s in Namespace: %s \n", name, namespace)
		vmi := &v1alpha1.VirtualMachineImage{}
		if err := framework.Global.Client.Get(context.Background(), types.NamespacedName{Namespace: namespace, Name: name}, vmi); err != nil {
			if errors.IsNotFound(err) {
				return false, nil
			}
			return false, err
		}
		found, cond := util.GetConditionByType(vmi.Status.Conditions, v1alpha1.ConditionReadyToUse)
		if found {
			// TODO: check error condition
			return cond.Status == corev1.ConditionTrue, nil
		}
		return false, nil
	})
}

func waitForVmiStateError(t *testing.T, namespace, name string) error {
	return wait.Poll(retryInterval, timeout, func() (done bool, err error) {
		t.Logf("Waiting for creating vmi: %s in Namespace: %s \n", name, namespace)
		vmi := &v1alpha1.VirtualMachineImage{}
		if err := framework.Global.Client.Get(context.Background(), types.NamespacedName{Namespace: namespace, Name: name}, vmi); err != nil {
			if errors.IsNotFound(err) {
				return false, nil
			}
			return false, err
		}
		return vmi.Status.State == v1alpha1.VirtualMachineImageStateError, nil
	})
}

func newVmi(ns, name string) *v1alpha1.VirtualMachineImage {
	scName := storageClassName
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
				StorageClassName: &scName,
			},
			SnapshotClassName: snapshotClassName,
		},
	}
}

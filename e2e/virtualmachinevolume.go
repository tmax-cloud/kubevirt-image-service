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
	vmv "kubevirt-image-service/pkg/controller/virtualmachinevolume"
	"kubevirt-image-service/pkg/util"
	"testing"
)

func virtualMachineVolumeTest(t *testing.T, ctx *framework.Context) {
	ns, err := ctx.GetWatchNamespace()
	if err != nil {
		t.Fatal(err)
	}
	sc := &hc.VirtualMachineVolume{}
	if err := framework.AddToFrameworkScheme(apis.AddToScheme, sc); err != nil {
		t.Fatal(err)
	}
	if err := testVolume(t, ns); err != nil {
		t.Fatal(err)
	}
}

func testVolume(t *testing.T, ns string) error {
	volumeName := "availvmv"
	volume := newVolume(ns, volumeName)
	if err := framework.Global.Client.Create(context.Background(), volume, &cleanupOptions); err != nil {
		return err
	}
	if err := waitForVolumeStatus(t, ns, volumeName); err != nil {
		return err
	}
	t.Logf("Vmv is ready to use (%s/%s)\n", ns, volumeName)

	pvc := &corev1.PersistentVolumeClaim{}
	if err := framework.Global.Client.Get(context.Background(), types.NamespacedName{Namespace: ns, Name: vmv.GetVolumePvcName(volumeName)}, pvc); err != nil {
		return err
	}
	if pvc.Status.Phase != corev1.ClaimBound {
		return goerrors.New("Vmv pvc is not available (" + ns + "/" + vmv.GetVolumePvcName(volumeName) + ")")
	}
	t.Logf("Vmv pvc is ready to use (%s/%s)\n", ns, vmv.GetVolumePvcName(volumeName))
	return nil
}

func waitForVolumeStatus(t *testing.T, namespace, name string) error {
	return wait.Poll(retryInterval, timeout, func() (done bool, err error) {
		volume := &hc.VirtualMachineVolume{}
		if err = framework.Global.Client.Get(context.Background(), types.NamespacedName{Namespace: namespace, Name: name}, volume); err != nil {
			if errors.IsNotFound(err) {
				t.Logf("Waiting for creating volume: %s in Namespace: %s \n", name, namespace)
				return false, nil
			}
			return false, err
		}
		t.Logf("Waiting for volume %s status\n", name)
		found, cond := util.GetConditionByType(volume.Status.Conditions, hc.ConditionReadyToUse)
		if found {
			return cond.Status == corev1.ConditionTrue, nil
		}
		return false, nil
	})
}

func newVolume(ns, name string) *hc.VirtualMachineVolume {
	return &hc.VirtualMachineVolume{
		TypeMeta: v1.TypeMeta{
			Kind:       "VirtualMachineVolume",
			APIVersion: "hypercloud.tmaxanc.com/v1alpha1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      name,
			Namespace: ns,
		},
		Spec: hc.VirtualMachineVolumeSpec{
			VirtualMachineImage: hc.VirtualMachineImageName{
				Name: "availvmi",
			},
			Capacity: corev1.ResourceList{
				corev1.ResourceStorage: resource.MustParse("3Gi"),
			},
		},
	}
}

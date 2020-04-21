package e2e

import (
	"context"
	framework "github.com/operator-framework/operator-sdk/pkg/test"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kubevirt-image-service/pkg/apis"
	"kubevirt-image-service/pkg/apis/hypercloud/v1alpha1"
	"testing"
)

func virtualMachinImageTest(t *testing.T, ctx *framework.Context, cleanupOptions *framework.CleanupOptions) {
	ns, err := ctx.GetWatchNamespace()
	if err != nil {
		t.Fatal(err)
	}

	sc := &v1alpha1.VirtualMachineImage{}
	err = framework.AddToFrameworkScheme(apis.AddToScheme, sc)
	if err != nil {
		t.Fatal(err)
	}

	vmi := &v1alpha1.VirtualMachineImage{
		TypeMeta: v1.TypeMeta{
			Kind:       "VirtualMachineImage",
			APIVersion: "hypercloud.tmaxanc.com/v1alpha1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      "testcr",
			Namespace: ns,
		},
		Spec: v1alpha1.VirtualMachineImageSpec{
			Source: v1alpha1.VirtualMachineImageSource{
				HTTP: "",
			},
			PVC: corev1.PersistentVolumeClaimSpec{
				AccessModes: []corev1.PersistentVolumeAccessMode{},
				Resources: corev1.ResourceRequirements{
					Limits:   nil,
					Requests: nil,
				},
				StorageClassName: nil,
				VolumeMode:       nil,
			},
		},
		Status: v1alpha1.VirtualMachineImageStatus{
			State:  "",
			Reason: "",
		},
	}

	f := framework.Global
	err = f.Client.Create(context.Background(), vmi, cleanupOptions)
	if err != nil {
		t.Log(err)
		t.Fatal(err)
	}

	// TODO...
}

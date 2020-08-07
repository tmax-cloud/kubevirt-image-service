package e2e

import (
	"context"
	"k8s.io/klog"
	"kubevirt-image-service/pkg/apis"
	"kubevirt-image-service/pkg/apis/hypercloud/v1alpha1"
	"kubevirt-image-service/pkg/util"
	"testing"

	bktv1alpha1 "github.com/kube-object-storage/lib-bucket-provisioner/pkg/apis/objectbucket.io/v1alpha1"
	framework "github.com/operator-framework/operator-sdk/pkg/test"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
)

func virtualMachineVolumeExportTest(t *testing.T, ctx *framework.Context) error {
	ns, err := ctx.GetWatchNamespace()
	if err != nil {
		return err
	}
	vmve := &v1alpha1.VirtualMachineVolumeExport{}
	if err := framework.AddToFrameworkScheme(apis.AddToScheme, vmve); err != nil {
		return err
	}
	if err := testLocalVmve(t, &cleanupOptions, ns); err != nil {
		return err
	}

	bucket := &bktv1alpha1.ObjectBucketClaimList{}
	if err := framework.AddToFrameworkScheme(bktv1alpha1.AddToScheme, bucket); err != nil {
		klog.Errorf("%+v", err)
		return err
	}
	if err := testS3Vmve(t, &cleanupOptions, ns); err != nil {
		return err
	}
	return nil
}

func testLocalVmve(t *testing.T, cleanupOptions *framework.CleanupOptions, ns string) error {
	vmveName := "localvmve"
	destination := v1alpha1.VirtualMachineVolumeExportDestination{
		Local: &v1alpha1.VirtualMachineVolumeExportDestinationLocal{},
	}
	vmve := newVmve(ns, vmveName, destination)
	if err := framework.Global.Client.Create(context.Background(), vmve, cleanupOptions); err != nil {
		return err
	}

	if err := waitForVmveState(t, ns, vmveName); err != nil {
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

func testS3Vmve(t *testing.T, cleanupOptions *framework.CleanupOptions, ns string) error {
	bucketName := "ceph-bucket"
	bucket := newBucket(ns, bucketName)
	if err := framework.Global.Client.Create(context.Background(), bucket, cleanupOptions); err != nil {
		return err
	}

	destination := v1alpha1.VirtualMachineVolumeExportDestination{
		S3: &v1alpha1.VirtualMachineVolumeExportDestinationS3{
			URL:       "http://"+bucketName+".rook-ceph-rgw-my-store.rook-ceph:80/disk.img",
			SecretRef: bucketName,
		},
	}
	vmveName := "s3vmve"
	vmve := newVmve(ns, vmveName, destination)
	if err := framework.Global.Client.Create(context.Background(), vmve, cleanupOptions); err != nil {
		return err
	}

	if err := waitForVmveState(t, ns, vmveName); err != nil {
		return err
	}
	t.Logf("Vmve %s available\n", vmveName)

	return nil
}

func newVmve(ns, vmveName string, destination v1alpha1.VirtualMachineVolumeExportDestination) *v1alpha1.VirtualMachineVolumeExport {
	return &v1alpha1.VirtualMachineVolumeExport{
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
				Name: "availvmv",
			},
			Destination: destination,
		},
	}
}

func newBucket(ns, bucketName string) *bktv1alpha1.ObjectBucketClaim{
	return &bktv1alpha1.ObjectBucketClaim{
		TypeMeta:   v1.TypeMeta{
			Kind: "ObjectBucketClaim",
			APIVersion: "objectbucket.io/v1alpha1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name: bucketName,
			Namespace: ns,
		},
		Spec: bktv1alpha1.ObjectBucketClaimSpec{
			BucketName: bucketName,
			StorageClassName: "rook-ceph-bucket",
		},
	}
}

func waitForVmveState(t *testing.T, namespace, name string) error {
	return wait.Poll(retryInterval, timeout, func() (done bool, err error) {
		t.Logf("Waiting for creating vmve: %s in Namespace: %s \n", name, namespace)
		vmve := &v1alpha1.VirtualMachineVolumeExport{}
		if err = framework.Global.Client.Get(context.Background(), types.NamespacedName{Namespace: namespace, Name: name}, vmve); err != nil {
			if errors.IsNotFound(err) {
				return false, nil
			}
			return false, err
		}
		found, cond := util.GetConditionByType(vmve.Status.Conditions, v1alpha1.VirtualMachineVolumeExportConditionReadyToUse)
		if found {
			// TODO: check error condition
			return cond.Status == corev1.ConditionTrue, nil
		}
		return false, nil
	})
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

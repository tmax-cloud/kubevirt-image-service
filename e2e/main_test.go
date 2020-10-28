package e2e

import (
	"context"
	"fmt"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"path/filepath"
	kbtestutils "sigs.k8s.io/kubebuilder/test/e2e/utils"
	"strings"
	"testing"
)

func TestMain(m *testing.M) {
	test.MainEntry(m)
}

func TestKubevirtImageService(t *testing.T) {
	tc, err := kbtestutils.NewTestContext(operatorName, "GO111MODULE=on")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		ctx.Cleanup()
	}()
	if err := deployResources(t, tc); err != nil {
		t.Fatal(err)
	}
	if err := waitForOperator(t, tc); err != nil {
		t.Fatal(err)
	}
	if err := virtualMachineImageTest(t, tc); err != nil {
		t.Fatal(err)
	}
	virtualMachineVolumeTest(t, ctx)
	if err := virtualMachineVolumeExportTest(t, tc); err != nil {
		t.Fatal(err)
	}
}

func deployResources(t *testing.T, ctx *kbtestutils.TestContext) error {
	t.Log("Deploying cluster resources...")
	err := ctx.InitializeClusterResources(&cleanupOptions)
	if err != nil {
		return err
	}
	t.Log("Initialized cluster resources")
	return nil
}

func waitForOperator(t *testing.T, c kubernetes.Interface, namespace string, replicas int) error {
	t.Log("Waiting for operator...")
	return wait.Poll(retryInterval, timeout, func() (done bool, err error) {
		deployment, err := c.AppsV1().Deployments(namespace).Get(context.TODO(), operatorName, metav1.GetOptions{})
		if err != nil {
			if apierrors.IsNotFound(err) {
				t.Logf("Waiting for availability of Deployment: %s in Namespace: %s \n", operatorName, namespace)
				return false, nil
			}
			return false, err
		}

		if int(deployment.Status.AvailableReplicas) >= replicas {
			return true, nil
		}
		t.Logf("Waiting for full availability of %s deployment (%d/%d)\n", operatorName,
			deployment.Status.AvailableReplicas, replicas)
		return false, nil
	})
}

func cleanUp(t *testing.T, ctx *kbtestutils.TestContext) error {
	t.Log("cleaning up the operator and resources")
	defaultOutput, err := ctx.KustomizeBuild(filepath.Join("config", "default"))
	_, err = ctx.Kubectl.WithInput(string(defaultOutput)).Command("delete", "-f", "-")

	t.Log("deleting Curl Pod created")
	_, _ = ctx.Kubectl.Delete(true, "pod", "curl")

	t.Log("cleaning up permissions")
	_, _ = ctx.Kubectl.Command("delete", "clusterrolebinding",
		fmt.Sprintf("metrics-%s", ctx.TestSuffix))

	t.Log("undeploy project")
	_ = ctx.Make("undeploy")

	t.Log("ensuring that the namespace was deleted")
	_, err := ctx.Kubectl.Command("get", "namespace", ctx.Kubectl.Namespace)
	if strings.Contains(err.Error(), "(NotFound): namespaces") {
		return err
	}
}

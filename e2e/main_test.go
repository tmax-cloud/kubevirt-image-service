package e2e

import (
	"github.com/operator-framework/operator-sdk/pkg/test"
	framework "github.com/operator-framework/operator-sdk/pkg/test"
	"github.com/operator-framework/operator-sdk/pkg/test/e2eutil"
	"testing"
	"time"
)

const (
	retryInterval        = time.Second * 5
	timeout              = time.Second * 300
	cleanupRetryInterval = time.Second * 1
	cleanupTimeout       = time.Second * 60
	operatorName         = "kubevirt-image-service"
)

func TestMain(m *testing.M) {
	test.MainEntry(m)
}

func TestKubevirtImageService(t *testing.T) {
	ctx := framework.NewContext(t)

	deployResources(t, ctx)
	waitForOperator(t, ctx)

	virtualMachinImageTest(t, ctx, &framework.CleanupOptions{TestContext: ctx, Timeout: cleanupTimeout, RetryInterval: cleanupRetryInterval}, retryInterval, timeout)
	virtualMachineVolumeTest(t, ctx, &framework.CleanupOptions{TestContext: ctx, Timeout: cleanupTimeout, RetryInterval: cleanupRetryInterval}, retryInterval, timeout)

	// Cleanup only when all tests are succeed for debugging
	ctx.Cleanup()
}

func deployResources(t *testing.T, ctx *framework.Context) {
	t.Log("Deploying cluster resources...")
	err := ctx.InitializeClusterResources(&framework.CleanupOptions{TestContext: ctx, Timeout: cleanupTimeout, RetryInterval: cleanupRetryInterval})
	if err != nil {
		t.Fatalf("Failed to initialize cluster resources: %v", err)
	}
	t.Log("Initialized cluster resources")
}

func waitForOperator(t *testing.T, ctx *framework.Context) {
	t.Log("Waiting for operator...")
	f := framework.Global
	operatorNamespace, err := ctx.GetOperatorNamespace()
	if err != nil {
		t.Fatal(err)
	}
	if err := e2eutil.WaitForOperatorDeployment(t, f.KubeClient, operatorNamespace, operatorName, 3, retryInterval, timeout); err != nil {
		t.Fatal(err)
	}
}

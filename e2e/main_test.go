package e2e

import (
	"testing"

	"github.com/operator-framework/operator-sdk/pkg/test"
	framework "github.com/operator-framework/operator-sdk/pkg/test"
	"github.com/operator-framework/operator-sdk/pkg/test/e2eutil"
)

func TestMain(m *testing.M) {
	test.MainEntry(m)
}

func TestKubevirtImageService(t *testing.T) {
	ctx := framework.NewContext(t)
	defer func() {
		ctx.Cleanup()
	}()
	if err := deployResources(t, ctx); err != nil {
		t.Fatal(err)
	}
	if err := waitForOperator(t, ctx); err != nil {
		t.Fatal(err)
	}
	if err := virtualMachineImageTest(t, ctx); err != nil {
		t.Fatal(err)
	}
	virtualMachineVolumeTest(t, ctx)
	if err := virtualMachineVolumeExportTest(t, ctx); err != nil {
		t.Fatal(err)
	}
}

func deployResources(t *testing.T, ctx *framework.Context) error {
	t.Log("Deploying cluster resources...")
	err := ctx.InitializeClusterResources(&cleanupOptions)
	if err != nil {
		return err
	}
	t.Log("Initialized cluster resources")
	return nil
}

func waitForOperator(t *testing.T, ctx *framework.Context) error {
	t.Log("Waiting for operator...")
	operatorNamespace, err := ctx.GetOperatorNamespace()
	if err != nil {
		return err
	}
	if err := e2eutil.WaitForOperatorDeployment(t, framework.Global.KubeClient, operatorNamespace, operatorName, 3, retryInterval, timeout); err != nil {
		return err
	}
	return nil
}

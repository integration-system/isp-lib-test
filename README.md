# isp-lib-test

## For development environment
Before start writing integration test on local machine 
set some environment variables
```bash
APP_MODE=dev #test context loaded from config file ./conf/config_test.yml
```
## Simple integration test example
```go
package main

import (
	"github.com/integration-system/isp-lib-test/ctx"
	"github.com/integration-system/isp-lib-test/docker"
	"github.com/integration-system/isp-lib-test/utils/postgres"
	"os"
	"testing"
	"time"
)

type TestConfig struct {
	Base         ctx.BaseTestConfiguration
	Dependencies struct {
		MdmService string
	}
}

func (c *TestConfig) GetBaseConfiguration() ctx.BaseTestConfiguration {
	return c.Base
}

func TestMain(m *testing.M) {
	cfg := TestConfig{}
	test, err := ctx.NewIntegrationTest(m, &cfg, setup)
	if err != nil {
		panic(err)
	}
	test.PrepareAndRun()
}

func setup(testCtx *ctx.TestContext, runTest func() int) int {
	cfg := testCtx.BaseConfiguration()
	cli, err := docker.NewClient()
	if err != nil {
		panic(err)
	}
	defer cli.Close()
	env := docker.NewTestEnvironment(testCtx, cli)
	defer env.Cleanup()

	_, pgCfg := env.RunPGContainer()
	_, err = postgres.Wait(pgCfg, 10*time.Second)
	if err != nil {
		panic(err)
	}
	env.RunConfigServiceContainer()	

    // setup others containers here
	appCtx := env.RunAppContainer(
		cfg.Images.Module,
		appConfig,
		remoteConf,
		docker.WithLogger(os.Stdout),
	)

	time.Sleep(3 * time.Second)

	return runTest()
}

//tests here
```

## Notes
By default integration tests skips if flag `-test.short == true` was set in command line

# isp-lib-test

## For development environment
Before start writing integration test on local machine 
set some environment variables
```bash
APP_MODE=dev #test context loaded from config file ./conf/config_test.yml
DOCKER_HOST_MACHINE=localhost 
```
## Simple integration test example
```go
package main

import (
	"github.com/integration-system/isp-lib-test/ctx"
	"github.com/integration-system/isp-lib-test/docker"
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

	netCtx, err := cli.CreateNetwork(testCtx.GetDockerNetwork())
	defer netCtx.Close()
	if err != nil {
		panic(err)
	}

	pgCfg := testCtx.GetDBConfiguration()
	pgCtx, err := cli.RunPGContainer(
		docker.DefaultPGImage, pgCfg.Database, pgCfg.Password,
		docker.WithName(pgCfg.Address),
		docker.WithNetwork(netCtx),
		docker.PullImage("", ""),
	)
	defer pgCtx.ForceRemoveContainer()
	if err != nil {
		panic(err)
	}

	configServiceAddr := testCtx.GetConfigServiceAddress()
	cfgCtx, err := cli.RunAppContainer(
		cfg.Images.ConfigService, testCtx.GetConfigServiceConfiguration(), nil,
		docker.WithName(configServiceAddr.IP),
		docker.WithNetwork(netCtx),
		docker.WithEnv(map[string]string{"APP_MIGRATION_PATH": "migrations_temp"}),
		docker.PullImage(cfg.Registry.Username, cfg.Registry.Password),
	)
	defer cfgCtx.Close()
	if err != nil {
		panic(err)
	}

    // setup others containers here
    // dont forget to defer call ContainerContext.Close()

	time.Sleep(3 * time.Second)

	return runTest()
}

//tests here
```

## Notes
By default integration tests skips if flag `-test.short == true` was set in command line

package docker

import (
	"github.com/hashicorp/go-multierror"
	"github.com/integration-system/isp-lib-test/ctx"
	"github.com/integration-system/isp-lib/structure"
)

type TestEnvironment struct {
	testCtx         *ctx.TestContext
	cfg             ctx.BaseTestConfiguration
	cli             *ispDockerClient
	network         *NetworkContext
	basicContainers []*ContainerContext
	appContainers   []*ContainerContext
}

func (te *TestEnvironment) Network() *NetworkContext {
	return te.network
}

func (te *TestEnvironment) Cleanup() error {
	var errors *multierror.Error
	for i := len(te.appContainers) - 1; i > 0; i-- {
		container := te.appContainers[i]
		err := container.Close()
		errors = multierror.Append(errors, err)
	}
	for i := len(te.basicContainers) - 1; i > 0; i-- {
		container := te.basicContainers[i]
		err := container.ForceRemoveContainer()
		errors = multierror.Append(errors, err)
	}
	err := te.network.Close()
	errors = multierror.Append(errors, err)
	return errors.ErrorOrNil()
}

func (te *TestEnvironment) RunAppContainer(image string, localConfig interface{}, remoteConfig interface{}, opts ...Option) *ContainerContext {
	defaultOpts := []Option{
		WithNetwork(te.network),
		PullImage(te.cfg.Registry.Username, te.cfg.Registry.Password),
	}
	defaultOpts = append(defaultOpts, opts...)
	appCtx, err := te.cli.RunAppContainer(
		image,
		localConfig,
		remoteConfig,
		defaultOpts...,
	)
	te.appContainers = append(te.appContainers, appCtx)
	if err != nil {
		panic(err)
	}
	return appCtx
}

func (te *TestEnvironment) RunConfigServiceContainer(opts ...Option) (*ContainerContext, structure.AddressConfiguration) {
	configServiceAddr := te.testCtx.GetConfigServiceAddress()
	cfgCtx := te.RunAppContainer(te.cfg.Images.ConfigService,
		te.testCtx.GetConfigServiceConfiguration(),
		nil,
		WithName(configServiceAddr.IP),
	)
	configServiceAddr.IP = cfgCtx.GetIPAddress()
	return cfgCtx, configServiceAddr
}

func (te *TestEnvironment) RunPGContainer(opts ...Option) (*ContainerContext, structure.DBConfiguration) {
	pgCfg := te.testCtx.GetDBConfiguration()
	defaultOpts := []Option{
		WithName(pgCfg.Address),
		WithNetwork(te.network),
		PullImage("", ""),
	}
	defaultOpts = append(defaultOpts, opts...)
	pgCtx, err := te.cli.RunPGContainer(
		DefaultPGImage,
		pgCfg.Database,
		pgCfg.Password,
		defaultOpts...,
	)
	te.basicContainers = append(te.basicContainers, pgCtx)
	if err != nil {
		panic(err)
	}
	pgCfg.Address = pgCtx.GetIPAddress()
	return pgCtx, pgCfg
}

func (te *TestEnvironment) RunRabbitContainer(opts ...Option) (*ContainerContext, structure.RabbitConfig) {
	rabbitCfg := te.testCtx.GetRabbitConfiguration()
	defaultOpts := []Option{
		WithName(rabbitCfg.Address.IP),
		WithNetwork(te.network),
		PullImage("", ""),
	}
	defaultOpts = append(defaultOpts, opts...)
	rabbitCtx, err := te.cli.RunContainer(
		DefaultRabbitImage,
		defaultOpts...,
	)
	te.basicContainers = append(te.basicContainers, rabbitCtx)
	if err != nil {
		panic(err)
	}
	rabbitCfg.Address.IP = rabbitCtx.GetIPAddress()
	return rabbitCtx, rabbitCfg
}

func NewTestEnvironment(testCtx *ctx.TestContext, cli *ispDockerClient) *TestEnvironment {
	netCtx, err := cli.CreateNetwork(testCtx.GetDockerNetwork())
	if err != nil {
		netCtx.Close()
		panic(err)
	}
	env := &TestEnvironment{
		testCtx: testCtx,
		cfg:     testCtx.BaseConfiguration(),
		cli:     cli,
		network: netCtx,
	}
	return env
}

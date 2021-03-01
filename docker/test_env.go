package docker

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/integration-system/isp-event-lib/mq"
	"github.com/integration-system/isp-lib-test/ctx"
	"github.com/integration-system/isp-lib/v2/structure"
)

type TestEnvironment struct {
	testCtx         *ctx.TestContext
	cfg             ctx.BaseTestConfiguration
	cli             *ispDockerClient
	network         *NetworkContext
	basicContainers []*ContainerContext
	appContainers   []*ContainerContext
	cleanupFlag     bool
	mu              *sync.Mutex
}

func (te *TestEnvironment) Network() *NetworkContext {
	return te.network
}

func (te *TestEnvironment) Cleanup() error {
	te.mu.Lock()
	defer te.mu.Unlock()
	if te.cleanupFlag {
		return nil
	}
	te.cleanupFlag = true

	var errors *multierror.Error
	for i := len(te.appContainers) - 1; i >= 0; i-- {
		container := te.appContainers[i]
		err := container.Close()
		errors = multierror.Append(errors, err)
	}
	for i := len(te.basicContainers) - 1; i >= 0; i-- {
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
	opts = append([]Option{
		WithName(configServiceAddr.IP),
		PullImage(te.cfg.Registry.Username, te.cfg.Registry.Password),
	}, opts...)
	cfgCtx := te.RunAppContainer(te.cfg.Images.ConfigService,
		te.testCtx.GetConfigServiceConfiguration(),
		nil,
		opts...,
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

func (te *TestEnvironment) RunRabbitContainer(opts ...Option) (*ContainerContext, mq.Config) {
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

func (te *TestEnvironment) RunElasticContainer(opts ...Option) (*ContainerContext, structure.ElasticConfiguration) {
	elasticConfig := te.testCtx.GetElasticConfiguration()
	elasticContainerName := te.testCtx.GetContainer("elasticsearch")
	defaultOpts := []Option{
		WithName(elasticContainerName),
		WithNetwork(te.network),
		PullImage("", ""),
		WithEnv(map[string]string{"discovery.type": "single-node", "ES_JAVA_OPTS": "-Xms512m -Xmx512m"}),
	}
	defaultOpts = append(defaultOpts, opts...)
	elasticCtx, err := te.cli.RunContainer(
		DefaultElasticImage,
		defaultOpts...,
	)
	te.basicContainers = append(te.basicContainers, elasticCtx)
	if err != nil {
		panic(err)
	}
	elasticConfig.URL = fmt.Sprintf("http://%s:%s", elasticCtx.GetIPAddress(), ctx.ElasticPort)
	return elasticCtx, elasticConfig
}

func (te *TestEnvironment) signalCleanupper() {
	quit := make(chan os.Signal, 2)
	signal.Notify(quit, signals...)
	fmt.Println("Receives signal: ", <-quit)
	timoutCh := time.After(3 * time.Second)
	done := make(chan struct{}, 1)

	go func() {
		err := te.Cleanup()
		if err != nil {
			fmt.Printf("Cleanup() was returned an error: %v", err)
		}
		done <- struct{}{}
	}()

	select {
	case <-timoutCh:
		fmt.Println("exit timeout reached: terminating...")
		os.Exit(-1)
	case sig := <-quit:
		fmt.Printf("duplicated exit signal: %s: terminating...\n", sig)
		os.Exit(-1)
	case <-done:
		fmt.Println("correctly exit by signal")
		os.Exit(0)
	}
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
		mu:      &sync.Mutex{},
	}
	go env.signalCleanupper()
	return env
}

package docker

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/hashicorp/go-multierror"
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
	cleanupFlag     int32
	wg              *sync.WaitGroup
}

func (te *TestEnvironment) Network() *NetworkContext {
	return te.network
}

func (te *TestEnvironment) Cleanup() error {
	if !atomic.CompareAndSwapInt32(&te.cleanupFlag, 0, 1) {
		defer te.wg.Wait()
		return nil
	}
	te.wg.Add(1)
	defer te.wg.Done()

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
	signal.Notify(quit, os.Interrupt, os.Kill, syscall.SIGTERM, syscall.SIGQUIT)
	fmt.Println("Receives signal: ", <-quit)
	timoutCh := time.After(3 * time.Second)
	done := make(chan int, 1)

	go func() {
		err := te.Cleanup()
		if err != nil {
			fmt.Printf("Cleanup() was returned an error: %v", err)
			done <- -1
		} else {
			done <- 0
		}
	}()

	select {
	case <-timoutCh:
		fmt.Println("exit timeout reached: terminating...")
		os.Exit(-1)
	case sig := <-quit:
		fmt.Printf("duplicated exit signal: %s: terminating...\n", sig)
		os.Exit(-1)
	case d := <-done:
		if d == 0 {
			fmt.Println("correctly exit by signal")
			os.Exit(0)
		}
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
		wg:      &sync.WaitGroup{},
	}
	go env.signalCleanupper()
	return env
}

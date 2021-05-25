package ctx

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"path"
	"strconv"
	"testing"
	"time"

	"github.com/integration-system/isp-event-lib/event"
	"github.com/integration-system/isp-event-lib/mq"
	"github.com/integration-system/isp-lib-test/internal"
	"github.com/integration-system/isp-lib/v2/structure"
	"github.com/integration-system/isp-lib/v2/utils"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

const (
	configServiceBaseHost = "isp-config-service"
	configServiceHttpPort = "9001"
	configServiceGrpcPort = "9002"
	configServiceSchema   = "config_service"
	configModuleName      = "config"

	bindAddress = "0.0.0.0"

	pgSqlBaseHost = "isp-pgsql"
	pgSqlPort     = "5432"
	PgSqlDbName   = "isp-test"
	PgSqlPassword = "123321"

	rabbitBaseHost = "isp-rabbit"
	rabbitPort     = "5672"
	rabbitUsername = "guest"
	rabbitPassword = "guest"

	elasticBaseHost = "isp-elastic"
	ElasticPort     = "9200"

	dockerNetwork = "isp-test-network"

	TestConfigEnvPrefix = "ISP_TEST"
)

type ConfigServiceLocalConfiguration struct {
	Database         structure.DBConfiguration
	GrpcOuterAddress structure.AddressConfiguration
	ModuleName       string
	WS               struct {
		Rest structure.AddressConfiguration
		Grpc structure.AddressConfiguration
	}
}

type DefaultLocalConfiguration struct {
	ConfigServiceAddress structure.AddressConfiguration
	GrpcOuterAddress     structure.AddressConfiguration
	GrpcInnerAddress     structure.AddressConfiguration
	ModuleName           string
}

// runner must prepare environment for test, defer all resources closing and finally call runTest
type Runner func(ctx *TestContext, runTest func() int) int

type IntegrationTestRunner struct {
	m      *testing.M
	ctx    *TestContext
	runner Runner
}

var (
	backupCleanupFlag = flag.Bool("cleanup", false,
		"removed docker containers and images that was not removed by previous test launch")
	currentSessionName = strconv.FormatInt(time.Now().UnixNano(), 10)
)

// run test only if test.short is false or not specified
func (r *IntegrationTestRunner) PrepareAndRun() {
	flag.Parse()

	if *backupCleanupFlag {
		if internal.CleanupByBackup == nil {
			fmt.Println("Can't cleanup by backup - docker package not imported")
		} else {
			fmt.Println("Start cleanup by backup")
			err := internal.CleanupByBackup()
			if err != nil {
				fmt.Printf("while docker backup cleanup: %v\n", err)
			}
		}
	}

	if testing.Short() {
		fmt.Println("SKIP integration tests")
		return
	}

	code := r.runner(r.ctx, r.m.Run)
	os.Exit(code)
}

func CurrentSessionName() string {
	return currentSessionName
}

type Testable interface {
	GetBaseConfiguration() BaseTestConfiguration
}

type BaseTestConfiguration struct {
	ModuleName string
	Registry   struct {
		Host     string
		Username string
		Password string
	}
	Images struct {
		ConfigService string
		Module        string
	}
}

func (tc *BaseTestConfiguration) GetBaseConfiguration() BaseTestConfiguration {
	return *tc
}

// produce isolated configurations for tests
type TestContext struct {
	cfg     Testable
	baseCfg BaseTestConfiguration
}

func (ctx *TestContext) Configuration() Testable {
	return ctx.cfg
}

func (ctx *TestContext) BaseConfiguration() BaseTestConfiguration {
	return ctx.baseCfg
}

// produce local configuration for config-service instance
func (ctx *TestContext) GetConfigServiceConfiguration() ConfigServiceLocalConfiguration {
	dbCfg := ctx.GetDBConfiguration()
	dbCfg.Schema = configServiceSchema
	return ConfigServiceLocalConfiguration{
		Database: dbCfg,
		GrpcOuterAddress: structure.AddressConfiguration{
			IP:   fmt.Sprintf("%s-%s", configServiceBaseHost, ctx.buildName()),
			Port: configServiceGrpcPort,
		},
		ModuleName: configModuleName,
		WS: struct {
			Rest structure.AddressConfiguration
			Grpc structure.AddressConfiguration
		}{Rest: structure.AddressConfiguration{
			IP:   bindAddress,
			Port: configServiceHttpPort,
		}, Grpc: structure.AddressConfiguration{
			IP:   bindAddress,
			Port: configServiceGrpcPort,
		}},
	}
}

func (ctx *TestContext) GetDBConfiguration() structure.DBConfiguration {
	return structure.DBConfiguration{
		Address:      fmt.Sprintf("%s-%s", pgSqlBaseHost, ctx.buildName()),
		Port:         pgSqlPort,
		Database:     PgSqlDbName,
		Username:     PgSqlDbName,
		Password:     PgSqlPassword,
		CreateSchema: true,
	}
}

func (ctx *TestContext) GetRabbitConfiguration() mq.Config {
	return mq.Config{
		Address: event.AddressConfiguration{
			IP:   fmt.Sprintf("%s-%s", rabbitBaseHost, ctx.buildName()),
			Port: rabbitPort,
		},
		User:     rabbitUsername,
		Password: rabbitPassword,
	}
}

func (ctx *TestContext) GetElasticConfiguration() structure.ElasticConfiguration {
	f := false
	containerName := fmt.Sprintf("%s-%s", elasticBaseHost, ctx.buildName())
	return structure.ElasticConfiguration{
		URL:   fmt.Sprintf("http://%s:%s", containerName, ElasticPort),
		Sniff: &f,
	}
}

func (ctx *TestContext) GetConfigServiceAddress() structure.AddressConfiguration {
	return structure.AddressConfiguration{
		IP:   fmt.Sprintf("%s-%s", configServiceBaseHost, ctx.buildName()),
		Port: configServiceHttpPort,
	}
}

func (ctx *TestContext) GetDockerNetwork() string {
	return fmt.Sprintf("%s-%s", dockerNetwork, ctx.buildName())
}

func (ctx *TestContext) GetModuleLocalConfig(port, moduleName string) DefaultLocalConfiguration {
	return DefaultLocalConfiguration{
		ConfigServiceAddress: ctx.GetConfigServiceAddress(),
		GrpcOuterAddress:     structure.AddressConfiguration{Port: port, IP: ctx.GetContainer(moduleName)},
		GrpcInnerAddress:     structure.AddressConfiguration{Port: port, IP: bindAddress},
		ModuleName:           moduleName,
	}
}

func (ctx *TestContext) GetImage(imageName string) string {
	return fmt.Sprintf("%s/%s", ctx.baseCfg.Registry.Host, imageName)
}

func (ctx *TestContext) GetContainer(baseContainerName string) string {
	return fmt.Sprintf("isp-test-%s-%s", baseContainerName, ctx.buildName())
}

func (ctx *TestContext) buildName() string {
	return ctx.baseCfg.ModuleName + CurrentSessionName()
}

// NewIntegrationTest creates integration test context, loads test configuration from file.
func NewIntegrationTest(m *testing.M, configPtr Testable, runner Runner) (*IntegrationTestRunner, error) {
	ctx, err := loadCtx(configPtr)
	if err != nil {
		return nil, err
	}
	return &IntegrationTestRunner{
		m:      m,
		ctx:    ctx,
		runner: runner,
	}, nil
}

func loadCtx(configPtr Testable) (*TestContext, error) {
	viper := viper.New()

	viper.SetEnvPrefix(TestConfigEnvPrefix)
	viper.AutomaticEnv()
	bindEnvs(viper, configPtr)

	envConfigName := "config_test"
	ex, _ := os.Executable()
	configPath := path.Dir(ex)
	if utils.DEV {
		configPath = "./conf/"
	} else if utils.EnvConfigPath != "" {
		configPath = utils.EnvConfigPath
	}
	viper.SetConfigName(envConfigName)
	viper.AddConfigPath(configPath)

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	} else if err := viper.Unmarshal(configPtr); err != nil {
		return nil, err
	}

	return &TestContext{
		cfg:     configPtr,
		baseCfg: configPtr.GetBaseConfiguration(),
	}, nil
}

// Workaround because viper does not treat env vars the same as other config.
// See https://github.com/spf13/viper/issues/761.
func bindEnvs(cfg *viper.Viper, rawVal interface{}) {
	for _, k := range allKeys(rawVal) {
		_ = cfg.BindEnv(k)
	}
}

func allKeys(rawVal interface{}) []string {
	b, err := yaml.Marshal(rawVal)
	if err != nil {
		return nil
	}

	v := viper.New()
	v.SetConfigType("yaml")
	if err := v.ReadConfig(bytes.NewReader(b)); err != nil {
		return nil
	}

	return v.AllKeys()
}

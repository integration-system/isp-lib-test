package ctx

import (
	"fmt"
	"github.com/integration-system/isp-lib/database"
	"github.com/integration-system/isp-lib/structure"
	"github.com/integration-system/isp-lib/utils"
	"github.com/spf13/viper"
	"os"
	"path"
	"strconv"
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

	dockerNetwork = "isp-test-network"

	DefaultIspInstanceId = "bf482806-0c3d-4e0d-b9d4-12c037b12d70"

	TestCofigEnvPrefix = "ISP_TEST"
)

type ConfigServiceLocalConfiguration struct {
	Database         database.DBConfiguration
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
	InstanceUuid         string
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
}

func (tc *BaseTestConfiguration) GetBaseConfiguration() BaseTestConfiguration {
	return *tc
}

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

func (ctx *TestContext) GetConfigServiceConfiguration() ConfigServiceLocalConfiguration {
	dbCfg := ctx.GetDBConfiguration()
	dbCfg.Schema = configServiceSchema
	return ConfigServiceLocalConfiguration{
		Database: dbCfg,
		GrpcOuterAddress: structure.AddressConfiguration{
			IP:   fmt.Sprintf("%s-%s", configServiceBaseHost, ctx.baseCfg.ModuleName),
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

func (ctx *TestContext) GetDBConfiguration() database.DBConfiguration {
	return database.DBConfiguration{
		Address:      fmt.Sprintf("%s-%s", pgSqlBaseHost, ctx.baseCfg.ModuleName),
		Port:         pgSqlPort,
		Database:     PgSqlDbName,
		Username:     PgSqlDbName,
		Password:     PgSqlPassword,
		CreateSchema: true,
	}
}

func (ctx *TestContext) GetConfigServiceAddress() structure.AddressConfiguration {
	return structure.AddressConfiguration{
		IP:   fmt.Sprintf("%s-%s", configServiceBaseHost, ctx.baseCfg.ModuleName),
		Port: configServiceHttpPort,
	}
}

func (ctx *TestContext) GetDockerNetwork() string {
	return fmt.Sprintf("%s-%s", dockerNetwork, ctx.baseCfg.ModuleName)
}

func (ctx *TestContext) GetModuleLocalConfig(port int) DefaultLocalConfiguration {
	strPort := strconv.Itoa(port)
	return DefaultLocalConfiguration{
		ConfigServiceAddress: ctx.GetConfigServiceAddress(),
		GrpcOuterAddress:     structure.AddressConfiguration{Port: strPort, IP: ctx.GetContainer(ctx.baseCfg.ModuleName)},
		GrpcInnerAddress:     structure.AddressConfiguration{Port: strPort, IP: bindAddress},
		ModuleName:           ctx.baseCfg.ModuleName,
		InstanceUuid:         DefaultIspInstanceId,
	}
}

func (ctx *TestContext) GetImage(imageName string) string {
	return fmt.Sprintf("%s/%s", ctx.baseCfg.Registry.Host, imageName)
}

func (ctx *TestContext) GetContainer(baseContainerName string) string {
	return fmt.Sprintf("isp-test-%s-%s", baseContainerName, ctx.baseCfg.ModuleName)
}

func LoadContext(configPtr Testable) (*TestContext, error) {
	viper := viper.New()

	viper.SetEnvPrefix(TestCofigEnvPrefix)
	viper.AutomaticEnv()

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

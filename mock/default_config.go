package mock

import (
	"fmt"
	"github.com/integration-system/isp-lib/database"
	"github.com/integration-system/isp-lib/structure"
)

const (
	defaultConfigServiceHost = "isp-config-service"
	defaultConfigServicePort = "9001"

	DefaultIspInstanceId = "bf482806-0c3d-4e0d-b9d4-12c037b12d70"
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

func ConfigServiceConfiguration(hostSuffix string) ConfigServiceLocalConfiguration {
	dbCfg := DbConfiguration(hostSuffix)
	dbCfg.Schema = "config_service"
	return ConfigServiceLocalConfiguration{
		Database: dbCfg,
		GrpcOuterAddress: structure.AddressConfiguration{
			IP:   fmt.Sprintf("%s-%s", defaultConfigServiceHost, hostSuffix),
			Port: "9002",
		},
		ModuleName: "config",
		WS: struct {
			Rest structure.AddressConfiguration
			Grpc structure.AddressConfiguration
		}{Rest: structure.AddressConfiguration{
			IP:   "0.0.0.0",
			Port: defaultConfigServicePort,
		}, Grpc: structure.AddressConfiguration{
			IP:   "0.0.0.0",
			Port: "9002",
		}},
	}
}

func DbConfiguration(hostSuffix string) database.DBConfiguration {
	return database.DBConfiguration{
		Address:      fmt.Sprintf("%s-%s", "isp-pgsql", hostSuffix),
		Port:         "5432",
		Database:     "isp-test",
		Username:     "isp-test",
		Password:     "123321",
		CreateSchema: true,
	}
}

func ConfigServiceAddress(hostSuffix string) structure.AddressConfiguration {
	return structure.AddressConfiguration{
		IP:   fmt.Sprintf("%s-%s", defaultConfigServiceHost, hostSuffix),
		Port: defaultConfigServicePort,
	}
}

func DockerNetwork(hostSuffix string) string {
	return fmt.Sprintf("%s-%s", "isp-test-network", hostSuffix)
}

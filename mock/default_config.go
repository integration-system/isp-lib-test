package mock

import (
	"github.com/integration-system/isp-lib/database"
	"github.com/integration-system/isp-lib/structure"
)

const (
	DefaultConfigServiceHost = "isp-config-service"
	DefaultConfigServicePort = "9001"
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

func DefaultConfigServiceConfiguration() ConfigServiceLocalConfiguration {
	dbCfg := DefaultDbConfiguration()
	dbCfg.Schema = "config-service"
	return ConfigServiceLocalConfiguration{
		Database: dbCfg,
		GrpcOuterAddress: structure.AddressConfiguration{
			IP:   DefaultConfigServiceHost,
			Port: "9002",
		},
		ModuleName: "config",
		WS: struct {
			Rest structure.AddressConfiguration
			Grpc structure.AddressConfiguration
		}{Rest: structure.AddressConfiguration{
			IP:   "0.0.0.0",
			Port: DefaultConfigServicePort,
		}, Grpc: structure.AddressConfiguration{
			IP:   "0.0.0.0",
			Port: "9002",
		}},
	}
}

func DefaultDbConfiguration() database.DBConfiguration {
	return database.DBConfiguration{
		Address:      "isp-pg",
		Port:         "5432",
		Database:     "isp-test",
		Username:     "isp-test",
		Password:     "123321",
		CreateSchema: true,
	}
}

package config

import (
	"time"

	"github.com/integration-system/isp-lib-test/utils"
	"github.com/integration-system/isp-lib/v2/backend"
	"github.com/integration-system/isp-lib/v2/structure"
	"google.golang.org/grpc"
)

const (
	getRoutesCommand           = "config/routing/get_routes"
	deleteCommonConfigsCommand = "config/common_config/delete_config"

	configServiceHttpPort = "9001"
	configServiceGrpcPort = "9002"
)

type configIdRequest struct {
	Id string `json:"id" valid:"required~Required"`
}

func Wait(configAddr structure.AddressConfiguration, timeout time.Duration) (*backend.RxGrpcClient, error) {
	if configAddr.Port == configServiceHttpPort {
		configAddr.Port = configServiceGrpcPort
	}
	client := backend.NewRxGrpcClient(backend.WithDialOptions(grpc.WithInsecure()))
	client.ReceiveAddressList([]structure.AddressConfiguration{configAddr})
	req := new(configIdRequest)
	req.Id = "339lc5u5s4"
	var resp interface{}
	_, err := utils.AwaitConnection(func() (interface{}, error) {
		err := client.Invoke(deleteCommonConfigsCommand, -1, req, resp)
		return nil, err
	}, timeout)
	if err != nil {
		return nil, err
	}
	return client, nil
}

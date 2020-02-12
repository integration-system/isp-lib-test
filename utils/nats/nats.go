package nats

import (
	"time"

	"github.com/integration-system/isp-lib-test/utils"
	"github.com/integration-system/isp-lib/v2/nats"
	"github.com/integration-system/isp-lib/v2/structure"
)

func Wait(natsConfig structure.NatsConfig, timeout time.Duration) (*nats.RxNatsClient, error) {
	client := nats.NewRxNatsClient()
	client.ReceiveConfiguration("test-client", natsConfig)

	_, err := utils.AwaitConnection(func() (interface{}, error) {
		// returns ErrNotConnected
		err := client.Visit(func(c *nats.NatsClient) error {
			return nil
		})
		return nil, err
	}, timeout)
	if err != nil {
		return nil, err
	}
	return client, nil
}

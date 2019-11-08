package nats

import (
	"github.com/integration-system/isp-lib-test/utils"
	"github.com/integration-system/isp-lib/nats"
	"github.com/integration-system/isp-lib/structure"
	"time"
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

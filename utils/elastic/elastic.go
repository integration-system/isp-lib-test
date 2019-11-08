package elastic

import (
	"context"
	"github.com/integration-system/isp-lib-test/utils"
	"github.com/integration-system/isp-lib/structure"
	"github.com/olivere/elastic"
	"github.com/olivere/elastic/config"
	"time"
)

func Wait(cfg structure.ElasticConfiguration, timeout time.Duration) (*elastic.Client, error) {
	elasticConfig := config.Config{}
	if err := cfg.ConvertTo(&elasticConfig); err != nil {
		return nil, err
	}

	client, err := utils.AwaitConnection(func() (interface{}, error) {
		return elastic.DialWithConfig(context.Background(), &elasticConfig)
	}, timeout)
	if err != nil {
		return nil, err
	}
	return client.(*elastic.Client), nil
}

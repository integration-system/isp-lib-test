package rabbit

import (
	"time"

	"github.com/integration-system/isp-lib-test/utils"
	"github.com/integration-system/isp-lib/v2/structure"
	"github.com/streadway/amqp"
	"github.com/stretchr/testify/assert"
)

func Wait(rabbitCfg structure.RabbitConfig, timeout time.Duration) (*amqp.Connection, error) {
	uri := rabbitCfg.GetUri()

	conn, err := utils.AwaitConnection(func() (interface{}, error) {
		c, err := amqp.Dial(uri)
		return c, err
	}, timeout)
	if err != nil {
		return nil, err
	}
	return conn.(*amqp.Connection), nil
}

func DeclareQueue(conn *amqp.Connection, queueName string) error {
	ch, err := conn.Channel()
	if ch != nil {
		defer ch.Close()
	}
	if err != nil {
		return err
	}

	_, err = ch.QueueDeclare(queueName, true, false, false, false, nil)
	return err
}

func MakeRabbitChannel(config structure.RabbitConfig) (*amqp.Channel, error) {
	conn, err := amqp.Dial(config.GetUri())
	if err != nil {
		return nil, err
	}
	return conn.Channel()
}

func PublishMessage(body []byte, c *amqp.Channel, queue string, assert *assert.Assertions) bool {
	err := c.Publish("", queue, false, false, amqp.Publishing{
		DeliveryMode: amqp.Persistent,
		Body:         body,
	})
	if !assert.NoError(err) {
		return false
	}
	return true
}

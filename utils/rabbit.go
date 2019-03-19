package utils

import (
	"github.com/integration-system/isp-lib-test/ctx"
	"github.com/integration-system/isp-lib/structure"
	"github.com/streadway/amqp"
	"github.com/stretchr/testify/assert"
	"time"
)

func WaitRabbit(rabbitCfg structure.RabbitConfig, port string, timeout time.Duration) (*amqp.Connection, error) {
	rabbitCfg.Address.IP = ctx.DockerHostMachine
	rabbitCfg.Address.Port = port
	uri := rabbitCfg.GetUri()

	conn, err := AwaitConnection(func() (interface{}, error) {
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

func MakeRabbitChannel(config structure.RabbitConfig, port string) (*amqp.Channel, error) {
	config.Address.IP = ctx.DockerHostMachine
	config.Address.Port = port
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

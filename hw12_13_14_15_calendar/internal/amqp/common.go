package amqp

import (
	"fmt"
	"github.com/Azimkhan/hw-golang/hw12_13_14_15_calendar/internal/conf"
	"github.com/streadway/amqp"
)

func NewChannel(config *conf.AMQPConfig) (*amqp.Connection, *amqp.Channel, error) {
	connection, err := amqp.Dial(config.URI)
	if err != nil {
		return nil, nil, fmt.Errorf("dial: %s", err)
	}

	channel, err := connection.Channel()
	if err != nil {
		return nil, nil, fmt.Errorf("channel: %s", err)
	}

	if err = channel.ExchangeDeclare(
		config.Exchange,
		"direct",
		true,  // durable
		false, // auto-deleted
		false, // internal
		false, // noWait
		nil,   // arguments
	); err != nil {
		return nil, nil, fmt.Errorf("exchange Declare: %s", err)
	}

	return connection, channel, nil
}

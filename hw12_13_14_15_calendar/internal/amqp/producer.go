package amqp

import (
	"encoding/json"
	"fmt"
	"github.com/Azimkhan/hw-golang/hw12_13_14_15_calendar/internal/conf"
	"github.com/streadway/amqp"
)

type Producer struct {
	connection *amqp.Connection
	channel    *amqp.Channel
	config     *conf.AMQPConfig
}

func NewProducer(config *conf.AMQPConfig) (*Producer, error) {
	connection, channel, err := NewChannel(config)
	if err != nil {
		return nil, err
	}
	return &Producer{
		connection: connection,
		channel:    channel,
		config:     config,
	}, nil
}

func (p *Producer) Publish(object interface{}) error {
	serializer, err := json.Marshal(object)
	if err != nil {
		return fmt.Errorf("failed to create amqp channel: %w", err)
	}
	err = p.channel.Publish(
		p.config.Exchange,
		p.config.RoutingKey,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        serializer,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}
	return nil
}

func (p *Producer) Close() error {
	if err := p.channel.Close(); err != nil {
		return fmt.Errorf("failed to close channel: %s", err)
	}
	if err := p.connection.Close(); err != nil {
		return fmt.Errorf("failed to close connection: %s", err)
	}
	return nil
}

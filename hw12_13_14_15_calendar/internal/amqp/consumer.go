package amqp

import (
	"context"
	"fmt"

	"github.com/Azimkhan/hw-golang/hw12_13_14_15_calendar/internal/conf"
	"github.com/Azimkhan/hw-golang/hw12_13_14_15_calendar/internal/logger"
	"github.com/streadway/amqp"
)

type ConsumeHandler func([]byte) error

type Consumer struct {
	conn    *amqp.Connection
	tag     string
	channel *amqp.Channel
	handler ConsumeHandler
	config  *conf.AMQPConfig
	logger  *logger.Logger
}

func NewConsumer(logger *logger.Logger, config *conf.AMQPConfig, handler ConsumeHandler) (*Consumer, error) {
	c := &Consumer{
		config:  config,
		tag:     "consumer",
		logger:  logger,
		conn:    nil,
		channel: nil,
		handler: handler,
	}

	var err error

	c.conn, c.channel, err = NewChannel(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create amqp channel: %w", err)
	}

	logger.Info(fmt.Sprintf("declared Exchange, declaring Queue %q", config.Queue))
	queue, err := c.channel.QueueDeclare(
		config.Queue, // name of the queue
		true,         // durable
		false,        // delete when unused
		false,        // exclusive
		false,        // noWait
		nil,          // arguments
	)
	if err != nil {
		return nil, fmt.Errorf("queue Declare: %w", err)
	}

	logger.Info(fmt.Sprintf("declared Queue (%q %d messages, %d consumers), binding to Exchange (key %q)",
		queue.Name, queue.Messages, queue.Consumers, config.RoutingKey))

	if err = c.channel.QueueBind(
		queue.Name,        // name of the queue
		config.RoutingKey, // bindingKey
		config.Exchange,   // sourceExchange
		false,             // noWait
		nil,               // arguments
	); err != nil {
		return nil, fmt.Errorf("queue Bind: %w", err)
	}
	return c, nil
}

func (c *Consumer) Consume(ctx context.Context) error {
	deliveries, err := c.channel.Consume(
		c.config.Queue, // queue
		c.tag,          // consumer
		false,          // auto-ack
		false,          // exclusive
		false,          // no-local
		false,          // no-wait
		nil,            // args
	)
	if err != nil {
		return fmt.Errorf("queue Consume: %w", err)
	}
	c.logger.Info("start consuming")

	for {
		select {
		case <-ctx.Done():
			return nil
		case d, ok := <-deliveries:
			if !ok {
				return nil
			}
			err := c.handler(d.Body)
			if err != nil {
				c.logger.Error(fmt.Sprintf("failed to handle message: %s", err))
			}
			if err = d.Ack(false); err != nil {
				c.logger.Error(fmt.Sprintf("failed to acknowledge message: %s", err))
			}
		}
	}
}

func (c *Consumer) Stop() error {
	if err := c.channel.Cancel(c.tag, true); err != nil {
		return fmt.Errorf("failed to cancel consumer: %w", err)
	}
	if err := c.conn.Close(); err != nil {
		return fmt.Errorf("failed to close connection: %w", err)
	}
	return nil
}

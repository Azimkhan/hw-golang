package sender

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Azimkhan/hw-golang/hw12_13_14_15_calendar/internal/amqp"
	"github.com/Azimkhan/hw-golang/hw12_13_14_15_calendar/internal/conf"
	"github.com/Azimkhan/hw-golang/hw12_13_14_15_calendar/internal/logger"
	"github.com/Azimkhan/hw-golang/hw12_13_14_15_calendar/internal/messages"
	"github.com/Azimkhan/hw-golang/hw12_13_14_15_calendar/internal/storage"
	"time"
)

type Logger interface {
	Info(msg string)
	Error(msg string)
}

type App struct {
	storage  storage.Storage
	logger   Logger
	config   *conf.SenderConfig
	consumer *amqp.Consumer
}

func New(config *conf.SenderConfig) *App {
	return &App{
		config: config,
	}
}

func (a *App) Run(ctx context.Context) error {
	// create logger
	logg, err := logger.New(a.config.Logger.Level)
	if err != nil {
		return err
	}
	a.logger = logg

	// create storage
	s, closeFunc, err := storage.NewFromConfig(&a.config.Storage)
	if err != nil {
		logg.Error(fmt.Sprintf("failed to create storage: %s", err))
		return err
	}
	if closeFunc != nil {
		defer func() {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			if err := closeFunc(ctx); err != nil {
				logg.Error(fmt.Sprintf("failed to close storage: %s", err))
			}
		}()
	}
	a.storage = s

	// create consumer
	consumer, err := amqp.NewConsumer(logg, &a.config.AMQP, a.handleNotification)
	if err != nil {
		logg.Error(fmt.Sprintf("failed to create consumer: %s", err))
		return err
	}
	a.consumer = consumer
	return consumer.Consume(ctx)
}

func (a *App) handleNotification(msg []byte) error {
	notification := &messages.Notification{}
	if err := json.Unmarshal(msg, notification); err != nil {
		return fmt.Errorf("failed to unmarshal message: %w", err)
	}
	a.logger.Info(fmt.Sprintf("received notification: %v", notification))
	return nil
}

func (a *App) Stop() error {
	if err := a.consumer.Stop(); err != nil {
		return fmt.Errorf("failed to close consumer: %w", err)
	}
	return nil
}

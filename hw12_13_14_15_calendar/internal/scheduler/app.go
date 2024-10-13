package scheduler

import (
	"context"
	"fmt"
	"time"

	"github.com/Azimkhan/hw-golang/hw12_13_14_15_calendar/internal/amqp"
	"github.com/Azimkhan/hw-golang/hw12_13_14_15_calendar/internal/conf"
	"github.com/Azimkhan/hw-golang/hw12_13_14_15_calendar/internal/logger"
	"github.com/Azimkhan/hw-golang/hw12_13_14_15_calendar/internal/storage"
)

type Logger interface {
	Info(msg string)
	Error(msg string)
}

type App struct {
	storage  storage.Storage
	logger   Logger
	config   *conf.SchedulerConfig
	producer *amqp.Producer
}

func New(config *conf.SchedulerConfig) *App {
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
		return fmt.Errorf("failed to create storage: %w", err)
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

	// create producer
	producer, err := amqp.NewProducer(&a.config.AMQP)
	if err != nil {
		return fmt.Errorf("failed to create producer: %w", err)
	}
	a.producer = producer
	defer func() {
		if err := producer.Close(); err != nil {
			logg.Error(fmt.Sprintf("failed to close producer: %s", err))
		}
	}()

	a.logger.Info("starting scheduler")
	return a.runInternal(ctx)
}

func (a *App) runInternal(ctx context.Context) error {
	ticker := time.NewTicker(time.Duration(a.config.ScanInterval) * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			a.logger.Info("scanning events")
			now := time.Now()

			// filter events that belong to range now <= event.StartTime - event.NotifyDelay < now + ScanInterval
			rangeStart := now
			rangeEnd := now.Add(time.Duration(a.config.ScanInterval) * time.Second)
			n, err := a.scanAndSendEvents(rangeStart, rangeEnd)
			if err != nil {
				a.logger.Error(fmt.Sprintf("failed to scan events: %s", err))
			} else {
				a.logger.Info(fmt.Sprintf("sent %d notifications", n))
			}
		}
	}
}

func (a *App) Stop() error {
	if err := a.producer.Close(); err != nil {
		return fmt.Errorf("failed to close producer: %w", err)
	}
	return nil
}

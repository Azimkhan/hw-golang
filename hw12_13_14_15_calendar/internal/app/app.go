package app

import (
	"context"
	"github.com/Azimkhan/hw12_13_14_15_calendar/internal/storage"
	"github.com/Azimkhan/hw12_13_14_15_calendar/internal/storage/model"
)

type App struct {
	Storage  storage.Storage
	Logger   Logger
	BindAddr string
}

type Logger interface {
	Info(msg string)
	Error(msg string)
}

func New(logger Logger, storage storage.Storage, bindAddr string) *App {
	return &App{
		Storage:  storage,
		Logger:   logger,
		BindAddr: bindAddr,
	}
}

func (a *App) CreateEvent(ctx context.Context, id, title string) error {
	event := &model.Event{ID: id, Title: title}
	return a.Storage.CreateEvent(ctx, event)
}

func (a *App) GetHTTPBindAddr() string {
	return a.BindAddr
}

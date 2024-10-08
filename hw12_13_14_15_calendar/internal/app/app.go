package app

import (
	"context"
	"time"

	"github.com/Azimkhan/hw12_13_14_15_calendar/internal/storage"
)

type App struct {
	storage  Storage
	logger   Logger
	bindAddr string
}

type Logger interface {
	Info(msg string)
	Error(msg string)
}

type Storage interface {
	CreateEvent(ctx context.Context, event *storage.Event) error
	UpdateEvent(ctx context.Context, event *storage.Event) error
	RemoveEvent(ctx context.Context, eventID string) error
	FilterEventsByDay(ctx context.Context, date time.Time) ([]*storage.Event, error)
	FilterEventsByWeek(ctx context.Context, weekStart time.Time) ([]*storage.Event, error)
	FilterEventsByMonth(ctx context.Context, monthStart time.Time) ([]*storage.Event, error)
}

func New(logger Logger, storage Storage, bindAddr string) *App {
	return &App{
		storage:  storage,
		logger:   logger,
		bindAddr: bindAddr,
	}
}

func (a *App) CreateEvent(ctx context.Context, id, title string) error {
	event := &storage.Event{ID: id, Title: title}
	return a.storage.CreateEvent(ctx, event)
}

func (a *App) GetHTTPBindAddr() string {
	return a.bindAddr
}

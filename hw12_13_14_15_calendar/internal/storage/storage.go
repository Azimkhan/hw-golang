package storage

import (
	"context"
	"errors"
	"time"

	"github.com/Azimkhan/hw-golang/hw12_13_14_15_calendar/internal/conf"
	memorystorage "github.com/Azimkhan/hw-golang/hw12_13_14_15_calendar/internal/storage/memory"
	"github.com/Azimkhan/hw-golang/hw12_13_14_15_calendar/internal/storage/model"
	sqlstorage "github.com/Azimkhan/hw-golang/hw12_13_14_15_calendar/internal/storage/sql"
)

var ErrUnknownStorageType = errors.New("unknown storage type")

type Storage interface {
	CreateEvent(ctx context.Context, event *model.Event) error
	UpdateEvent(ctx context.Context, event *model.Event) error
	RemoveEvent(ctx context.Context, eventID string) error
	FilterEventsByDay(ctx context.Context, date time.Time) ([]*model.Event, error)
	FilterEventsByWeek(ctx context.Context, weekStart time.Time) ([]*model.Event, error)
	FilterEventsByMonth(ctx context.Context, monthStart time.Time) ([]*model.Event, error)
	DeleteEventsOlderThan(ctx context.Context, threshold time.Time) (int64, error)
}

func NewFromConfig(conf *conf.StorageConf) (Storage, func(ctx context.Context) error, error) {
	switch conf.Type {
	case "inmemory":
		return memorystorage.New(), nil, nil
	case "sql":
		timeout := context.Background()
		pgStorage := sqlstorage.New(conf.DSN)
		if err := pgStorage.Connect(timeout); err != nil {
			return nil, nil, err
		}
		return pgStorage, pgStorage.Close, nil
	default:
		return nil, nil, ErrUnknownStorageType
	}
}

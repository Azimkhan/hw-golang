package storage

import (
	"errors"
	"time"
)

var ErrEventNotFound = errors.New("event not found")

type Event struct {
	ID          string
	Title       string
	StartTime   time.Time
	EndTime     time.Time
	UserID      string
	NotifyDelta int // in minutes
}

type EventStorage interface {
	Add(event *Event) (*Event, error)
	Update(event *Event) error
	Remove(eventID string) error
	FilterByDay(date time.Time) ([]*Event, error)
	FilterByWeek(weekStart time.Time) ([]*Event, error)
	FilterByMonth(monthStart time.Time) ([]*Event, error)
}

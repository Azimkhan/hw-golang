package memorystorage

import (
	"context"
	"sync"
	"time"

	"github.com/Azimkhan/hw-golang/hw12_13_14_15_calendar/internal/storage/model"
)

type Storage struct {
	events map[string]*model.Event
	mu     sync.RWMutex
}

func New() *Storage {
	return &Storage{
		events: make(map[string]*model.Event),
	}
}

func NewWithEvents(events []*model.Event) *Storage {
	eventsMap := make(map[string]*model.Event)
	for _, event := range events {
		eventsMap[event.ID] = event
	}
	return &Storage{
		events: eventsMap,
	}
}

func (s *Storage) CreateEvent(_ context.Context, event *model.Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if event.ID == "" {
		return model.ErrEmptyID
	}
	if _, ok := s.events[event.ID]; ok {
		return model.ErrAlreadyExists
	}
	s.events[event.ID] = event
	return nil
}

func (s *Storage) UpdateEvent(_ context.Context, event *model.Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if event.ID == "" {
		return model.ErrEmptyID
	}
	if _, ok := s.events[event.ID]; !ok {
		return model.ErrEventNotFound
	}
	s.events[event.ID] = event
	return nil
}

func (s *Storage) RemoveEvent(_ context.Context, eventID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.events[eventID]; !ok {
		return model.ErrEventNotFound
	}
	delete(s.events, eventID)
	return nil
}

func (s *Storage) FilterEventsByDay(_ context.Context, date time.Time) ([]*model.Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var events []*model.Event

	for _, event := range s.events {
		if event.StartTime.Day() == date.Day() {
			events = append(events, event)
		}
	}
	return events, nil
}

func (s *Storage) FilterEventsByWeek(_ context.Context, weekStart time.Time) ([]*model.Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var events []*model.Event

	for _, event := range s.events {
		year0, w0 := event.StartTime.ISOWeek()
		year1, w1 := weekStart.ISOWeek()
		if year0 == year1 && w0 == w1 {
			events = append(events, event)
		}
	}
	return events, nil
}

func (s *Storage) FilterEventsByMonth(_ context.Context, monthStart time.Time) ([]*model.Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var events []*model.Event

	for _, event := range s.events {
		if event.StartTime.Month() == monthStart.Month() &&
			event.StartTime.Year() == monthStart.Year() {
			events = append(events, event)
		}
	}
	return events, nil
}

func (s *Storage) DeleteEventsOlderThan(_ context.Context, threshold time.Time) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var n int64
	for id, event := range s.events {
		if event.StartTime.Before(threshold) {
			delete(s.events, id)
			n++
		}
	}
	return n, nil
}

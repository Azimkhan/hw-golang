package memorystorage

import (
	"strconv"
	"sync"
	"time"
)
import "github.com/Azimkhan/hw12_13_14_15_calendar/internal/storage"

type Storage struct {
	events map[string]*storage.Event
	mu     sync.RWMutex
	maxID  int
}

func New() *Storage {
	return &Storage{
		events: make(map[string]*storage.Event),
	}
}

func NewWithEvents(events []*storage.Event) *Storage {
	eventsMap := make(map[string]*storage.Event)
	for _, event := range events {
		eventsMap[event.ID] = event
	}
	return &Storage{
		events: eventsMap,
	}
}

func (s *Storage) Add(event *storage.Event) (*storage.Event, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.maxID++
	event.ID = strconv.Itoa(s.maxID)
	s.events[event.ID] = event
	return event, nil
}

func (s *Storage) Update(event *storage.Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.events[event.ID]; !ok {
		return storage.ErrEventNotFound
	}
	s.events[event.ID] = event
	return nil
}

func (s *Storage) Remove(eventID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.events[eventID]; !ok {
		return storage.ErrEventNotFound
	}
	delete(s.events, eventID)
	return nil
}

func (s *Storage) FilterByDay(date time.Time) ([]*storage.Event, error) {
	// iterrate over all events and filter by day
	// go:
	s.mu.RLock()
	defer s.mu.RUnlock()

	var events []*storage.Event

	for _, event := range s.events {
		if event.StartTime.Day() == date.Day() {
			events = append(events, event)
		}
	}
	return events, nil
}

func (s *Storage) FilterByWeek(weekStart time.Time) ([]*storage.Event, error) {
	// iterrate over all events and filter by week
	// go:
	s.mu.RLock()
	defer s.mu.RUnlock()

	var events []*storage.Event

	for _, event := range s.events {
		year0, w0 := event.StartTime.ISOWeek()
		year1, w1 := weekStart.ISOWeek()
		if year0 == year1 && w0 == w1 {
			events = append(events, event)
		}
	}
	return events, nil
}

func (s *Storage) FilterByMonth(monthStart time.Time) ([]*storage.Event, error) {
	// iterrate over all events and filter by month
	// go:
	s.mu.RLock()
	defer s.mu.RUnlock()

	var events []*storage.Event

	for _, event := range s.events {
		if event.StartTime.Month() == monthStart.Month() && event.StartTime.Year() == monthStart.Year() {
			events = append(events, event)
		}
	}
	return events, nil
}

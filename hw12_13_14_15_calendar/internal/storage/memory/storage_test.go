package memorystorage

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Azimkhan/hw12_13_14_15_calendar/internal/storage"
)

func TestCreate(t *testing.T) {
	s := New()
	event := storage.Event{
		ID:    "1",
		Title: "test",
	}
	err := s.CreateEvent(context.TODO(), &event)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUpdate(t *testing.T) {
	testData := []struct {
		name  string
		event storage.Event
		err   error
	}{
		{
			name: "event not found",
			event: storage.Event{
				ID:    "2",
				Title: "test",
			},
			err: storage.ErrEventNotFound,
		},
		{
			name: "event found",
			event: storage.Event{
				ID:    "1",
				Title: "test 2",
			},
			err: nil,
		},
	}

	for _, tt := range testData {
		t.Run(tt.name, func(t *testing.T) {
			s := NewWithEvents([]*storage.Event{{
				ID:    "1",
				Title: "test",
			}})
			err := s.UpdateEvent(context.TODO(), &tt.event)
			if !errors.Is(err, tt.err) {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestRemove(t *testing.T) {
	s := NewWithEvents([]*storage.Event{{
		ID:    "1",
		Title: "test",
	}})

	err := s.RemoveEvent(context.TODO(), "1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = s.RemoveEvent(context.TODO(), "1")
	if !errors.Is(err, storage.ErrEventNotFound) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestFilterByDay(t *testing.T) {
	s := NewWithEvents([]*storage.Event{
		{
			ID:        "1",
			Title:     "test",
			StartTime: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			ID:        "2",
			Title:     "test 2",
			StartTime: time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
		},
		{
			ID:        "3",
			Title:     "test 3",
			StartTime: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		},
	})

	events, err := s.FilterEventsByDay(context.TODO(), time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("unexpected events count: %v", len(events))
	}
}

func TestFilterByWeek(t *testing.T) {
	s := NewWithEvents([]*storage.Event{
		// week 1
		{
			ID:        "1",
			Title:     "test",
			StartTime: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		},

		// week 1
		{
			ID:        "2",
			Title:     "test 2",
			StartTime: time.Date(2024, 1, 7, 0, 0, 0, 0, time.UTC),
		},
		// week 2
		{
			ID:        "3",
			Title:     "test 3",
			StartTime: time.Date(2024, 1, 8, 0, 0, 0, 0, time.UTC),
		},
	})

	events, err := s.FilterEventsByWeek(context.TODO(), time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("unexpected events count: %v", len(events))
	}
}

func TestFilterByMonth(t *testing.T) {
	s := NewWithEvents([]*storage.Event{
		// month 1
		{
			ID:        "1",
			Title:     "test",
			StartTime: time.Date(2024, 10, 1, 0, 0, 0, 0, time.UTC),
		},

		// month 1
		{
			ID:        "2",
			Title:     "test 2",
			StartTime: time.Date(2024, 10, 7, 0, 0, 0, 0, time.UTC),
		},
		// month 2
		{
			ID:        "3",
			Title:     "test 3",
			StartTime: time.Date(2024, 11, 8, 0, 0, 0, 0, time.UTC),
		},
	})

	events, err := s.FilterEventsByMonth(context.TODO(), time.Date(2024, 10, 1, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("unexpected events count: %v", len(events))
	}
}

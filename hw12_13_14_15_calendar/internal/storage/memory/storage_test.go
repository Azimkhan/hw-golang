package memorystorage

import (
	"errors"
	"github.com/Azimkhan/hw12_13_14_15_calendar/internal/storage"
	"testing"
	"time"
)

func TestStorage(t *testing.T) {
	t.Run("Add", func(t *testing.T) {
		s := New()
		event := storage.Event{
			Title: "test",
		}
		res, err := s.Add(&event)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res.ID != "1" {
			t.Fatalf("unexpected event id: %v", event.ID)
		}

	})

	t.Run("Update", func(t *testing.T) {
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
				err := s.Update(&tt.event)
				if !errors.Is(err, tt.err) {
					t.Fatalf("unexpected error: %v", err)
				}
			})
		}
	})

	t.Run("Remove", func(t *testing.T) {
		s := NewWithEvents([]*storage.Event{{
			ID:    "1",
			Title: "test",
		}})

		err := s.Remove("1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		err = s.Remove("1")
		if !errors.Is(err, storage.ErrEventNotFound) {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("FilterByDay", func(t *testing.T) {
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

		events, err := s.FilterByDay(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)

		}
		if len(events) != 2 {
			t.Fatalf("unexpected events count: %v", len(events))
		}
	})

	t.Run("FilterByWeek", func(t *testing.T) {
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

		events, err := s.FilterByWeek(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)

		}
		if len(events) != 2 {
			t.Fatalf("unexpected events count: %v", len(events))
		}
	})

	t.Run("FilterByMonth", func(t *testing.T) {
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

		events, err := s.FilterByMonth(time.Date(2024, 10, 1, 0, 0, 0, 0, time.UTC))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)

		}
		if len(events) != 2 {
			t.Fatalf("unexpected events count: %v", len(events))
		}
	})

}

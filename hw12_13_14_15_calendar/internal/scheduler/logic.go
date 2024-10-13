package scheduler

import (
	"context"
	"fmt"
	"github.com/Azimkhan/hw-golang/hw12_13_14_15_calendar/internal/messages"
	"github.com/Azimkhan/hw-golang/hw12_13_14_15_calendar/internal/storage/model"
	"time"
)

func (a *App) sendNotification(event *model.Event) error {
	notification := messages.Notification{
		Title:     event.Title,
		UserID:    event.UserID,
		StartTime: event.StartTime,
	}
	if err := a.producer.Publish(&notification); err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}
	return nil
}

func (a *App) scanAndSendEvents(rangeStart, rangeEnd time.Time) (int, error) {
	ctx, cancel := context.WithTimeout(
		context.Background(),
		time.Duration(a.config.ScanInterval)*time.Second-100*time.Millisecond,
	)
	defer cancel()

	// we'll use FilterEventsByDay to get all events for the day
	// and then filter them by time range
	// just for the sake of simplicity
	n := 0
	events, err := a.storage.FilterEventsByDay(ctx, rangeStart)
	if err != nil {
		return n, fmt.Errorf("failed to get events: %w", err)
	}

	for _, event := range events {
		adjustedTime := event.StartTime.Add(-time.Duration(event.NotifyDelta) * time.Second)
		if !(adjustedTime.After(rangeStart) && adjustedTime.Before(rangeEnd)) {
			continue
		}
		if err := a.sendNotification(event); err != nil {
			return n, fmt.Errorf("failed to send for event %s: %w", event.ID, err)
		}
		n++
	}

	return n, nil
}

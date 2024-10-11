package sqlstorage

import (
	"context"
	"testing"
	"time"

	"github.com/Azimkhan/hw-golang/hw12_13_14_15_calendar/internal/storage/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
)

// FilterEventsByDateFixture returns a slice of three events two of which are on the same day,
// and one is on the next day. The second return value is the date of the first event.
func FilterEventsByDateFixture() ([]*model.Event, time.Time) {
	date := time.Date(2024, 10, 1, 13, 0, 0, 0, time.Local)
	return []*model.Event{
		{
			ID:          uuid.NewString(),
			Title:       "Kickoff meeting",
			StartTime:   date,
			EndTime:     date.Add(time.Hour),
			UserID:      uuid.NewString(),
			NotifyDelta: 10,
		},
		{
			ID:          uuid.NewString(),
			Title:       "Kickoff meeting 2",
			StartTime:   date.Add(time.Hour),
			EndTime:     date.Add(2 * time.Hour),
			UserID:      uuid.NewString(),
			NotifyDelta: 20,
		},
		{
			ID:          uuid.NewString(),
			Title:       "Kickoff meeting 3",
			StartTime:   date.Add(24 * time.Hour),
			EndTime:     date.Add(24 * time.Hour),
			UserID:      uuid.NewString(),
			NotifyDelta: 20,
		},
	}, date
}

// FilterEventsByWeekFixture returns a slice of three events two of which are on the same week,
// and one is on the next week. The second return value is the date of the week start.
func FilterEventsByWeekFixture() ([]*model.Event, time.Time) {
	date := time.Date(2024, 9, 30, 0, 0, 0, 0, time.Local)
	return []*model.Event{
		{
			ID:          uuid.NewString(),
			Title:       "Kickoff meeting",
			StartTime:   time.Date(2024, 10, 4, 0, 0, 0, 0, time.Local),
			EndTime:     time.Date(2024, 10, 4, 1, 0, 0, 0, time.Local),
			UserID:      uuid.NewString(),
			NotifyDelta: 10,
		},
		{
			ID:          uuid.NewString(),
			Title:       "Kickoff meeting 2",
			StartTime:   time.Date(2024, 10, 5, 0, 0, 0, 0, time.Local),
			EndTime:     time.Date(2024, 10, 5, 1, 0, 0, 0, time.Local),
			UserID:      uuid.NewString(),
			NotifyDelta: 20,
		},
		{
			ID:          uuid.NewString(),
			Title:       "Kickoff meeting 3",
			StartTime:   time.Date(2024, 10, 7, 0, 0, 0, 0, time.Local),
			EndTime:     time.Date(2024, 10, 7, 1, 0, 0, 0, time.Local),
			UserID:      uuid.NewString(),
			NotifyDelta: 20,
		},
	}, date
}

// FilterEventsByMonthFixture returns a slice of three events two of which are on the same month,
// and one is on the next month. The second return value is the date of the month start.
func FilterEventsByMonthFixture() ([]*model.Event, time.Time) {
	date := time.Date(2024, 10, 1, 0, 0, 0, 0, time.Local)
	return []*model.Event{
		{
			ID:          uuid.NewString(),
			Title:       "IT Conference",
			StartTime:   time.Date(2024, 10, 11, 0, 0, 0, 0, time.Local),
			EndTime:     time.Date(2024, 10, 11, 1, 0, 0, 0, time.Local),
			UserID:      uuid.NewString(),
			NotifyDelta: 10,
		},
		{
			ID:          uuid.NewString(),
			Title:       "IT Conference 2",
			StartTime:   time.Date(2024, 10, 12, 0, 0, 0, 0, time.Local),
			EndTime:     time.Date(2024, 10, 12, 1, 0, 0, 0, time.Local),
			UserID:      uuid.NewString(),
			NotifyDelta: 20,
		},
		{
			ID:          uuid.NewString(),
			Title:       "IT Conference 3",
			StartTime:   time.Date(2024, 11, 7, 0, 0, 0, 0, time.Local),
			EndTime:     time.Date(2024, 11, 7, 1, 0, 0, 0, time.Local),
			UserID:      uuid.NewString(),
			NotifyDelta: 20,
		},
	}, date
}

func InsertEvent(
	conn *pgx.Conn, id string, title string, startTime time.Time, endTime time.Time, userID string, notifyDelta int,
) error {
	_, err := conn.Exec(context.TODO(),
		`INSERT INTO events (id, title, start_time, end_time, user_id, notify_delta) 
VALUES ($1, $2, $3, $4, $5, $6)`,
		id, title, startTime, endTime, userID, notifyDelta,
	)
	return err
}

func InsertEvents(t *testing.T, testData []*model.Event, s *Storage) {
	t.Helper()
	for _, event := range testData {
		err := InsertEvent(s.Conn, event.ID, event.Title, event.StartTime, event.EndTime, event.UserID, event.NotifyDelta)
		require.NoError(t, err)
	}
}

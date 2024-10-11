package sqlstorage

import (
	"context"
	"github.com/Azimkhan/hw12_13_14_15_calendar/internal/storage/model"
	"time"

	"github.com/jackc/pgx/v5"
)

func New(dsn string) *Storage {
	return &Storage{
		dsn: dsn,
	}
}

type Storage struct {
	dsn  string
	Conn *pgx.Conn
}

func (s *Storage) CreateEvent(ctx context.Context, event *model.Event) error {
	_, err := s.Conn.Exec(ctx,
		`
INSERT INTO events (id, title, start_time, end_time, user_id, notify_delta) 
VALUES ($1, $2, $3, $4, $5, $6)`,
		event.ID, event.Title, event.StartTime, event.EndTime, event.UserID, event.NotifyDelta)
	return err
}

func (s *Storage) UpdateEvent(ctx context.Context, event *model.Event) error {
	// write an implementation for updating an event:
	res, err := s.Conn.Exec(ctx,
		`
UPDATE events SET title = $1, start_time = $2, end_time = $3, user_id = $4, notify_delta = $5 
WHERE id = $6`,
		event.Title, event.StartTime, event.EndTime, event.UserID, event.NotifyDelta, event.ID)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return model.ErrEventNotFound
	}
	return nil
}

func (s *Storage) RemoveEvent(ctx context.Context, eventID string) error {
	res, err := s.Conn.Exec(ctx,
		"DELETE FROM events WHERE id = $1",
		eventID)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return model.ErrEventNotFound
	}
	return nil
}

func (s *Storage) FilterEventsByDay(ctx context.Context, date time.Time) ([]*model.Event, error) {
	beginningOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.Local)
	endOfDay := time.Date(date.Year(), date.Month(), date.Day(), 23, 59, 59, 0, time.Local)
	rows, err := s.Conn.Query(ctx,
		`
SELECT id, title, start_time, end_time, user_id, notify_delta 
FROM events WHERE start_time >= $1 AND start_time <= $2`,
		beginningOfDay, endOfDay)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return s.fetchRows(rows)
}

func (s *Storage) FilterEventsByWeek(ctx context.Context, weekStart time.Time) ([]*model.Event, error) {
	// define the beginning and the end of the week, weekStart can be any day of the week
	// make sure replace the time part with 00:00:00 and 23:59:59
	beginningOfWeek := time.Date(weekStart.Year(), weekStart.Month(), weekStart.Day(), 0, 0, 0, 0, time.Local)
	beginningOfWeek = beginningOfWeek.AddDate(0, 0, -int(beginningOfWeek.Weekday())+1)
	endOfWeek := beginningOfWeek.AddDate(0, 0, 6)
	endOfWeek = time.Date(endOfWeek.Year(), endOfWeek.Month(), endOfWeek.Day(), 23, 59, 59, 0, time.Local)

	rows, err := s.Conn.Query(ctx,
		`
SELECT id, title, start_time, end_time, user_id, notify_delta 
FROM events WHERE start_time >= $1 AND start_time <= $2`,
		beginningOfWeek, endOfWeek)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return s.fetchRows(rows)
}

func (s *Storage) FilterEventsByMonth(ctx context.Context, monthStart time.Time) ([]*model.Event, error) {
	// define the beginning and the end of the month
	// convert month start to local time
	localTime := monthStart.In(time.Local)
	beginningOfMonth := time.Date(localTime.Year(), localTime.Month(), 1, 0, 0, 0, 0, time.Local)
	endOfMonth := beginningOfMonth.AddDate(0, 1, -1)
	endOfMonth = time.Date(endOfMonth.Year(), endOfMonth.Month(), endOfMonth.Day(), 23, 59, 59, 0, time.Local)

	rows, err := s.Conn.Query(ctx,
		`
SELECT id, title, start_time, end_time, user_id, notify_delta 
FROM events WHERE start_time >= $1 AND start_time <= $2`,
		beginningOfMonth, endOfMonth)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return s.fetchRows(rows)
}

func (s *Storage) fetchRows(rows pgx.Rows) ([]*model.Event, error) {
	events := make([]*model.Event, 0)
	var err error
	for rows.Next() {
		event := &model.Event{}
		err = rows.Scan(&event.ID, &event.Title, &event.StartTime, &event.EndTime, &event.UserID, &event.NotifyDelta)
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}

	return events, nil
}

func (s *Storage) Connect(ctx context.Context) error {
	conn, err := pgx.Connect(ctx, s.dsn)
	if err != nil {
		return err
	}
	if err = conn.Ping(ctx); err != nil {
		return err
	}
	s.Conn = conn
	return nil
}

func (s *Storage) Close(ctx context.Context) error {
	if s.Conn != nil {
		err := s.Conn.Close(ctx)
		if err != nil {
			return err
		}
		s.Conn = nil
	}
	return nil
}

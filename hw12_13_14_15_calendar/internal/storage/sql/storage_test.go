package sqlstorage

import (
	"context"
	"testing"
	"time"

	"github.com/Azimkhan/hw-golang/hw12_13_14_15_calendar/internal/storage/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	testDBName     = "test-db"
	testDBUser     = "postgres"
	testDBPassword = "postgres"
	testDBImage    = "postgres:16"
)

func TestCreate(t *testing.T) {
	ctx := context.Background()
	connStr, err := createPostgresContainer(ctx, t)
	require.NoError(t, err)
	s := createStorage(t, connStr)

	migrateDB(ctx, t, s)
	require.NoError(t, err, "migration failed")
	event := model.Event{
		ID:          uuid.NewString(),
		Title:       "Kickoff meeting",
		StartTime:   time.Now(),
		EndTime:     time.Now().Add(time.Hour),
		UserID:      uuid.NewString(),
		NotifyDelta: 10,
	}
	err = s.CreateEvent(context.TODO(), &event)
	require.NoError(t, err)
	// check that event was created
	row := s.Conn.QueryRow(
		context.TODO(),
		"SELECT id, title, start_time, end_time, user_id, notify_delta FROM events WHERE id = $1",
		event.ID,
	)

	verifyEvent(t, row, event)
}

func TestUpdate(t *testing.T) {
	ctx := context.Background()
	connStr, err := createPostgresContainer(ctx, t)
	require.NoError(t, err)

	s := createStorage(t, connStr)

	migrateDB(ctx, t, s)
	require.NoError(t, err)

	// insert sample event
	uid, err := createTestEvent(s.Conn)
	require.NoError(t, err)

	// update event
	event := model.Event{
		ID:          uid,
		Title:       "Kickoff meeting 2",
		StartTime:   time.Now().Add(time.Hour),
		EndTime:     time.Now().Add(2 * time.Hour),
		UserID:      uuid.NewString(),
		NotifyDelta: 20,
	}
	err = s.UpdateEvent(context.TODO(), &event)
	require.NoError(t, err)

	// check that event was updated
	row := s.Conn.QueryRow(
		context.TODO(),
		`
SELECT id, title, start_time, end_time, user_id, notify_delta FROM events
WHERE id = $1`,
		event.ID,
	)

	verifyEvent(t, row, event)
}

func TestRemove(t *testing.T) {
	ctx := context.Background()
	connStr, err := createPostgresContainer(ctx, t)
	require.NoError(t, err)
	s := createStorage(t, connStr)

	migrateDB(ctx, t, s)
	require.NoError(t, err, "migration failed")

	// insert sample event
	uid, err := createTestEvent(s.Conn)
	require.NoError(t, err)

	// remove event
	err = s.RemoveEvent(context.TODO(), uid)
	require.NoError(t, err)

	// check that event was removed
	row := s.Conn.QueryRow(
		context.TODO(),
		"SELECT count(*) FROM events WHERE id = $1",
		uid,
	)
	var count int
	err = row.Scan(&count)
	require.NoError(t, err)
}

func TestFilterEventsByDay(t *testing.T) {
	ctx := context.Background()
	connStr, err := createPostgresContainer(ctx, t)
	require.NoError(t, err)
	s := createStorage(t, connStr)

	migrateDB(ctx, t, s)
	require.NoError(t, err, "migration failed")

	testData, date := FilterEventsByDateFixture()
	InsertEvents(t, testData, s)

	// filter events by day
	events, err := s.FilterEventsByDay(context.TODO(), date)
	require.NoError(t, err)
	require.Len(t, events, 2)

	compareEvents(t, testData[:2], events)
}

func TestFilterEventsByWeek(t *testing.T) {
	ctx := context.Background()
	connStr, err := createPostgresContainer(ctx, t)
	require.NoError(t, err)
	s := createStorage(t, connStr)

	migrateDB(ctx, t, s)
	require.NoError(t, err, "migration failed")

	// define test data
	testData, weekStart := FilterEventsByWeekFixture()
	InsertEvents(t, testData, s)

	// filter events by week
	events, err := s.FilterEventsByWeek(context.TODO(), weekStart)
	require.NoError(t, err)
	require.Len(t, events, 2)

	compareEvents(t, testData[:2], events)
}

func TestFilterEventsByMonth(t *testing.T) {
	ctx := context.Background()
	connStr, err := createPostgresContainer(ctx, t)
	require.NoError(t, err)
	s := createStorage(t, connStr)
	migrateDB(ctx, t, s)
	require.NoError(t, err, "migration failed")

	// define test data
	testData, monthStart := FilterEventsByMonthFixture()
	for _, event := range testData {
		err = InsertEvent(s.Conn, event.ID, event.Title, event.StartTime, event.EndTime, event.UserID, event.NotifyDelta)
		require.NoError(t, err)
	}

	// filter events by month

	events, err := s.FilterEventsByMonth(context.TODO(), monthStart)
	require.NoError(t, err)
	require.Len(t, events, 2)

	compareEvents(t, testData[:2], events)
}

func compareEvents(t *testing.T, expected []*model.Event, actual []*model.Event) {
	t.Helper()
	require.Len(t, actual, len(expected))
	eventIDs := make(map[string]struct{})
	for _, event := range expected {
		eventIDs[event.ID] = struct{}{}
	}
	for _, event := range actual {
		_, ok := eventIDs[event.ID]
		require.True(t, ok)
	}
}

func createTestEvent(conn *pgx.Conn) (string, error) {
	uid := uuid.NewString()
	err := InsertEvent(conn, uid, "Kickoff meeting", time.Now(), time.Now().Add(time.Hour), uuid.NewString(), 10)
	return uid, err
}

func verifyEvent(t *testing.T, row pgx.Row, event model.Event) {
	t.Helper()
	var id, title, userID string
	var startTime, endTime time.Time
	var notifyDelta int
	err := row.Scan(&id, &title, &startTime, &endTime, &userID, &notifyDelta)

	require.NoError(t, err)
	require.Equal(t, event.ID, id)
	require.Equal(t, event.Title, title)
	require.Equal(t, event.UserID, userID)
	require.Equal(t, event.NotifyDelta, notifyDelta)
	require.WithinDuration(t, event.StartTime, startTime, time.Second)
	require.WithinDuration(t, event.EndTime, endTime, time.Second)
}

func migrateDB(ctx context.Context, t *testing.T, s *Storage) {
	t.Helper()
	err := s.Migrate(ctx, nil)
	require.NoError(t, err)
}

func createPostgresContainer(ctx context.Context, t *testing.T) (string, error) {
	t.Helper()
	pgContainer, err := postgres.Run(ctx,
		testDBImage,
		postgres.WithDatabase(testDBName),
		postgres.WithUsername(testDBUser),
		postgres.WithPassword(testDBPassword),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).WithStartupTimeout(5*time.Second)),
	)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		if err := pgContainer.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate pgContainer: %s", err)
		}
	})
	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	return connStr, err
}

func createStorage(t *testing.T, connStr string) *Storage {
	t.Helper()
	s := New(connStr)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err := s.Connect(ctx)
	require.NoError(t, err, "failed to connect to db")
	return s
}

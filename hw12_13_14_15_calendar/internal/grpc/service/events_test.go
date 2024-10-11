package service

import (
	"context"
	"log"
	"net"
	"testing"
	"time"

	"github.com/Azimkhan/hw12_13_14_15_calendar/gen/events/pb"
	"github.com/Azimkhan/hw12_13_14_15_calendar/internal/app"
	"github.com/Azimkhan/hw12_13_14_15_calendar/internal/logger"
	sqlstorage "github.com/Azimkhan/hw12_13_14_15_calendar/internal/storage/sql"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	selectCountStatement = "SELECT count(*) FROM events WHERE" +
		" id = $1"
	selectStatement = "SELECT id, title, start_time, end_time, user_id, notify_delta " +
		"FROM events WHERE id = $1"
	insertStatement = "INSERT INTO events (id, title, start_time, end_time, user_id, notify_delta) " +
		"VALUES ($1, $2, $3, $4, $5, $6)"
)

func checkEvent(t *testing.T, event *pb.Event, row pgx.Row) {
	t.Helper()
	var id, title, userID string
	var startTime, endTime time.Time
	var notifyDelta int
	err := row.Scan(
		&id, &title, &startTime, &endTime, &userID, &notifyDelta,
	)
	require.NoError(t, err)
	require.Equal(t, event.Id, id)
	require.Equal(t, event.Title, title)
	require.Equal(t, event.UserId, userID)
	require.Equal(t, int(event.NotifyDelta), notifyDelta)
	require.WithinDuration(t, event.Start.AsTime(), startTime, time.Second)
	require.WithinDuration(t, event.End.AsTime(), endTime, time.Second)
}

func TestCreate(t *testing.T) {
	ctx := context.Background()
	testApp, pgStorage := createApp(ctx, t)

	client := testServer(ctx, t, testApp)

	eventID := uuid.NewString()
	event := &pb.Event{
		Id:          eventID,
		Title:       "Kickoff meeting",
		Start:       timestamppb.New(time.Now()),
		End:         timestamppb.New(time.Now().Add(time.Hour)),
		UserId:      uuid.NewString(),
		NotifyDelta: 10,
	}
	response, err := client.CreateEvent(context.Background(), &pb.CreateEventRequest{Event: event})
	require.NoError(t, err)

	require.NotNil(t, response)

	row := pgStorage.Conn.QueryRow(context.TODO(),
		selectStatement,
		eventID,
	)

	checkEvent(t, event, row)
}

func TestUpdate(t *testing.T) {
	ctx := context.Background()
	testApp, pgStorage := createApp(ctx, t)

	eventID := uuid.NewString()
	_, err := pgStorage.Conn.Exec(context.TODO(),
		insertStatement,
		eventID, "Kickoff meeting", time.Now(), time.Now().Add(time.Hour), uuid.NewString(), 10,
	)

	require.NoError(t, err)
	client := testServer(ctx, t, testApp)

	event := &pb.Event{
		Id:          eventID,
		Title:       "Kickoff meeting 2",
		Start:       timestamppb.New(time.Now()),
		End:         timestamppb.New(time.Now().Add(time.Hour)),
		UserId:      uuid.NewString(),
		NotifyDelta: 20,
	}
	response, err := client.UpdateEvent(context.Background(), &pb.UpdateEventRequest{Event: event})
	require.NoError(t, err)

	require.NotNil(t, response)

	row := pgStorage.Conn.QueryRow(context.TODO(),
		selectStatement,
		eventID,
	)

	checkEvent(t, event, row)
}

func TestRemove(t *testing.T) {
	ctx := context.Background()
	testApp, pgStorage := createApp(ctx, t)

	eventID := uuid.NewString()
	_, err := pgStorage.Conn.Exec(context.TODO(),
		insertStatement,
		eventID, "Kickoff meeting", time.Now(), time.Now().Add(time.Hour), uuid.NewString(), 10,
	)

	require.NoError(t, err)
	client := testServer(ctx, t, testApp)

	response, err := client.RemoveEvent(context.Background(), &pb.RemoveEventRequest{Id: eventID})

	require.NoError(t, err)
	require.NotNil(t, response)

	row := pgStorage.Conn.QueryRow(context.TODO(),
		selectCountStatement,
		eventID,
	)
	count := 1
	err = row.Scan(&count)
	require.NoError(t, err)
	require.Equal(t, 0, count)
}

func TestFilterEventsByDay(t *testing.T) {
	ctx := context.Background()
	testApp, pgStorage := createApp(ctx, t)

	client := testServer(ctx, t, testApp)
	events, date := sqlstorage.FilterEventsByDateFixture()
	sqlstorage.InsertEvents(t, events, pgStorage)

	response, err := client.FilterEventsByDay(
		context.Background(),
		&pb.FilterEventsByDayRequest{Date: timestamppb.New(date)},
	)

	require.NoError(t, err)
	require.NotNil(t, response)
	require.Len(t, response.Events, 2)
}

func TestFilterEventsByWeek(t *testing.T) {
	ctx := context.Background()
	testApp, pgStorage := createApp(ctx, t)

	client := testServer(ctx, t, testApp)
	events, weekStart := sqlstorage.FilterEventsByWeekFixture()
	sqlstorage.InsertEvents(t, events, pgStorage)

	response, err := client.FilterEventsByWeek(
		context.Background(),
		&pb.FilterEventsByWeekRequest{Date: timestamppb.New(weekStart)},
	)

	require.NoError(t, err)
	require.NotNil(t, response)
	require.Len(t, response.Events, 2)
}

func TestFilterEventsByMonth(t *testing.T) {
	ctx := context.Background()
	testApp, pgStorage := createApp(ctx, t)

	client := testServer(ctx, t, testApp)
	events, monthStart := sqlstorage.FilterEventsByMonthFixture()
	sqlstorage.InsertEvents(t, events, pgStorage)

	response, err := client.FilterEventsByMonth(
		context.Background(),
		&pb.FilterEventsByMonthRequest{Date: timestamppb.New(monthStart)},
	)

	require.NoError(t, err)
	require.NotNil(t, response)
	require.Len(t, response.Events, 2)
}

func createApp(ctx context.Context, t *testing.T) (*app.App, *sqlstorage.Storage) {
	t.Helper()
	logg, err := logger.New("DEBUG")
	require.NoError(t, err, "failed to create logger")
	pgContainer, err := postgres.Run(ctx,
		"postgres:16",
		// postgres.WithInitScripts(filepath.Join("..", "testdata", "init-db.sql")),
		postgres.WithDatabase("test-db"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
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
	// return connStr, err
	pgStorage := sqlstorage.New(connStr)
	err = pgStorage.Connect(ctx)
	t.Cleanup(func() {
		if err := pgStorage.Close(ctx); err != nil {
			t.Fatalf("failed to close pgStorage: %s", err)
		}
	})
	require.NoError(t, err, "failed to connect to db")
	err = pgStorage.Migrate(ctx, nil)
	require.NoError(t, err, "migration failed")
	testApp := app.New(logg, pgStorage)
	return testApp, pgStorage
}

func testServer(_ context.Context, t *testing.T, testApp *app.App) pb.EventServiceClient {
	t.Helper()
	lis := bufconn.Listen(101024 * 1024)
	baseServer := grpc.NewServer()
	pb.RegisterEventServiceServer(baseServer, NewEventsService(testApp))
	go func() {
		if err := baseServer.Serve(lis); err != nil {
			log.Printf("error serving server: %v", err)
		}
	}()

	t.Cleanup(func() {
		err := lis.Close()
		if err != nil {
			log.Printf("error closing listener: %v", err)
		}
		baseServer.Stop()
	})
	conn, err := grpc.NewClient(
		"passthrough://bufnet",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return lis.Dial()
		}), grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	client := pb.NewEventServiceClient(conn)
	return client
}

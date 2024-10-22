package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/Azimkhan/hw-golang/hw12_13_14_15_calendar/gen/events/pb"
	"github.com/Azimkhan/hw-golang/hw12_13_14_15_calendar/internal/app"
	"github.com/Azimkhan/hw-golang/hw12_13_14_15_calendar/internal/conf"
	"github.com/Azimkhan/hw-golang/hw12_13_14_15_calendar/internal/grpc"
	"github.com/Azimkhan/hw-golang/hw12_13_14_15_calendar/internal/logger"
	"github.com/Azimkhan/hw-golang/hw12_13_14_15_calendar/internal/messages"
	"github.com/Azimkhan/hw-golang/hw12_13_14_15_calendar/internal/scheduler"
	"github.com/Azimkhan/hw-golang/hw12_13_14_15_calendar/internal/sender"
	_ "github.com/Azimkhan/hw-golang/hw12_13_14_15_calendar/internal/storage"
	sqlstorage "github.com/Azimkhan/hw-golang/hw12_13_14_15_calendar/internal/storage/sql"
	"github.com/jackc/pgx/v5"
	"github.com/streadway/amqp"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	grpc2 "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const dsn = "host=postgres user=calendar_test password=calendar_test dbname=calendar_test sslmode=disable"

const (
	notificationTestQueue = "notificationTester"
	amqpURI               = "amqp://guest:guest@rabbitmq:5672/"
	grpcAddr              = "localhost:50051"
)

var amqpConfig = conf.AMQPConfig{
	URI:          amqpURI,
	Exchange:     "calendar",
	ExchangeType: "direct",
	Queue:        "notification",
	RoutingKey:   "notification",
}

var storageConf = conf.StorageConf{
	Type: "sql",
	DSN:  dsn,
}

type MainTestSuite struct {
	suite.Suite
	s          *sqlstorage.Storage
	pgConn     *pgx.Conn
	amqpConn   *amqp.Connection
	amqpCh     *amqp.Channel
	grpcServer *grpc.Server
}

func (suite *MainTestSuite) SetupSuite() {
	suite.setupPostgres()

	// AMQP connection
	suite.setupAMQP()

	// Create app
	application := suite.createApp()

	// Create gRPC server
	grpcServer, err := grpc.NewServer(application, &conf.GRPCConf{
		BindAddr: grpcAddr,
	})
	require.NoError(suite.T(), err)
	go func() {
		require.NoError(suite.T(), grpcServer.Serve())
	}()
	suite.grpcServer = grpcServer
}

func (suite *MainTestSuite) createApp() *app.App {
	log, err := logger.New("DEBUG")
	require.NoError(suite.T(), err)

	storage := sqlstorage.New(dsn)
	require.NoError(suite.T(), storage.Connect(context.Background()))
	suite.s = storage

	application := app.New(log, storage)
	return application
}

func (suite *MainTestSuite) setupAMQP() {
	conn, err := amqp.Dial(amqpURI)
	require.NoError(suite.T(), err)
	ch, err := conn.Channel()
	require.NoError(suite.T(), err)
	suite.amqpConn = conn
	suite.amqpCh = ch
}

func (suite *MainTestSuite) declareExchange() {
	if err := suite.amqpCh.ExchangeDeclare(
		amqpConfig.Exchange,
		amqpConfig.ExchangeType,
		true,  // durable
		false, // auto-deleted
		false, // internal
		false, // noWait
		nil,   // arguments
	); err != nil {
		require.NoError(suite.T(), err)
	}

	if _, err := suite.amqpCh.QueueDeclare(
		amqpConfig.Queue,
		true,  // durable
		false, // autoDelete
		false, // exclusive
		false, // noWait
		nil,   // args
	); err != nil {
		require.NoError(suite.T(), err)
	}

	if err := suite.amqpCh.QueueBind(
		amqpConfig.Queue,
		amqpConfig.RoutingKey,
		amqpConfig.Exchange,
		false,
		nil,
	); err != nil {
		require.NoError(suite.T(), err)
	}
}

func (suite *MainTestSuite) setupPostgres() {
	// Connect to the database
	pgConn, err := pgx.Connect(context.Background(), dsn)
	require.NoError(suite.T(), err)
	suite.pgConn = pgConn
}

func (suite *MainTestSuite) TearDownSuite() {
	if err := suite.s.Close(context.Background()); err != nil {
		suite.T().Fatal(err)
	}

	if err := suite.amqpCh.Close(); err != nil {
		suite.T().Fatal(err)
	}

	if err := suite.amqpConn.Close(); err != nil {
		suite.T().Fatal(err)
	}
}

func (suite *MainTestSuite) SetupTest() {
	// Redeclare exchange
	_ = suite.amqpCh.ExchangeDelete(amqpConfig.Exchange, false, false)
	suite.declareExchange()

	// Recreate schema and migrate
	_, err := suite.pgConn.Exec(context.Background(), "DROP SCHEMA public CASCADE")
	require.NoError(suite.T(), err)

	_, err = suite.pgConn.Exec(context.Background(), "CREATE SCHEMA public")
	require.NoError(suite.T(), err)

	if err := suite.s.Migrate(context.Background(), nil); err != nil {
		suite.T().Fatal(err)
	}
}

func (suite *MainTestSuite) startConsumingNotifications() (<-chan *messages.Notification, error) {
	successChan := make(chan *messages.Notification)

	if _, err := suite.amqpCh.QueueDeclare(
		notificationTestQueue,
		true,  // durable
		false, // autoDelete
		false, // exclusive
		false, // noWait
		nil,   // args
	); err != nil {
		return nil, err
	}

	if err := suite.amqpCh.QueueBind(
		notificationTestQueue,
		amqpConfig.RoutingKey,
		amqpConfig.Exchange,
		false,
		nil,
	); err != nil {
		return nil, err
	}

	go func() {
		defer close(successChan)
		msgs, err := suite.amqpCh.Consume(
			notificationTestQueue,
			"",
			true,
			false,
			false,
			false,
			nil,
		)
		if err != nil {
			suite.T().Log("Failed to consume messages", err)
			return
		}
		for msgBytes := range msgs {
			var notification messages.Notification
			if err := json.Unmarshal(msgBytes.Body, &notification); err != nil {
				suite.T().Log("Failed to unmarshal message", err)
				continue
			}
			successChan <- &notification
			return
		}
	}()
	return successChan, nil
}

// TestNotificationIsSent tests that the notification is sent
// when new event is created.
func (suite *MainTestSuite) TestNotificationIsSent() {
	// Start consuming notifications
	successChan, err := suite.startConsumingNotifications()
	require.NoError(suite.T(), err)

	// Create scheduler
	sched := scheduler.New(&conf.SchedulerConfig{
		CleanInterval:      1800,
		CleanThresholdDays: 180,
		ScanInterval:       1,
		Logger: conf.LoggerConf{
			Level: "DEBUG",
		},
		Storage: storageConf,
		AMQP:    amqpConfig,
	})
	go func() {
		err := sched.Run(context.Background())
		require.NoError(suite.T(), err)
	}()
	defer func() {
		require.NoError(suite.T(), sched.Stop())
	}()
	// Send gRPC request
	client := suite.createGRPCClient()
	response, err := client.CreateEvent(context.Background(), &pb.CreateEventRequest{
		Event: &pb.Event{
			Id:          "1",
			Title:       "test",
			Start:       timestamppb.New(time.Now().Add(3 * time.Second)),
			End:         timestamppb.New(time.Now()),
			UserId:      "1",
			NotifyDelta: 1,
		},
	})
	// Ensure the event is created
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), response)
	require.NotNil(suite.T(), response.Event)
	require.Equal(suite.T(), "1", response.Event.Id)

	result := suite.pgConn.QueryRow(context.Background(), "SELECT count(*) FROM events WHERE id = $1", "1")
	var count int
	require.NoError(suite.T(), result.Scan(&count))
	require.Equal(suite.T(), 1, count)

	// Wait for the notification
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	require.NoError(suite.T(), err)
	defer cancel()

	var notification *messages.Notification
	// Test that the notification is sent from the scheduler
	select {
	case <-ctx.Done():
		require.Fail(suite.T(), "Notification read timeout")

	case notification = <-successChan:
		fmt.Println("Notification received")
	}

	// verify the notification
	require.NotNil(suite.T(), notification)
	require.Equal(suite.T(), "1", notification.EventID)
	require.Equal(suite.T(), "test", notification.Title)
	require.Equal(suite.T(), "1", notification.UserID)
}

func (suite *MainTestSuite) createGRPCClient() pb.EventServiceClient {
	grpcClient, err := grpc2.NewClient(grpcAddr, grpc2.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(suite.T(), err)
	client := pb.NewEventServiceClient(grpcClient)
	return client
}

// TestNotificationIsConsumed tests that the notification is consumed
// when the event is sent to the sender.
func (suite *MainTestSuite) TestNotificationIsConsumed() {
	application := sender.New(&conf.SenderConfig{
		Logger: conf.LoggerConf{
			Level: "DEBUG",
		},
		AMQP:    amqpConfig,
		Storage: storageConf,
	})
	go func() {
		err := application.Run(context.Background())
		require.NoError(suite.T(), err)
	}()
	defer func() {
		require.NoError(suite.T(), application.Stop())
	}()

	// create event
	eventID := "2"
	_, err := suite.pgConn.Exec(
		context.Background(),
		"INSERT INTO events (id, title, start_time, end_time, user_id, notify_delta, notification_sent) "+
			"VALUES ($1, $2, $3, $4, $5, $6, $7)",
		eventID, "Meeting", time.Now(), time.Now(), "3", 0, false,
	)
	require.NoError(suite.T(), err)
	err = suite.amqpCh.Publish(
		amqpConfig.Exchange,
		amqpConfig.RoutingKey,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        []byte(`{"eventId": "2", "title": "Meeting", "userId": "3", "startTime": "2024-10-10T10:00:00Z"}`),
		},
	)
	require.NoError(suite.T(), err)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			require.Fail(suite.T(), "Notification read timeout")
		case <-ticker.C:
			result := suite.pgConn.QueryRow(
				context.Background(),
				"SELECT notification_sent FROM events WHERE id = $1",
				eventID,
			)
			var notificationSent bool
			require.NoError(suite.T(), result.Scan(&notificationSent))
			if notificationSent {
				return
			}
		}
	}
}

// TestEventFiltering tests that events are filtered correctly by day, week, and month.
func (suite *MainTestSuite) TestEventFiltering() {
	// table test
	testData := []struct {
		name        string
		fixtureFile string
		expectedID  []string
		date        time.Time
		filter      func(context.Context, pb.EventServiceClient, time.Time) ([]*pb.Event, error)
	}{
		{
			name:        "day",
			fixtureFile: "../fixtures/events_day.sql",
			expectedID:  []string{"event1", "event2"},
			date:        time.Date(2024, 10, 13, 0, 0, 0, 0, time.UTC),
			filter: func(ctx context.Context, client pb.EventServiceClient, date time.Time) ([]*pb.Event, error) {
				response, err := client.FilterEventsByDay(ctx, &pb.FilterEventsByDayRequest{
					Date: timestamppb.New(date),
				})
				if err != nil {
					return nil, err
				}
				return response.Events, nil
			},
		},
		{
			name:        "week",
			fixtureFile: "../fixtures/events_week.sql",
			expectedID:  []string{"event77", "event81"},
			date:        time.Date(2024, 10, 7, 0, 0, 0, 0, time.UTC),
			filter: func(ctx context.Context, client pb.EventServiceClient, date time.Time) ([]*pb.Event, error) {
				response, err := client.FilterEventsByWeek(ctx, &pb.FilterEventsByWeekRequest{
					Date: timestamppb.New(date),
				})
				if err != nil {
					return nil, err
				}
				return response.Events, nil
			},
		},
		{
			name:        "month",
			fixtureFile: "../fixtures/events_month.sql",
			expectedID:  []string{"event101", "event107"},
			date:        time.Date(2024, 10, 1, 0, 0, 0, 0, time.UTC),
			filter: func(ctx context.Context, client pb.EventServiceClient, date time.Time) ([]*pb.Event, error) {
				response, err := client.FilterEventsByMonth(ctx, &pb.FilterEventsByMonthRequest{
					Date: timestamppb.New(date),
				})
				if err != nil {
					return nil, err
				}
				return response.Events, nil
			},
		},
	}

	client := suite.createGRPCClient()
	for _, data := range testData {
		suite.SetupTest()
		suite.T().Run(data.name, func(t *testing.T) {
			// Load fixtures
			query, err := os.ReadFile(data.fixtureFile)
			require.NoError(t, err)
			_, err = suite.pgConn.Exec(context.Background(), string(query))
			require.NoError(t, err)

			// Filter events
			events, err := data.filter(context.Background(), client, data.date)
			require.NoError(t, err)

			// Check the number of events
			require.Len(t, events, len(data.expectedID))
			// collect IDs
			var ids []string
			for _, event := range events {
				ids = append(ids, event.Id)
			}
			// compare IDs
			require.ElementsMatch(t, data.expectedID, ids)
		})
	}
}

func TestMainTestSuite(t *testing.T) {
	suite.Run(t, new(MainTestSuite))
}

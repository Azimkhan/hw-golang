package integration

// Basic imports
import (
	"context"
	"encoding/json"
	"fmt"
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
	"testing"
	"time"
)

const dsn = "host=postgres user=calendar_test password=calendar_test dbname=calendar_test sslmode=disable"

const notificationTestQueue = "notificationTester"
const amqpURI = "amqp://guest:guest@rabbitmq:5672/"
const grpcAddr = "localhost:50051"

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

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type MainTestSuite struct {
	suite.Suite
	s          *sqlstorage.Storage
	pgConn     *pgx.Conn
	tx         pgx.Tx
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

// before each test
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
			suite.T().Fatal(err)
		}
		for msgBytes := range msgs {
			var notification messages.Notification
			if err := json.Unmarshal(msgBytes.Body, &notification); err != nil {
				suite.T().Fatal(err)
			}
			successChan <- &notification
			return
		}
	}()
	return successChan, nil
}

// == Test cases ==
func (suite *MainTestSuite) TestNotificationIsSent() {
	// Start consuming notifications
	successChan, err := suite.startConsumingNotifications()
	require.NoError(suite.T(), err)

	// Create scheduler
	sched := scheduler.New(&conf.SchedulerConfig{
		CleanInterval:      1800,
		CleanThresholdDays: 180,
		ScanInterval:       2,
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
	grpcClient, err := grpc2.NewClient(grpcAddr, grpc2.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(suite.T(), err)
	pb.NewEventServiceClient(grpcClient)
	client := pb.NewEventServiceClient(grpcClient)
	response, err := client.CreateEvent(context.Background(), &pb.CreateEventRequest{
		Event: &pb.Event{
			Id:          "1",
			Title:       "test",
			Start:       timestamppb.New(time.Now().Add(5 * time.Second)),
			End:         timestamppb.New(time.Now()),
			UserId:      "1",
			NotifyDelta: 0,
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
	eventId := "2"
	_, err := suite.pgConn.Exec(
		context.Background(),
		"INSERT INTO events (id, title, start_time, end_time, user_id, notify_delta, notification_sent) VALUES ($1, $2, $3, $4, $5, $6, $7)",
		eventId, "Meeting", time.Now(), time.Now(), "3", 0, false,
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
			result := suite.pgConn.QueryRow(context.Background(), "SELECT notification_sent FROM events WHERE id = $1", eventId)
			var notificationSent bool
			require.NoError(suite.T(), result.Scan(&notificationSent))
			if notificationSent {
				return
			}
		}
	}

}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestExampleTestSuite(t *testing.T) {
	suite.Run(t, new(MainTestSuite))
}

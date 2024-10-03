package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Azimkhan/hw12_13_14_15_calendar/internal/app"
	"github.com/Azimkhan/hw12_13_14_15_calendar/internal/logger"
	internalhttp "github.com/Azimkhan/hw12_13_14_15_calendar/internal/server/http"
	memorystorage "github.com/Azimkhan/hw12_13_14_15_calendar/internal/storage/memory"
	sqlstorage "github.com/Azimkhan/hw12_13_14_15_calendar/internal/storage/sql"
)

var configFile string

func init() {
	flag.StringVar(&configFile, "config", "/etc/calendar/config.toml", "Path to configuration file")
}

func main() {
	flag.Parse()

	if flag.Arg(0) == "version" {
		printVersion()
		return
	}

	config := NewConfig()
	if err := config.LoadFromFile(configFile); err != nil {
		log.Fatalf("failed to load config: %v", err)
	}
	logg, err := logger.New(config.Logger.Level)
	if err != nil {
		log.Fatalf("failed to create logger: %v", err)
	}

	var storage app.Storage
	switch config.Storage.Type {
	case "inmemory":
		storage = memorystorage.New()
	case "sql":
		timeout, cancelFunc := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancelFunc()
		pgStorage := sqlstorage.New(config.Storage.DSN)
		if err := pgStorage.Connect(timeout); err != nil {
			logg.Error("failed to connect to db: " + err.Error())
			os.Exit(1) //nolint:gocritic
		}
		defer func() {
			timeout, cancelFunc := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancelFunc()
			if err := pgStorage.Close(timeout); err != nil {
				logg.Error("failed to close connection to db: " + err.Error())
			}
		}()
		storage = pgStorage
	default:
		panic("unknown storage type")
	}
	calendar := app.New(logg, storage, config.HTTP.BindAddr)

	server := internalhttp.NewServer(logg, calendar)

	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()

	go func() {
		<-ctx.Done()

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
		defer cancel()

		if err := server.Stop(ctx); err != nil {
			logg.Error("failed to stop http server: " + err.Error())
		}
	}()

	logg.Info("calendar is running...")

	if err := server.Start(ctx); err != nil {
		logg.Error("failed to start http server: " + err.Error())
		cancel()
		os.Exit(1)
	}
}

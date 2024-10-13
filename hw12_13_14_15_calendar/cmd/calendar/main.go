package main

import (
	"context"
	"flag"
	"fmt"
	"os/signal"
	"syscall"
	"time"

	"github.com/Azimkhan/hw-golang/hw12_13_14_15_calendar/internal/app"
	"github.com/Azimkhan/hw-golang/hw12_13_14_15_calendar/internal/conf"
	appGrpc "github.com/Azimkhan/hw-golang/hw12_13_14_15_calendar/internal/grpc"
	"github.com/Azimkhan/hw-golang/hw12_13_14_15_calendar/internal/logger"
	internalhttp "github.com/Azimkhan/hw-golang/hw12_13_14_15_calendar/internal/server/http"
	"github.com/Azimkhan/hw-golang/hw12_13_14_15_calendar/internal/storage"
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
	// load config
	config := conf.NewConfig()
	if err := config.LoadFromFile(configFile); err != nil {
		fmt.Println("failed to load config: " + err.Error())
		return
	}

	// create logger
	logg, err := logger.New(config.Logger.Level)
	if err != nil {
		fmt.Println("failed to create logger: " + err.Error())
		return
	}

	// create storage
	s, closeFunc, err := storage.NewFromConfig(&config.Storage)
	if err != nil {
		logg.Error("failed to create storage: " + err.Error())
		return
	}
	if closeFunc != nil {
		defer func() {
			timeout, cancelFunc := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancelFunc()
			if err := closeFunc(timeout); err != nil {
				logg.Error("failed to close storage: " + err.Error())
			}
		}()
	}

	// create app
	calendar := app.New(logg, s)

	// create context
	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()

	// create gRPC server
	grpcServer, err := appGrpc.NewServer(calendar, &config.GRPC)
	if err != nil {
		logg.Error("failed to create gRPC server: " + err.Error())
		return
	}

	// run gRPC server
	go func() {
		if err := grpcServer.Serve(); err != nil {
			logg.Error("failed to start gRPC server: " + err.Error())
			return
		}
	}()

	// create http server
	gwmux, err := grpcServer.CreateGatewayMux(ctx)
	if err != nil {
		logg.Error("failed to register gateway: " + err.Error())
		return
	}

	httpServer := internalhttp.NewServer(logg, gwmux.ServeHTTP, calendar, config.HTTP.BindAddr)

	// signal handling
	go func() {
		<-ctx.Done()

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
		defer cancel()

		logg.Info("Signal received, stopping servers...")
		grpcServer.Stop()
		if err := httpServer.Stop(ctx); err != nil {
			logg.Error("failed to stop http server: " + err.Error())
		}
	}()

	if err := httpServer.Start(ctx); err != nil {
		logg.Error("failed to start http server: " + err.Error())
		cancel()
	}
}

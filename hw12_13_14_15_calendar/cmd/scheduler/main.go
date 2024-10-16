package main

import (
	"context"
	"flag"
	"fmt"
	"os/signal"
	"syscall"

	"github.com/Azimkhan/hw-golang/hw12_13_14_15_calendar/internal/conf"
	"github.com/Azimkhan/hw-golang/hw12_13_14_15_calendar/internal/scheduler"
	_ "github.com/streadway/amqp"
)

var configFile string

func init() {
	flag.StringVar(&configFile, "config", "/etc/calendar/scheduler_config.toml", "Path to configuration file")
}

func main() {
	flag.Parse()
	config := conf.NewSchedulerConfig()
	if err := config.LoadFromFile(configFile); err != nil {
		fmt.Printf("failed to load config: %s", err)
		return
	}

	app := scheduler.New(&config)
	// create context
	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()
	go func() {
		<-ctx.Done()
		if err := app.Stop(); err != nil {
			fmt.Printf("failed to stop app: %s", err)
		}
	}()
	if err := app.Run(ctx); err != nil {
		fmt.Printf("failed to run app: %s", err)
		return
	}
}

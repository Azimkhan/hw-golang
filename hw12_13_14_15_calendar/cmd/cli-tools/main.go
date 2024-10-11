package main

import (
	"context"
	"flag"
	"log"
	"time"

	"github.com/Azimkhan/hw-golang/hw12_13_14_15_calendar/internal/conf"
	sqlstorage "github.com/Azimkhan/hw-golang/hw12_13_14_15_calendar/internal/storage/sql"
)

var configFile string

func init() {
	flag.StringVar(&configFile, "config", "/etc/calendar/config.toml", "Path to configuration file")
}

func main() {
	flag.Parse()
	config := conf.NewConfig()
	if err := config.LoadFromFile(configFile); err != nil {
		log.Fatal("failed to load config: " + err.Error())
	}

	if flag.Arg(0) == "migrate" {
		migrate(&config)
	}
}

func migrate(config *conf.Config) {
	log.Println("Migration started")
	if config.Storage.Type != "sql" {
		log.Fatal("migrate is only supported for sql storage")
	}
	ctx := context.Background()
	s := sqlstorage.New(config.Storage.DSN)
	defer func() {
		if err := s.Close(ctx); err != nil {
			log.Fatal("failed to close storage: " + err.Error())
		}
	}()
	if err := s.Connect(ctx); err != nil {
		log.Printf("failed to connect to storage: %s\n", err)
		return
	}
	migrationCallback := func(_ int32, name, direction, sql string) {
		log.Printf(
			"%s executing %s %s\n%s\n\n", time.Now().Format("2006-01-02 15:04:05"), name, direction, sql,
		)
	}

	if err := s.Migrate(ctx, migrationCallback); err != nil {
		log.Printf("failed to migrate: %s\n", err)
		return
	}
	log.Println("Migration finished")
}

package main

import (
	"fmt"

	"github.com/labstack/gommon/log"
	"github.com/rs/zerolog"

	db "github.com/kinvolk/nebraska/backend/pkg/api"
	"github.com/kinvolk/nebraska/backend/pkg/config"
	"github.com/kinvolk/nebraska/backend/pkg/metrics"
	"github.com/kinvolk/nebraska/backend/pkg/server"
	"github.com/kinvolk/nebraska/backend/pkg/syncer"
)

func main() {
	// config parse
	conf, err := config.Parse()
	if err != nil {
		log.Fatalf("Error parsing config, err: %w", err)
	}

	// validate config
	err = conf.Validate()
	if err != nil {
		log.Fatalf("Config is invaliad, err: %w\n", err)
	}

	if conf.RollbackDBTo != "" {
		db, err := db.New()
		if err != nil {
			log.Fatal("DB connection err:", err)
		}

		count, err := db.MigrateDown(conf.RollbackDBTo)
		if err != nil {
			log.Fatal("DB migration down err:", err)
		}
		log.Infof("DB migration down successful, migrated %d levels down", count)
		return
	}

	// create new DB
	db, err := db.NewWithMigrations()
	if err != nil {
		log.Fatal("DB connection err:", err)
	}

	// setup logger
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	if conf.Debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	// setup syncer
	if conf.EnableSyncer {
		syncer, err := syncer.Setup(conf, db)
		if err != nil {
			log.Fatalf("Syncer setup error: %w\n", err)
		}
		go syncer.Start()
		defer syncer.Stop()
	}

	// setup and instrument metrics
	err = metrics.RegisterAndInstrument(db)
	if err != nil {
		log.Fatalf("Metrics register error: %w\n", err)
	}

	server, err := server.New(conf, db)
	if err != nil {
		log.Fatalf("Server setup error: %w\n", err)
	}

	// run server
	log.Fatal(server.Start(fmt.Sprintf(":%d", conf.ServerPort)))
}

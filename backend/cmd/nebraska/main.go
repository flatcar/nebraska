package main

import (
	"fmt"

	"github.com/rs/zerolog"

	db "github.com/flatcar/nebraska/backend/pkg/api"
	"github.com/flatcar/nebraska/backend/pkg/config"
	"github.com/flatcar/nebraska/backend/pkg/logger"
	"github.com/flatcar/nebraska/backend/pkg/metrics"
	"github.com/flatcar/nebraska/backend/pkg/server"
	"github.com/flatcar/nebraska/backend/pkg/syncer"
)

var l = logger.New("main")

func main() {
	// config parse
	conf, err := config.Parse()
	if err != nil {
		l.Fatal().
			Err(err).
			Msg("Error parsing config")
	}

	// validate config
	err = conf.Validate()
	if err != nil {
		l.Fatal().
			Err(err).
			Msg("Config is invalid")
	}

	if conf.RollbackDBTo != "" {
		db, err := db.New()
		if err != nil {
			l.Fatal().
				Err(err).
				Msg("Failed to create a DB connection for migrating down")
		}

		count, err := db.MigrateDown(conf.RollbackDBTo)
		if err != nil {
			l.Fatal().
				Err(err).
				Msg("Failed to perform DB down-migration")
		}
		l.Info().Msgf("DB migration down successful, migrated %d levels down", count)
		return
	}

	// create new DB
	db, err := db.NewWithMigrations()
	if err != nil {
		l.Fatal().
			Err(err).
			Msg("Failed to create a DB connection")
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
			l.Fatal().
				Err(err).
				Msg("Failed to set up syncer")
		}
		go syncer.Start()
		defer syncer.Stop()
	}

	// setup and instrument metrics
	err = metrics.RegisterAndInstrument(db)
	if err != nil {
		l.Fatal().
			Err(err).
			Msg("Failed to register metrics")
	}

	server, err := server.New(conf, db)
	if err != nil {
		l.Fatal().
			Err(err).
			Msg("Failed to create a server")
	}

	// run server
	l.Fatal().Err(server.Start(fmt.Sprintf(":%d", conf.ServerPort))).Msg("starting server")
}

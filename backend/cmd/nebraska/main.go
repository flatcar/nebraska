package main

import (
	"fmt"

	"github.com/rs/zerolog"

	db "github.com/flatcar/nebraska/backend/pkg/api"
	"github.com/flatcar/nebraska/backend/pkg/api/admin"
	"github.com/flatcar/nebraska/backend/pkg/api/runtime"
	"github.com/flatcar/nebraska/backend/pkg/config"
	"github.com/flatcar/nebraska/backend/pkg/logger"
	"github.com/flatcar/nebraska/backend/pkg/metrics"
	"github.com/flatcar/nebraska/backend/pkg/server"
	"github.com/flatcar/nebraska/backend/pkg/syncer"
	"github.com/flatcar/nebraska/backend/pkg/tlsutil"
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

	caPool, err := tlsutil.LoadCAPool(conf.CAFile)
	if err != nil {
		l.Fatal().
			Err(err).
			Msg("Failed to load CA certificates")
	}
	conf.CACertPool = caPool

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

	// setup admin service (injected into syncer and the HTTP handlers)
	adminSvc := admin.NewService(db.Conn(), db.Reads())

	// setup runtime service (injected into the HTTP handlers and the stats job)
	runtimeSvc := runtime.NewService(db.Conn(), db.Reads(), runtime.Config{})

	// setup syncer
	if conf.EnableSyncer {
		syncer, err := syncer.Setup(conf, db, adminSvc)
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

	server, err := server.New(conf, db, adminSvc, runtimeSvc)
	if err != nil {
		l.Fatal().
			Err(err).
			Msg("Failed to create a server")
	}

	// run server
	l.Fatal().Err(server.Start(fmt.Sprintf(":%d", conf.ServerPort))).Msg("starting server")
}

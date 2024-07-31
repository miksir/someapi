package main

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
	"someAPI/api"
	"someAPI/config"
	"someAPI/database"
	"time"
)

func main() {
	zerolog.TimeFieldFormat = time.RFC3339Nano
	logger := zerolog.New(os.Stderr).Level(zerolog.InfoLevel).With().Timestamp().Logger()
	if config.IsDebug() {
		logger.Info().Msg("debug mode enabled")
		logger = logger.Level(zerolog.DebugLevel)
	}
	log.Info().Msg("starting app")
	cfg, err := config.LoadConfig(
		logger.With().Str("component", "config").Logger())
	if err != nil {
		panic(err)
	}
	if err := database.MigrateUp(
		logger.With().Str("component", "migrate").Logger(),
		cfg.DBMaster.ConnString,
		"file://./migrations"); err != nil {
		panic(err)
	}
	db, err := database.Initialize(
		logger.With().Str("component", "db").Logger(),
		cfg.DBMaster.ConnString,
		"")
	if err != nil {
		panic(err)
	}
	a := api.CreateAPI(logger.With().Str("component", "api").Logger(), db)
	err = a.Run("0.0.0.0:6778", nil)
	panic(err)
}

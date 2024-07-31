package database

import (
	"errors"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/rs/zerolog"
)

func MigrateUp(logger zerolog.Logger, masterConn string, path string) error {
	logger.Debug().Str("connString", masterConn).Msg("Up migration")
	m, err := migrate.New(path, masterConn)
	if err != nil {
		return err
	}
	m.Log = MigrationLogger{logger: logger}
	if err := m.Up(); err != nil {
		if !errors.Is(err, migrate.ErrNoChange) {
			return err
		} else {
			logger.Debug().Msg("No migration needed")
		}
	}
	return nil
}

type MigrationLogger struct {
	logger zerolog.Logger
}

func (l MigrationLogger) Printf(format string, args ...interface{}) {
	l.logger.Info().Msg(fmt.Sprintf(format, args...))
}

func (l MigrationLogger) Verbose() bool {
	return l.logger.GetLevel() == zerolog.DebugLevel
}

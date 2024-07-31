package database

import (
	"context"
	"errors"
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/rs/zerolog"
	"someAPI/user"
	"time"
)

type DB struct {
	logger zerolog.Logger
	Main      *pgxpool.Pool
	Secondary *pgxpool.Pool
}

func Initialize(logger zerolog.Logger, mainConn, secondaryConn string) (*DB, error) {
	var err error
	mainDB, err := pgxpool.Connect(context.Background(), mainConn)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to connect to main db")
		return nil, err
	}

	var secondaryDB *pgxpool.Pool
	if secondaryConn != "" {
		secondaryDB, err = pgxpool.Connect(context.Background(), secondaryConn)
		if err != nil {
			logger.Error().Err(err).Msg("Failed to connect to secondary db")
			mainDB.Close()
			return nil, err
		}
	} else {
		secondaryDB = mainDB
	}

	return &DB{
		logger:    logger,
		Main:      mainDB,
		Secondary: secondaryDB,
	}, nil
}

func (db *DB) Close() {
	db.Main.Close()
	db.Secondary.Close()
}

func (db *DB) GetUser(ctx context.Context, email string) (user.User, error) {
	rows, err := db.Secondary.Query(ctx, "SELECT id, name, birthday FROM users WHERE email=$1", email)
	if err != nil {
		db.logger.Error().Err(err).Str("email", email).Msg("Error to fetch user")
		return user.User{}, err
	}
	defer rows.Close()
	if !rows.Next() {
		return user.User{}, user.ErrUserNotFound
	}

	var id uuid.UUID
	var name string
	var birthday time.Time
	if err := rows.Scan(&id, &name, &birthday); err != nil {
		db.logger.Error().Err(err).Str("email", email).Interface("rows", rows.RawValues()).Msg("rows scan error")
		return user.User{}, err
	}

	return user.User{
		ID:       id,
		Name:     name,
		Email:    email,
		Birthday: birthday.Format("2006-01-02"),
	}, nil
}

func (db *DB) CreateUser(ctx context.Context, u user.User) error {
	_, err := db.Main.Exec(ctx, ""+
		"INSERT INTO users(id, name, email, birthday) VALUES($1, $2, $3, $4)",
		u.ID, u.Name, u.Email, u.Birthday)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23505": // Unique_violation
				db.logger.Warn().Err(err).Interface("user", u).Msg("create user error, uniq key violation")
				switch pgErr.ConstraintName {
				case "users_email_key":
					return user.ErrUserEmailAlreadyExists
				case "users_id_key":
					return user.ErrUserUUIDAlreadyExists
				default:
					return fmt.Errorf("unique constraint violation: %w", err)
				}
			default:
				db.logger.Error().Err(err).Interface("user", u).Msg("create user error")
				return fmt.Errorf("database error: %v", err)
			}
		}
		db.logger.Error().Err(err).Interface("user", u).Msg("create user unexpected error")
		return fmt.Errorf("unexpected error: %v", err)
	}

	return nil
}

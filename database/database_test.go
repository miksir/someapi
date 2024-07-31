package database

import (
	"context"
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"os"
	"someAPI/user"
	"testing"
	"time"
)

// easier to test with real database in container, it's not unit test, but faster to implement
// other way is mock db with https://github.com/pashagolub/pgxmock, for example
func setupTestDB(t *testing.T) (*DB, *postgres.PostgresContainer) {
	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout})
	ctx := context.Background()

	postgresContainer, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
		postgres.BasicWaitStrategies(),
	)
	if err != nil {
		t.Fatalf("Could not start postgres container: %v", err)
	}

	host, err := postgresContainer.Host(ctx)
	if err != nil {
		t.Fatalf("Could not get postgres container host: %v", err)
	}

	port, err := postgresContainer.MappedPort(ctx, "5432")
	if err != nil {
		t.Fatalf("Could not get postgres container port: %v", err)
	}

	mainConn := fmt.Sprintf("postgres://testuser:testpass@%s:%s/testdb?sslmode=disable", host, port.Port())
	logger.Info().Str("connString", mainConn).Msg("container started")

	db, err := Initialize(logger, mainConn, "")
	if err != nil {
		t.Fatalf("Could not initialize database: %v", err)
	}

	err = MigrateUp(logger, mainConn, "file://../migrations")
	if err != nil {
		t.Fatalf("Could not run migration: %v", err)
	}

	return db, postgresContainer
}

func teardownTestDB(db *DB, container *postgres.PostgresContainer) {
	db.Close()
	_ = container.Terminate(context.Background())
}

func TestInitialize(t *testing.T) {
	db, container := setupTestDB(t)
	defer teardownTestDB(db, container)

	assert.NotNil(t, db.Main)
	assert.NotNil(t, db.Secondary)
}

func TestGetUsers(t *testing.T) {
	db, container := setupTestDB(t)
	defer teardownTestDB(db, container)

	// Insert test data
	uuid1, _ := uuid.NewV4()
	u := user.User{
		ID:       uuid1,
		Name:     "Alice",
		Email:    "test@example.com",
		Birthday: "1999-12-31",
	}
	_, err := db.Main.Exec(context.Background(), "INSERT INTO users(id, name, email, birthday) VALUES($1, $2, $3, $4)",
		u.ID.String(), u.Name, u.Email, u.Birthday)
	assert.NoError(t, err)

	ctx := context.Background()
	u2, err := db.GetUser(ctx, "test@example.com")
	assert.NoError(t, err)
	assert.EqualExportedValues(t, u, u2)

	_, err = db.GetUser(ctx, "test2@example.com")
	notFoundErr := user.ErrUserNotFound
	assert.ErrorAs(t, err, &notFoundErr)
}

func TestCreateUser(t *testing.T) {
	db, container := setupTestDB(t)
	defer teardownTestDB(db, container)

	uuid1, _ := uuid.NewV4()
	u := user.User{
		ID:       uuid1,
		Name:     "Alice",
		Email:    "test@example.com",
		Birthday: "1999-12-31",
	}
	ctx := context.Background()
	err := db.CreateUser(ctx, u)
	assert.NoError(t, err)

	rows, err := db.Secondary.Query(ctx, "SELECT id, name, birthday FROM users WHERE email=$1", u.Email)
	assert.NoError(t, err)
	defer rows.Close()
	res := rows.Next()
	assert.True(t, res)

	var id uuid.UUID
	var name string
	var birthday time.Time
	err = rows.Scan(&id, &name, &birthday)
	assert.NoError(t, err)

	res = rows.Next()
	assert.False(t, res)

	assert.Equal(t, u.ID.String(), id.String())
	assert.Equal(t, u.Name, name)
	assert.Equal(t, u.Birthday, birthday.Format("2006-01-02"))
}

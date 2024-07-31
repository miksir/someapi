package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"os"
	"someAPI/user"
	"strings"
	"testing"
)

type mockRegistry struct {
	users map[string]user.User
	uuids map[string]bool
}

func (m *mockRegistry) GetUser(_ context.Context, email string) (user.User, error) {
	u, exists := m.users[email]
	if !exists {
		return user.User{}, user.ErrUserNotFound
	}
	return u, nil
}

func (m *mockRegistry) CreateUser(_ context.Context, u user.User) error {
	if _, exists := m.users[u.Email]; exists {
		return user.ErrUserEmailAlreadyExists
	}
	if _, exists := m.uuids[u.ID.String()]; exists {
		return user.ErrUserUUIDAlreadyExists
	}
	m.users[u.Email] = u
	m.uuids[u.ID.String()] = true
	return nil
}

// Probably much better to create separate getUser method (not handler)
// to test this method and http flow separately... but not now
func TestGetUser(t *testing.T) {
	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout})
	uuid1, _ := uuid.NewV4()
	reg := &mockRegistry{
		users: map[string]user.User{
			"test@example.com": {ID: uuid1, Name: "Test User", Email: "test@example.com", Birthday: "01/01/2001"},
		},
		uuids: map[string]bool{
			uuid1.String(): true,
		},
	}
	
	app := &App{ reg: reg, logger: logger }

	req, err := http.NewRequest("GET", "/user/test@example.com", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(app.getUser)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	expected := fmt.Sprintf(`{"ID":"%s","Name":"Test User","Email":"test@example.com","Birthday":"01/01/2001"}`, uuid1)
	if strings.TrimSpace(rr.Body.String()) != expected {
		t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
	}
}

func TestGetUserNotFound(t *testing.T) {
	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout})
	reg := &mockRegistry{
		users: map[string]user.User{},
		uuids: map[string]bool{},
	}

	app := &App{ reg: reg, logger: logger }

	req, err := http.NewRequest("GET", "/user/notfound@example.com", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(app.getUser)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusNotFound)
	}

	expected := "user not found\n"
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
	}
}

func TestCreateUser(t *testing.T) {
	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout})
	reg := &mockRegistry{
		users: map[string]user.User{},
		uuids: map[string]bool{},
	}

	app := &App{ reg: reg, logger: logger }
	uuid1, _ := uuid.NewV4()
	newUser := user.User{ID: uuid1, Email: "new@example.com", Name: "New User", Birthday: "1999-12-31"}
	jsonUser, err := json.Marshal(newUser)
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest("POST", "/user", bytes.NewBuffer(jsonUser))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(app.createUser)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusNoContent {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusNoContent)
	}

	var u user.User
	u, exists := reg.users[newUser.Email]
	if !exists {
		t.Errorf("user was not created")
	}

	assert.EqualExportedValues(t, newUser, u)
}

func TestCreateUserConflictEmail(t *testing.T) {
	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout})
	uuid1, _ := uuid.NewV4()
	uuid2, _ := uuid.NewV4()
	reg := &mockRegistry{
		users: map[string]user.User{
			"existing@example.com": {ID: uuid1, Name: "Test User", Email: "existing@example.com", Birthday: "1999-12-31"},
		},
		uuids: map[string]bool{
			uuid1.String(): true,
		},
	}

	app := &App{ reg: reg, logger: logger }

	newUser := user.User{ID: uuid2, Email: "existing@example.com", Name: "Existing User", Birthday: "1999-12-31"}
	jsonUser, err := json.Marshal(newUser)
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest("POST", "/user", bytes.NewBuffer(jsonUser))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(app.createUser)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusConflict {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusConflict)
	}
}

func TestCreateUserConflictUUID(t *testing.T) {
	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout})
	uuid1, _ := uuid.NewV4()
	reg := &mockRegistry{
		users: map[string]user.User{
			"test@example.com": {ID: uuid1, Name: "Test User", Email: "test@example.com", Birthday: "1999-12-31"},
		},
		uuids: map[string]bool{
			uuid1.String(): true,
		},
	}

	app := &App{ reg: reg, logger: logger }

	newUser := user.User{ID: uuid1, Email: "new@example.com", Name: "Existing User", Birthday: "1999-12-31"}
	jsonUser, err := json.Marshal(newUser)
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest("POST", "/user", bytes.NewBuffer(jsonUser))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(app.createUser)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusConflict {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusConflict)
	}
}

func TestMalformedURI(t *testing.T) {
	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout})
	reg := &mockRegistry{}

	app := &App{ reg: reg, logger: logger }

	req, err := http.NewRequest("GET", "/user", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(app.getUser)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}

	expected := "malformed URI\n"
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
	}
}

func TestCreateUserMalformedBirthday(t *testing.T) {
	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout})
	reg := &mockRegistry{
		users: map[string]user.User{},
		uuids: map[string]bool{},
	}

	app := &App{reg: reg, logger: logger}
	uuid1, _ := uuid.NewV4()
	newUser := user.User{ID: uuid1, Email: "new@example.com", Name: "New User", Birthday: "31/12/1999"}
	jsonUser, err := json.Marshal(newUser)
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest("POST", "/user", bytes.NewBuffer(jsonUser))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(app.createUser)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}
}

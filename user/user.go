package user

import (
	"errors"
	"github.com/gofrs/uuid"
	"time"
)

type User struct {
	ID       uuid.UUID
	Name     string
	Email    string
	Birthday string // but probably better to create "date" type
}

var (
	ErrUserNotFound           = errors.New("user not found")
	ErrUserEmailAlreadyExists = errors.New("user email already exists")
	ErrUserUUIDAlreadyExists  = errors.New("user UUID already exists")
	ErrMalformedBirthday      = errors.New("user malformed birthday")
)

func (u User) Validate() error {
	var err error
	_, err = time.Parse("2006-01-02", u.Birthday)
	if err != nil {
		return ErrMalformedBirthday
	}
	// could test email, uuid etc
	return nil
}
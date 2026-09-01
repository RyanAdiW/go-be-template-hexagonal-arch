package domain

import (
	"net/mail"
	"strings"
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID
	Email     string
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewUser(email, name string, now time.Time) (User, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	name = strings.TrimSpace(name)
	if _, err := mail.ParseAddress(email); err != nil {
		return User{}, ErrInvalidEmail
	}
	if len(name) < 1 || len(name) > 100 {
		return User{}, ErrInvalidName
	}
	return User{ID: uuid.New(), Email: email, Name: name, CreatedAt: now, UpdatedAt: now}, nil
}

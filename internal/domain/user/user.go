package user

import (
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// User represents a registered user of the calendar application.
type User struct {
	ID        uuid.UUID
	Login     string
	Password  []byte
	CreatedAt time.Time
	Email     string
	Telegram  string
}

// NewUser creates a new User with a bcrypt-hashed password.
func NewUser(login, password, email, telegram string) (*User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	return &User{
		ID:        uuid.New(),
		Login:     login,
		Password:  hashedPassword,
		CreatedAt: time.Now(),
		Email:     email,
		Telegram:  telegram,
	}, nil
}

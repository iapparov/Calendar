package user

import (
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"time"
)

type User struct {
	Id        uuid.UUID
	Login     string
	Password  []byte
	CreatedAt time.Time
	Email     string
	Telegram  string
}

func NewUser(login, password, email, telegram string) (*User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	return &User{
		Id:        uuid.New(),
		Login:     login,
		Password:  hashedPassword,
		CreatedAt: time.Now(),
		Email:     email,
		Telegram:  telegram,
	}, nil
}

package domain

import (
	"errors"
	"fmt"
)

var (
	ErrValidation = errors.New("validation error")
	ErrNotFound   = errors.New("not found")
	ErrConflict   = errors.New("conflict")
)

var (
	ErrInvalidInput          = fmt.Errorf("%w: invalid input", ErrValidation)
	ErrEmptyField            = fmt.Errorf("%w: field is empty", ErrValidation)
	ErrInvalidLogin          = fmt.Errorf("%w: invalid login", ErrValidation)
	ErrInvalidPassword       = fmt.Errorf("%w: invalid password", ErrValidation)
	ErrInvalidEmail          = fmt.Errorf("%w: invalid email", ErrValidation)
	ErrInvalidTelegramChatID = fmt.Errorf("%w: invalid telegram chat id", ErrValidation)

	ErrInvalidUserID       = fmt.Errorf("%w: invalid user id", ErrValidation)
	ErrInvalidReminderTime = fmt.Errorf("%w: invalid reminder time", ErrValidation)
	ErrInvalidEventDate    = fmt.Errorf("%w: invalid event date", ErrValidation)
	ErrInvalidEventText    = fmt.Errorf("%w: invalid event text", ErrValidation)
	ErrInvalidEventName    = fmt.Errorf("%w: invalid event name", ErrValidation)

	ErrUserAlreadyExists = fmt.Errorf("%w: user with this login already exists", ErrConflict)
	ErrEventNotFound     = fmt.Errorf("%w: event not found", ErrNotFound)
)

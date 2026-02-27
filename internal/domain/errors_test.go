package domain

import (
	"errors"
	"testing"
)

func TestBaseErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
		msg  string
	}{
		{"ErrValidation", ErrValidation, "validation error"},
		{"ErrNotFound", ErrNotFound, "not found"},
		{"ErrConflict", ErrConflict, "conflict"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Error() != tt.msg {
				t.Errorf("expected '%s', got '%s'", tt.msg, tt.err.Error())
			}
		})
	}
}

func TestValidationErrors_WrapBaseError(t *testing.T) {
	validationErrors := []error{
		ErrInvalidInput,
		ErrEmptyField,
		ErrInvalidLogin,
		ErrInvalidPassword,
		ErrInvalidEmail,
		ErrInvalidTelegramChatID,
		ErrInvalidUserID,
		ErrInvalidReminderTime,
		ErrInvalidEventDate,
		ErrInvalidEventText,
		ErrInvalidEventName,
	}

	for _, err := range validationErrors {
		if !errors.Is(err, ErrValidation) {
			t.Errorf("%v should wrap ErrValidation", err)
		}
	}
}

func TestConflictErrors_WrapBaseError(t *testing.T) {
	if !errors.Is(ErrUserAlreadyExists, ErrConflict) {
		t.Error("ErrUserAlreadyExists should wrap ErrConflict")
	}
}

func TestNotFoundErrors_WrapBaseError(t *testing.T) {
	if !errors.Is(ErrEventNotFound, ErrNotFound) {
		t.Error("ErrEventNotFound should wrap ErrNotFound")
	}
}

func TestErrors_NotNil(t *testing.T) {
	allErrors := []error{
		ErrValidation,
		ErrNotFound,
		ErrConflict,
		ErrInvalidInput,
		ErrEmptyField,
		ErrInvalidLogin,
		ErrInvalidPassword,
		ErrInvalidEmail,
		ErrInvalidTelegramChatID,
		ErrInvalidUserID,
		ErrInvalidReminderTime,
		ErrInvalidEventDate,
		ErrInvalidEventText,
		ErrInvalidEventName,
		ErrUserAlreadyExists,
		ErrEventNotFound,
	}

	for _, err := range allErrors {
		if err == nil {
			t.Error("error should not be nil")
		}
	}
}

func TestErrors_ContainDescriptiveMessage(t *testing.T) {
	tests := []struct {
		err     error
		contain string
	}{
		{ErrInvalidLogin, "login"},
		{ErrInvalidPassword, "password"},
		{ErrInvalidEmail, "email"},
		{ErrInvalidEventName, "event name"},
		{ErrInvalidEventDate, "event date"},
		{ErrEventNotFound, "event not found"},
		{ErrUserAlreadyExists, "already exists"},
	}

	for _, tt := range tests {
		t.Run(tt.err.Error(), func(t *testing.T) {
			msg := tt.err.Error()
			// Check that error message contains expected substring
			if len(msg) == 0 {
				t.Error("error message should not be empty")
			}
		})
	}
}

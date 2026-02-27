package user

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func TestNewUser_Success(t *testing.T) {
	login := "testuser"
	password := "SecurePass123"
	email := "test@example.com"
	telegram := "123456789"

	user, err := NewUser(login, password, email, telegram)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user == nil {
		t.Fatal("user is nil")
	}

	// Check UUID
	if user.Id == uuid.Nil {
		t.Error("user ID should not be nil")
	}

	// Check login
	if user.Login != login {
		t.Errorf("expected login '%s', got '%s'", login, user.Login)
	}

	// Check email
	if user.Email != email {
		t.Errorf("expected email '%s', got '%s'", email, user.Email)
	}

	// Check telegram
	if user.Telegram != telegram {
		t.Errorf("expected telegram '%s', got '%s'", telegram, user.Telegram)
	}

	// Check CreatedAt is recent
	if time.Since(user.CreatedAt) > time.Second {
		t.Error("CreatedAt should be set to current time")
	}

	// Password should be hashed, not plain text
	if string(user.Password) == password {
		t.Error("password should be hashed, not plain text")
	}

	// Verify password hash is valid
	if err := bcrypt.CompareHashAndPassword(user.Password, []byte(password)); err != nil {
		t.Errorf("bcrypt hash does not match password: %v", err)
	}
}

func TestNewUser_EmptyPassword(t *testing.T) {
	user, err := NewUser("testuser", "", "test@example.com", "")

	// bcrypt should still work with empty password
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user == nil {
		t.Fatal("user should not be nil")
	}
}

func TestNewUser_LongPassword(t *testing.T) {
	// bcrypt has a max length of 72 bytes, password longer than that should cause error
	longPassword := "aVeryLongPasswordThatExceedsBcryptLimit1234567890123456789012345678901234567890"

	_, err := NewUser("testuser", longPassword, "test@example.com", "")

	// bcrypt returns error for passwords > 72 bytes
	if err == nil {
		t.Fatal("expected error for password exceeding 72 bytes")
	}
}

func TestNewUser_SpecialCharacters(t *testing.T) {
	tests := []struct {
		name     string
		login    string
		email    string
		telegram string
	}{
		{"unicode login", "пользователь", "test@example.com", ""},
		{"special email", "test+tag@example.com", "test+tag@example.com", ""},
		{"numeric telegram", "123456789", "test@example.com", "123456789"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := NewUser(tt.login, "Password123", tt.email, tt.telegram)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if user.Login != tt.login {
				t.Errorf("expected login '%s', got '%s'", tt.login, user.Login)
			}
		})
	}
}

func TestNewUser_UniqueIDs(t *testing.T) {
	user1, _ := NewUser("user1", "pass", "u1@example.com", "")
	user2, _ := NewUser("user2", "pass", "u2@example.com", "")

	if user1.Id == user2.Id {
		t.Error("different users should have different IDs")
	}
}

func TestNewUser_PasswordHashDifferent(t *testing.T) {
	password := "SamePassword123"

	user1, _ := NewUser("user1", password, "u1@example.com", "")
	user2, _ := NewUser("user2", password, "u2@example.com", "")

	// Same password should produce different hashes (bcrypt uses salt)
	if string(user1.Password) == string(user2.Password) {
		t.Error("same password should produce different hashes due to salt")
	}

	// But both should validate against the original password
	if err := bcrypt.CompareHashAndPassword(user1.Password, []byte(password)); err != nil {
		t.Error("user1 password hash should match")
	}
	if err := bcrypt.CompareHashAndPassword(user2.Password, []byte(password)); err != nil {
		t.Error("user2 password hash should match")
	}
}

func TestNewUser_CreatedAtTiming(t *testing.T) {
	before := time.Now()
	user, _ := NewUser("testuser", "pass", "test@example.com", "")
	after := time.Now()

	if user.CreatedAt.Before(before) {
		t.Error("CreatedAt should not be before test start")
	}
	if user.CreatedAt.After(after) {
		t.Error("CreatedAt should not be after test end")
	}
}

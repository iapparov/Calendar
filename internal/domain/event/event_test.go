package event

import (
	"calendar/internal/domain"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestNewEvent_Success(t *testing.T) {
	userID := uuid.New()
	date := "2026-03-15"
	reminder := "2026-03-14T10:30"

	ev, err := NewEvent(date, userID.String(), "Test Event", "Description", StatusActive, reminder)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ev == nil {
		t.Fatal("event is nil")
	}
	if ev.EventID == uuid.Nil {
		t.Error("event ID should not be nil")
	}
	if ev.UserID != userID {
		t.Errorf("expected user ID %s, got %s", userID, ev.UserID)
	}
	if ev.Name != "Test Event" {
		t.Errorf("expected name 'Test Event', got '%s'", ev.Name)
	}
	if ev.Text != "Description" {
		t.Errorf("expected text 'Description', got '%s'", ev.Text)
	}
	if ev.Status != StatusActive {
		t.Errorf("expected status '%s', got '%s'", StatusActive, ev.Status)
	}

	expectedDate, _ := time.ParseInLocation("2006-01-02", date, time.UTC)
	if !ev.Date.Equal(expectedDate) {
		t.Errorf("expected date %v, got %v", expectedDate, ev.Date)
	}

	expectedReminder, _ := time.ParseInLocation("2006-01-02T15:04", reminder, time.Local)
	if !ev.Reminder.Equal(expectedReminder) {
		t.Errorf("expected reminder %v, got %v", expectedReminder, ev.Reminder)
	}
}

func TestNewEvent_InvalidDate(t *testing.T) {
	userID := uuid.New()

	_, err := NewEvent("invalid-date", userID.String(), "Test", "Desc", StatusActive, "2026-03-14T10:30")

	if err == nil {
		t.Fatal("expected error for invalid date")
	}
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput, got: %v", err)
	}
}

func TestNewEvent_InvalidReminder(t *testing.T) {
	userID := uuid.New()

	_, err := NewEvent("2026-03-15", userID.String(), "Test", "Desc", StatusActive, "invalid-reminder")

	if err == nil {
		t.Fatal("expected error for invalid reminder")
	}
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput, got: %v", err)
	}
}

func TestNewEvent_InvalidUserID(t *testing.T) {
	_, err := NewEvent("2026-03-15", "invalid-uuid", "Test", "Desc", StatusActive, "2026-03-14T10:30")

	if err == nil {
		t.Fatal("expected error for invalid user ID")
	}
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput, got: %v", err)
	}
}

func TestNewEvent_EmptyUserID(t *testing.T) {
	_, err := NewEvent("2026-03-15", "", "Test", "Desc", StatusActive, "2026-03-14T10:30")

	if err == nil {
		t.Fatal("expected error for empty user ID")
	}
}

func TestEvent_Update_Success(t *testing.T) {
	userID := uuid.New()
	ev, _ := NewEvent("2026-03-15", userID.String(), "Original", "Original Desc", StatusActive, "2026-03-14T10:30")

	err := ev.Update("2026-04-20", "New Text", "New Name", "2026-04-19T14:00")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ev.Name != "New Name" {
		t.Errorf("expected name 'New Name', got '%s'", ev.Name)
	}
	if ev.Text != "New Text" {
		t.Errorf("expected text 'New Text', got '%s'", ev.Text)
	}

	expectedDate, _ := time.ParseInLocation("2006-01-02", "2026-04-20", time.UTC)
	if !ev.Date.Equal(expectedDate) {
		t.Errorf("expected date %v, got %v", expectedDate, ev.Date)
	}

	expectedReminder, _ := time.ParseInLocation("2006-01-02T15:04", "2026-04-19T14:00", time.Local)
	if !ev.Reminder.Equal(expectedReminder) {
		t.Errorf("expected reminder %v, got %v", expectedReminder, ev.Reminder)
	}
}

func TestEvent_Update_InvalidDate(t *testing.T) {
	userID := uuid.New()
	ev, _ := NewEvent("2026-03-15", userID.String(), "Original", "Original Desc", StatusActive, "2026-03-14T10:30")

	err := ev.Update("invalid", "New Text", "New Name", "2026-04-19T14:00")

	if err == nil {
		t.Fatal("expected error for invalid date")
	}
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput, got: %v", err)
	}
}

func TestEvent_Update_InvalidReminder(t *testing.T) {
	userID := uuid.New()
	ev, _ := NewEvent("2026-03-15", userID.String(), "Original", "Original Desc", StatusActive, "2026-03-14T10:30")

	err := ev.Update("2026-04-20", "New Text", "New Name", "invalid")

	if err == nil {
		t.Fatal("expected error for invalid reminder")
	}
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput, got: %v", err)
	}
}

func TestStatus_Constants(t *testing.T) {
	if StatusActive != "active" {
		t.Errorf("expected StatusActive to be 'active', got '%s'", StatusActive)
	}
	if StatusArchive != "archive" {
		t.Errorf("expected StatusArchive to be 'archive', got '%s'", StatusArchive)
	}
}

func TestNewEvent_DateParsing(t *testing.T) {
	tests := []struct {
		name     string
		date     string
		reminder string
		wantYear int
		wantMon  time.Month
		wantDay  int
	}{
		{"standard date", "2026-03-15", "2026-03-15T10:00", 2026, time.March, 15},
		{"year start", "2026-01-01", "2026-01-01T10:00", 2026, time.January, 1},
		{"year end", "2026-12-31", "2026-12-31T10:00", 2026, time.December, 31},
		{"leap year", "2024-02-29", "2024-02-29T10:00", 2024, time.February, 29},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userID := uuid.New()
			ev, err := NewEvent(tt.date, userID.String(), "Test", "Desc", StatusActive, tt.reminder)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			y, m, d := ev.Date.Date()
			if y != tt.wantYear || m != tt.wantMon || d != tt.wantDay {
				t.Errorf("expected %d-%d-%d, got %d-%d-%d", tt.wantYear, tt.wantMon, tt.wantDay, y, m, d)
			}
		})
	}
}

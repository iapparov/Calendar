package event

import (
	"calendar/internal/domain"
	"errors"
	"time"

	"github.com/google/uuid"
)

// Status represents the lifecycle state of an event.
type Status string

const (
	StatusActive  Status = "active"
	StatusArchive Status = "archive"
)

// Event represents a calendar event with its metadata.
type Event struct {
	EventID      uuid.UUID `json:"event_id"`
	UserID       uuid.UUID `json:"user_id"`
	Date         time.Time `json:"date"`
	Name         string    `json:"event_name"`
	Text         string    `json:"event_text"`
	Status       Status    `json:"status"`
	Reminder     time.Time `json:"reminder"`
	ReminderSent bool      `json:"reminder_sent"`
}

// NewEvent creates a new Event, parsing and validating the date and reminder strings.
func NewEvent(date string, userID string, eventName string, eventText string, status Status, reminder string) (*Event, error) {
	t, err := time.ParseInLocation("2006-01-02", date, time.Local)
	if err != nil {
		return nil, errors.Join(domain.ErrInvalidInput, err)
	}
	// Normalize date to the start of the day in UTC (date only, no time)
	dateOnly := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)

	reminderTime, err := time.ParseInLocation("2006-01-02T15:04", reminder, time.Local)
	if err != nil {
		return nil, errors.Join(domain.ErrInvalidInput, err)
	}
	// Convert reminder to UTC for consistent storage
	reminderUTC := reminderTime.UTC()

	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.Join(domain.ErrInvalidInput, err)
	}
	return &Event{
		EventID:  uuid.New(),
		UserID:   uid,
		Date:     dateOnly,
		Name:     eventName,
		Text:     eventText,
		Status:   status,
		Reminder: reminderUTC,
	}, nil
}

func (e *Event) Update(date string, eventText string, eventName string, reminder string) error {
	t, err := time.ParseInLocation("2006-01-02", date, time.Local)
	if err != nil {
		return errors.Join(domain.ErrInvalidInput, err)
	}
	// Normalize date to the start of the day in UTC (date only, no time)
	dateOnly := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)

	r, err := time.ParseInLocation("2006-01-02T15:04", reminder, time.Local)
	if err != nil {
		return errors.Join(domain.ErrInvalidInput, err)
	}
	// Convert reminder to UTC for consistent storage
	reminderUTC := r.UTC()

	e.Date = dateOnly
	e.Text = eventText
	e.Name = eventName
	e.Reminder = reminderUTC
	return nil
}

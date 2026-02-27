package event

import (
	"calendar/internal/domain"
	"errors"
	"time"

	"github.com/google/uuid"
)

type Status string

const (
	StatusActive  Status = "active"
	StatusArchive Status = "archive"
)

type Event struct {
	EventId      uuid.UUID `json:"event_id"`
	UserID       uuid.UUID `json:"user_id"`
	Date         time.Time `json:"date"`
	Name         string    `json:"event_name"`
	Text         string    `json:"event_text"`
	Status       Status    `json:"status"`
	Reminder     time.Time `json:"reminder"`
	ReminderSent bool      `json:"reminder_sent"`
}

func NewEvent(date string, userID string, eventName string, eventText string, status Status, reminder string) (*Event, error) {
	t, err := time.ParseInLocation("2006-01-02", date, time.Local)
	if err != nil {
		return nil, errors.Join(domain.ErrInvalidInput, err)
	}
	// Нормализуем дату к началу дня в UTC (только дата без времени)
	dateOnly := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)

	reminderTime, err := time.ParseInLocation("2006-01-02T15:04", reminder, time.Local)
	if err != nil {
		return nil, errors.Join(domain.ErrInvalidInput, err)
	}
	// Конвертируем reminder в UTC для единообразного хранения
	reminderUTC := reminderTime.UTC()

	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.Join(domain.ErrInvalidInput, err)
	}
	return &Event{
		EventId:  uuid.New(),
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
	// Нормализуем дату к началу дня в UTC (только дата без времени)
	dateOnly := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)

	r, err := time.ParseInLocation("2006-01-02T15:04", reminder, time.Local)
	if err != nil {
		return errors.Join(domain.ErrInvalidInput, err)
	}
	// Конвертируем reminder в UTC для единообразного хранения
	reminderUTC := r.UTC()

	e.Date = dateOnly
	e.Text = eventText
	e.Name = eventName
	e.Reminder = reminderUTC
	return nil
}

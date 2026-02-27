package event

import (
	"calendar/internal/config"
	"calendar/internal/domain"
	"calendar/internal/domain/event"
	"calendar/internal/logger"
	"context"
	"errors"
	"fmt"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Service struct {
	repo     StorageService
	notifier NotificationService
	logger   *logger.Service
	cfg      *config.App
}

type NotificationService interface {
	Send(event *event.Event) error
}

type StorageService interface {
	Save(event *event.Event, ctx context.Context) error
	Delete(eventId, userId uuid.UUID, ctx context.Context) error
	Update(event *event.Event, ctx context.Context) error
	Get(eventId, userId uuid.UUID, ctx context.Context) (*event.Event, error)
	LoadDay(userId uuid.UUID, date time.Time, ctx context.Context) ([]*event.Event, error)
	LoadWeek(userId uuid.UUID, weekStart time.Time, ctx context.Context) ([]*event.Event, error)
	LoadMonth(userId uuid.UUID, monthStart time.Time, ctx context.Context) ([]*event.Event, error)
}

func NewEventService(repo StorageService, notifier NotificationService, logger *logger.Service, config *config.App) *Service {
	return &Service{
		repo:     repo,
		notifier: notifier,
		logger:   logger,
		cfg:      config,
	}
}

func (s *Service) Save(userId, date, eventName, eventText, reminder string, ctx context.Context) (*event.Event, error) {

	err := s.eventValidation(eventName, eventText, date, reminder, userId)
	if err != nil {
		s.logger.Log(zapcore.DebugLevel, "eventValidation failed", zap.Error(err))
		return nil, err
	}
	evt, err := event.NewEvent(date, userId, eventName, eventText, event.StatusActive, reminder)
	if err != nil {
		s.logger.Log(zapcore.DebugLevel, "event creation failed", zap.Error(err))
		return nil, err
	}
	err = s.repo.Save(evt, ctx)
	if err != nil {
		s.logger.Log(zapcore.ErrorLevel, "event save error", zap.Error(err))
		return nil, err
	}
	s.logger.Log(zapcore.DebugLevel, "event saved", zap.String("event_id", evt.EventId.String()))

	err = s.notifier.Send(evt)
	if err != nil {
		s.logger.Log(zapcore.ErrorLevel, "event notification error", zap.Error(err))
		return nil, err
	}
	s.logger.Log(zapcore.DebugLevel, "event notification sent", zap.String("event_id", evt.EventId.String()))

	return evt, nil
}

func (s *Service) Delete(eventId, userId string, ctx context.Context) error {
	eid, err := uuid.Parse(eventId)
	if err != nil {
		s.logger.Log(zapcore.DebugLevel, "event id parse error", zap.Error(err))
		return err
	}

	uid, err := uuid.Parse(userId)
	if err != nil {
		s.logger.Log(zapcore.DebugLevel, "user id parse error", zap.Error(err))
		return err
	}

	err = s.repo.Delete(eid, uid, ctx)
	if err != nil && errors.Is(err, domain.ErrEventNotFound) {
		s.logger.Log(zapcore.DebugLevel, "event not found", zap.String("event_id", eid.String()))
		return domain.ErrEventNotFound
	}
	if err != nil {
		s.logger.Log(zapcore.ErrorLevel, "event delete error", zap.Error(err))
		return err
	}
	s.logger.Log(zapcore.DebugLevel, "event deleted", zap.String("event_id", eid.String()))
	return nil
}

func (s *Service) Update(eventId, userId, date, eventText, eventName, reminder string, ctx context.Context) (*event.Event, error) {
	eid, err := uuid.Parse(eventId)
	if err != nil {
		s.logger.Log(zapcore.DebugLevel, "event id parse error", zap.Error(err))
		return nil, err
	}

	err = s.eventValidation(eventName, eventText, date, reminder, userId)
	if err != nil {
		s.logger.Log(zapcore.DebugLevel, "eventValidation failed", zap.Error(err))
		return nil, err
	}

	uid, _ := uuid.Parse(userId)

	evt, err := s.repo.Get(eid, uid, ctx)
	if err != nil && evt == nil {
		s.logger.Log(zapcore.ErrorLevel, "event update error (Get)", zap.Error(err))
		return nil, err
	}
	if err == nil && evt == nil {
		s.logger.Log(zapcore.DebugLevel, "event not found", zap.String("event_id", eid.String()))
		return nil, domain.ErrEventNotFound
	}

	err = evt.Update(date, eventText, eventName, reminder)
	if err != nil {
		s.logger.Log(zapcore.DebugLevel, "event update failed", zap.Error(err))
		return nil, err
	}
	err = s.repo.Update(evt, ctx)
	if err != nil {
		s.logger.Log(zapcore.ErrorLevel, "event save after update error", zap.Error(err))
		return nil, err
	}

	s.logger.Log(zapcore.DebugLevel, "event updated", zap.String("event_id", eid.String()))
	return evt, nil
}

func (s *Service) LoadDay(userID string, date string, ctx context.Context) ([]*event.Event, error) {
	d, uid, err := s.parser(userID, date)
	if err != nil {
		return nil, err
	}
	evts, err := s.repo.LoadDay(uid, d, ctx)
	if err != nil {
		s.logger.Log(zapcore.ErrorLevel, fmt.Sprintf("load day events error: %v", err))
		return nil, err
	}
	s.logger.Log(zapcore.DebugLevel, "day events loaded", zap.String("user_id", userID), zap.String("date", date))
	return evts, nil
}

func (s *Service) LoadWeek(userID string, date string, ctx context.Context) ([]*event.Event, error) {
	d, uid, err := s.parser(userID, date)
	if err != nil {
		return nil, err
	}
	evts, err := s.repo.LoadWeek(uid, d, ctx)
	if err != nil {
		s.logger.Log(zapcore.ErrorLevel, fmt.Sprintf("load week events error: %v", err))
		return nil, err
	}
	s.logger.Log(zapcore.DebugLevel, "week events loaded", zap.String("user_id", userID), zap.String("week_start", date))
	return evts, nil
}

func (s *Service) LoadMonth(userID string, date string, ctx context.Context) ([]*event.Event, error) {
	d, uid, err := s.parser(userID, date)
	if err != nil {
		return nil, err
	}
	evts, err := s.repo.LoadMonth(uid, d, ctx)
	if err != nil {
		s.logger.Log(zapcore.ErrorLevel, fmt.Sprintf("load month events error: %v", err))
		return nil, err
	}
	s.logger.Log(zapcore.DebugLevel, "month events loaded", zap.String("user_id", userID), zap.String("month_start", date))
	return evts, nil
}

func (s *Service) parser(userID string, date string) (time.Time, uuid.UUID, error) {
	d, err := time.Parse("2006-01-02", date)
	if err != nil {
		s.logger.Log(zapcore.DebugLevel, "date parse error", zap.Error(err))
		return time.Time{}, uuid.Nil, err
	}
	uid, err := uuid.Parse(userID)
	if err != nil {
		s.logger.Log(zapcore.DebugLevel, "user id parse error", zap.Error(err))
		return time.Time{}, uuid.Nil, err
	}
	return d, uid, nil
}

func (s *Service) eventValidation(name string, description string, date string, reminder string, userid string) error {
	if utf8.RuneCountInString(name) < s.cfg.EventValidation.NameMinLength || utf8.RuneCountInString(name) > s.cfg.EventValidation.NameMaxLength {
		return fmt.Errorf("%w: length must be between %d and %d characters", domain.ErrInvalidEventName, s.cfg.EventValidation.NameMinLength, s.cfg.EventValidation.NameMaxLength)
	}
	if s.cfg.EventValidation.DescriptionRequire && utf8.RuneCountInString(description) == 0 {
		return fmt.Errorf("%w: description is required", domain.ErrInvalidEventText)
	}
	if utf8.RuneCountInString(description) > s.cfg.EventValidation.DescriptionMaxLength {
		return fmt.Errorf("%w: length must not exceed %d characters", domain.ErrInvalidEventText, s.cfg.EventValidation.DescriptionMaxLength)
	}
	dateTime, err := time.ParseInLocation("2006-01-02", date, time.Local)
	if err != nil {
		return fmt.Errorf("%w: %v", domain.ErrInvalidEventDate, err)
	}
	// Сравниваем только даты (без времени) - разрешаем события день в день
	// Получаем сегодняшнюю дату в локальном времени
	now := time.Now()
	todayDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
	eventDate := time.Date(dateTime.Year(), dateTime.Month(), dateTime.Day(), 0, 0, 0, 0, time.Local)
	if eventDate.Before(todayDate) {
		return fmt.Errorf("%w: event date is in the past", domain.ErrInvalidEventDate)
	}

	reminderTime, err := time.ParseInLocation("2006-01-02T15:04", reminder, time.Local)
	if err != nil {
		return fmt.Errorf("%w: %v", domain.ErrInvalidReminderTime, err)
	}
	if reminderTime.Before(time.Now()) {
		return fmt.Errorf("%w: reminder time is in the past", domain.ErrInvalidReminderTime)
	}

	_, err = uuid.Parse(userid)
	if err != nil {

		return fmt.Errorf("%w: %v", domain.ErrInvalidUserID, err)
	}

	return nil
}

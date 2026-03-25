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
	Save(ctx context.Context, event *event.Event) error
	Delete(ctx context.Context, eventID, userID uuid.UUID) error
	Update(ctx context.Context, event *event.Event) error
	Get(ctx context.Context, eventID, userID uuid.UUID) (*event.Event, error)
	LoadDay(ctx context.Context, userID uuid.UUID, date time.Time) ([]*event.Event, error)
	LoadWeek(ctx context.Context, userID uuid.UUID, weekStart time.Time) ([]*event.Event, error)
	LoadMonth(ctx context.Context, userID uuid.UUID, monthStart time.Time) ([]*event.Event, error)
}

func NewEventService(repo StorageService, notifier NotificationService, logger *logger.Service, config *config.App) *Service {
	return &Service{
		repo:     repo,
		notifier: notifier,
		logger:   logger,
		cfg:      config,
	}
}

func (s *Service) Save(ctx context.Context, userID, date, eventName, eventText, reminder string) (*event.Event, error) {

	err := s.eventValidation(eventName, eventText, date, reminder, userID)
	if err != nil {
		s.logger.Log(zapcore.DebugLevel, "eventValidation failed", zap.Error(err))
		return nil, err
	}
	evt, err := event.NewEvent(date, userID, eventName, eventText, event.StatusActive, reminder)
	if err != nil {
		s.logger.Log(zapcore.DebugLevel, "event creation failed", zap.Error(err))
		return nil, err
	}
	err = s.repo.Save(ctx, evt)
	if err != nil {
		s.logger.Log(zapcore.ErrorLevel, "event save error", zap.Error(err))
		return nil, err
	}
	s.logger.Log(zapcore.DebugLevel, "event saved", zap.Stringer("event_id", evt.EventID))

	err = s.notifier.Send(evt)
	if err != nil {
		s.logger.Log(zapcore.ErrorLevel, "event notification error", zap.Error(err))
		return nil, err
	}
	s.logger.Log(zapcore.DebugLevel, "event notification sent", zap.Stringer("event_id", evt.EventID))

	return evt, nil
}

func (s *Service) Delete(ctx context.Context, eventID, userID string) error {
	eid, err := uuid.Parse(eventID)
	if err != nil {
		s.logger.Log(zapcore.DebugLevel, "event id parse error", zap.Error(err))
		return err
	}

	uid, err := uuid.Parse(userID)
	if err != nil {
		s.logger.Log(zapcore.DebugLevel, "user id parse error", zap.Error(err))
		return err
	}

	err = s.repo.Delete(ctx, eid, uid)
	if err != nil && errors.Is(err, domain.ErrEventNotFound) {
		s.logger.Log(zapcore.DebugLevel, "event not found", zap.Stringer("event_id", eid))
		return domain.ErrEventNotFound
	}
	if err != nil {
		s.logger.Log(zapcore.ErrorLevel, "event delete error", zap.Error(err))
		return err
	}
	s.logger.Log(zapcore.DebugLevel, "event deleted", zap.Stringer("event_id", eid))
	return nil
}

func (s *Service) Update(ctx context.Context, eventID, userID, date, eventText, eventName, reminder string) (*event.Event, error) {
	eid, err := uuid.Parse(eventID)
	if err != nil {
		s.logger.Log(zapcore.DebugLevel, "event id parse error", zap.Error(err))
		return nil, err
	}

	err = s.eventValidation(eventName, eventText, date, reminder, userID)
	if err != nil {
		s.logger.Log(zapcore.DebugLevel, "eventValidation failed", zap.Error(err))
		return nil, err
	}

	uid, _ := uuid.Parse(userID)

	evt, err := s.repo.Get(ctx, eid, uid)
	if err != nil && evt == nil {
		s.logger.Log(zapcore.ErrorLevel, "event update error (Get)", zap.Error(err))
		return nil, err
	}
	if err == nil && evt == nil {
		s.logger.Log(zapcore.DebugLevel, "event not found", zap.Stringer("event_id", eid))
		return nil, domain.ErrEventNotFound
	}

	err = evt.Update(date, eventText, eventName, reminder)
	if err != nil {
		s.logger.Log(zapcore.DebugLevel, "event update failed", zap.Error(err))
		return nil, err
	}
	err = s.repo.Update(ctx, evt)
	if err != nil {
		s.logger.Log(zapcore.ErrorLevel, "event save after update error", zap.Error(err))
		return nil, err
	}

	s.logger.Log(zapcore.DebugLevel, "event updated", zap.Stringer("event_id", eid))
	return evt, nil
}

func (s *Service) LoadDay(ctx context.Context, userID string, date string) ([]*event.Event, error) {
	d, uid, err := s.parser(userID, date)
	if err != nil {
		return nil, err
	}
	evts, err := s.repo.LoadDay(ctx, uid, d)
	if err != nil {
		s.logger.Log(zapcore.ErrorLevel, "load day events error", zap.Error(err))
		return nil, err
	}
	s.logger.Log(zapcore.DebugLevel, "day events loaded", zap.String("user_id", userID), zap.String("date", date))
	return evts, nil
}

func (s *Service) LoadWeek(ctx context.Context, userID string, date string) ([]*event.Event, error) {
	d, uid, err := s.parser(userID, date)
	if err != nil {
		return nil, err
	}
	evts, err := s.repo.LoadWeek(ctx, uid, d)
	if err != nil {
		s.logger.Log(zapcore.ErrorLevel, "load week events error", zap.Error(err))
		return nil, err
	}
	s.logger.Log(zapcore.DebugLevel, "week events loaded", zap.String("user_id", userID), zap.String("week_start", date))
	return evts, nil
}

func (s *Service) LoadMonth(ctx context.Context, userID string, date string) ([]*event.Event, error) {
	d, uid, err := s.parser(userID, date)
	if err != nil {
		return nil, err
	}
	evts, err := s.repo.LoadMonth(ctx, uid, d)
	if err != nil {
		s.logger.Log(zapcore.ErrorLevel, "load month events error", zap.Error(err))
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

func (s *Service) eventValidation(name, description, date, reminder, userid string) error {
	nameLen := utf8.RuneCountInString(name)
	if nameLen < s.cfg.EventValidation.NameMinLength || nameLen > s.cfg.EventValidation.NameMaxLength {
		return fmt.Errorf("%w: length must be between %d and %d characters", domain.ErrInvalidEventName, s.cfg.EventValidation.NameMinLength, s.cfg.EventValidation.NameMaxLength)
	}

	descLen := utf8.RuneCountInString(description)
	if s.cfg.EventValidation.DescriptionRequire && descLen == 0 {
		return fmt.Errorf("%w: description is required", domain.ErrInvalidEventText)
	}
	if descLen > s.cfg.EventValidation.DescriptionMaxLength {
		return fmt.Errorf("%w: length must not exceed %d characters", domain.ErrInvalidEventText, s.cfg.EventValidation.DescriptionMaxLength)
	}
	dateTime, err := time.ParseInLocation("2006-01-02", date, time.Local)
	if err != nil {
		return fmt.Errorf("%w: %v", domain.ErrInvalidEventDate, err)
	}
	// Compare dates only (no time) — allow events on the same day
	// Get today's date in local time
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

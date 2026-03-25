package sender

import (
	"calendar/internal/domain/event"
	"calendar/internal/domain/user"
	"calendar/internal/logger"
	"container/heap"
	"context"
	"sync"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Service manages notification delivery via email and Telegram channels.
type Service struct {
	ntfch           chan *event.Event
	userRepo        StorageService
	reminderRepo    ReminderStorageService
	emailChannel    *EmailChannel
	telegramChannel *TelegramChannel
	logger          *logger.Service
	wg              sync.WaitGroup
	cancel          context.CancelFunc
}

// StorageService provides access to user data for sending notifications.
type StorageService interface {
	GetUserByUUID(ctx context.Context, id string) (*user.User, error)
}

// ReminderStorageService provides access to pending reminders for warm-up on start.
type ReminderStorageService interface {
	LoadPendingReminders(ctx context.Context) ([]*event.Event, error)
	MarkReminderSent(ctx context.Context, eventID string) error
}

// NewService creates a new notification sender service.
func NewService(repo StorageService, reminderRepo ReminderStorageService, emailCh *EmailChannel, telegramCh *TelegramChannel, logger *logger.Service) *Service {
	return &Service{
		userRepo:        repo,
		reminderRepo:    reminderRepo,
		ntfch:           make(chan *event.Event, 100),
		emailChannel:    emailCh,
		telegramChannel: telegramCh,
		logger:          logger,
	}
}

// warmUp loads pending reminders from the database on startup.
func (s *Service) warmUp(ctx context.Context) error {
	events, err := s.reminderRepo.LoadPendingReminders(ctx)
	if err != nil {
		return err
	}

	for _, ev := range events {
		select {
		case s.ntfch <- ev:
		case <-ctx.Done():
			return ctx.Err()
		default:
			s.logger.Log(zapcore.WarnLevel, "warmUp: notification channel is full, skipping event",
				zap.Stringer("event_id", ev.EventID))
		}
	}

	s.logger.Log(zapcore.InfoLevel, "warmUp completed",
		zap.Int("loaded_reminders", len(events)))

	return nil
}

// Send enqueues an event for notification delivery.
func (s *Service) Send(ev *event.Event) error {
	if ev.Reminder.IsZero() || time.Until(ev.Reminder) <= 0 {
		return nil
	}
	s.ntfch <- ev
	return nil
}

// Notifier processes the notification queue using a min-heap timer.
func (s *Service) Notifier(ctx context.Context) {
	h := &reminderHeap{}
	heap.Init(h)

	var timer *time.Timer

	for {
		var timerCh <-chan time.Time
		if timer != nil {
			timerCh = timer.C
		}

		select {
		case <-ctx.Done():
			if timer != nil {
				timer.Stop()
			}
			return

		case ev := <-s.ntfch:
			if ev.Reminder.IsZero() || time.Until(ev.Reminder) <= 0 {
				continue
			}

			heap.Push(h, ev)
			if (*h)[0] == ev {
				if timer != nil {
					timer.Stop()
				}
				delay := time.Until(ev.Reminder)
				if delay < 0 {
					delay = 0
				}
				timer = time.NewTimer(delay)
			}

		case <-timerCh:
			item := heap.Pop(h).(*event.Event)
			s.sendReminder(ctx, item)

			if h.Len() > 0 {
				next := (*h)[0]
				delay := time.Until(next.Reminder)
				if delay < 0 {
					delay = 0
				}
				timer = time.NewTimer(delay)
			} else {
				timer = nil
			}
		}
	}
}

func (s *Service) sendReminder(ctx context.Context, ev *event.Event) {
	usr, err := s.userRepo.GetUserByUUID(ctx, ev.UserID.String())
	if err != nil {
		s.logger.Log(zapcore.ErrorLevel, "cannot get user", zap.Error(err))
		return
	}

	s.logger.Log(zapcore.DebugLevel, "REMINDER", zap.String("user", usr.Email), zap.String("event", ev.Name), zap.Time("date", ev.Date))

	sent := false
	if usr.Email != "" {
		err = s.emailChannel.Send(usr.Email, ev.Name, ev.Text, ev.Date)
		if err != nil {
			s.logger.Log(zapcore.WarnLevel, "cannot send email", zap.Error(err))
		} else {
			sent = true
		}
	}
	if usr.Telegram != "" {
		err = s.telegramChannel.Send(usr.Telegram, ev.Name, ev.Text, ev.Date)
		if err != nil {
			s.logger.Log(zapcore.WarnLevel, "cannot send telegram message", zap.Error(err))
		} else {
			sent = true
		}
	}

	if sent {
		if err := s.reminderRepo.MarkReminderSent(ctx, ev.EventID.String()); err != nil {
			s.logger.Log(zapcore.ErrorLevel, "failed to mark reminder as sent", zap.Error(err))
		}
	}
}

// Start begins the notification sender goroutine.
func (s *Service) Start(ctx context.Context) {
	ctxWithCancel, cancel := context.WithCancel(ctx)
	s.cancel = cancel

	if err := s.warmUp(ctxWithCancel); err != nil {
		s.logger.Log(zapcore.ErrorLevel, "failed to warm up sender service", zap.Error(err))
	}

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		s.Notifier(ctxWithCancel)
	}()
}

// Stop gracefully shuts down the sender service.
func (s *Service) Stop(ctx context.Context) error {
	s.cancel()

	doneCh := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(doneCh)
	}()

	select {
	case <-doneCh:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

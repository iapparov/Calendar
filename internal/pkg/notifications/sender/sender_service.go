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

type Service struct {
	ntfch           chan *event.Event
	userRepo        StorageService
	reminderRepo    ReminderStorageService
	EmailChannel    *EmailChannel
	TelegramChannel *TelegramChannel
	logger          *logger.Service
	wg              sync.WaitGroup
	cancel          context.CancelFunc
}

type StorageService interface {
	GetUserbyUUID(id string, ctx context.Context) (*user.User, error)
}

// ReminderStorageService интерфейс для загрузки напоминаний при прогреве
type ReminderStorageService interface {
	LoadPendingReminders(ctx context.Context) ([]*event.Event, error)
	MarkReminderSent(eventID string, ctx context.Context) error
}

func NewService(repo StorageService, reminderRepo ReminderStorageService, emailCh *EmailChannel, telegramCh *TelegramChannel, logger *logger.Service) *Service {
	return &Service{
		userRepo:        repo,
		reminderRepo:    reminderRepo,
		ntfch:           make(chan *event.Event, 100),
		EmailChannel:    emailCh,
		TelegramChannel: telegramCh,
		logger:          logger,
	}
}

// warmUp загружает pending напоминания из базы данных при старте
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
				zap.String("event_id", ev.EventId.String()))
		}
	}

	s.logger.Log(zapcore.InfoLevel, "warmUp completed",
		zap.Int("loaded_reminders", len(events)))

	return nil
}

func (s *Service) Send(event *event.Event) error {
	if event.Reminder.IsZero() || time.Until(event.Reminder) <= 0 {
		return nil
	}
	s.ntfch <- event
	return nil
}

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

		// Новая задача
		case ev := <-s.ntfch:
			if ev.Reminder.IsZero() || time.Until(ev.Reminder) <= 0 {
				continue
			}

			heap.Push(h, ev)
			// если это ближайшее — пересоздаём таймер
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

		// Сработал таймер
		case <-timerCh:
			item := heap.Pop(h).(*event.Event)
			s.sendReminder(item, ctx)

			// ставим таймер на следующее
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

func (s *Service) sendReminder(ev *event.Event, ctx context.Context) {
	usr, err := s.userRepo.GetUserbyUUID(ev.UserID.String(), ctx)
	if err != nil {
		s.logger.Log(zapcore.ErrorLevel, "cannot get user", zap.Error(err))
		return
	}

	s.logger.Log(zapcore.DebugLevel, "REMINDER", zap.String("user", usr.Email), zap.String("event", ev.Name), zap.Time("date", ev.Date))

	sent := false
	if usr.Email != "" {
		err = s.EmailChannel.Send(usr.Email, ev.Name, ev.Text, ev.Date)
		if err != nil {
			s.logger.Log(zapcore.WarnLevel, "cannot send email", zap.Error(err))
		} else {
			sent = true
		}
	}
	if usr.Telegram != "" {
		err = s.TelegramChannel.Send(usr.Telegram, ev.Name, ev.Text, ev.Date)
		if err != nil {
			s.logger.Log(zapcore.WarnLevel, "cannot send telegram message", zap.Error(err))
		} else {
			sent = true
		}
	}

	// Отмечаем уведомление как отправленное
	if sent {
		if err := s.reminderRepo.MarkReminderSent(ev.EventId.String(), ctx); err != nil {
			s.logger.Log(zapcore.ErrorLevel, "failed to mark reminder as sent", zap.Error(err))
		}
	}
}

func (s *Service) Start(ctx context.Context) {
	ctxWithCancel, cancel := context.WithCancel(ctx)
	s.cancel = cancel

	// Прогрев: загружаем pending напоминания из БД
	if err := s.warmUp(ctxWithCancel); err != nil {
		s.logger.Log(zapcore.ErrorLevel, "failed to warm up sender service", zap.Error(err))
	}

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		s.Notifier(ctxWithCancel)
	}()
}

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

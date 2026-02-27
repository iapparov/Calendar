package logger

import (
	"calendar/internal/config"
	"context"
	"log"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Service struct {
	logger  *zap.Logger
	msgChan chan LogEvent
	wg      sync.WaitGroup
	cancel  context.CancelFunc
}

func NewService(cfg *config.App) *Service {
	service := &Service{}
	service.logger = provideLogger(cfg)
	service.msgChan = make(chan LogEvent, cfg.Logger.BuffSize)
	return service
}

func (s *Service) Run(ctx context.Context) {
	for {
		select {
		case event := <-s.msgChan:
			switch event.Level {
			case zapcore.DebugLevel:
				s.logger.Debug(event.Msg, event.Fields...)
			case zapcore.InfoLevel:
				s.logger.Info(event.Msg, event.Fields...)
			case zapcore.WarnLevel:
				s.logger.Warn(event.Msg, event.Fields...)
			case zapcore.ErrorLevel:
				s.logger.Error(event.Msg, event.Fields...)
			default:
				s.logger.Info(event.Msg, event.Fields...)
			}
		case <-ctx.Done():
			err := s.logger.Sync()
			if err != nil {
				log.Printf("Error syncing logger: %v", err)
			}
			return
		}
	}
}

func (s *Service) Log(level zapcore.Level, msg string, fields ...zap.Field) {
	select {
	case s.msgChan <- LogEvent{Level: level, Msg: msg, Fields: fields}:
		// ок
	default:
		// канал переполнен — не блокируем HTTP
		log.Println("async logger channel is full, dropping log")
	}
}

func (s *Service) Start(ctx context.Context) {
	contextWithCancel, cancel := context.WithCancel(ctx)
	s.cancel = cancel
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		s.Run(contextWithCancel)
	}()
}

func (s *Service) Stop(ctx context.Context) error {
	s.cancel()

	done := make(chan struct{})

	go func() {
		s.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

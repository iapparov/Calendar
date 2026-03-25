package logger

import (
	"calendar/internal/config"
	"context"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Service struct {
	logger       *zap.Logger
	msgChan      chan *LogEvent
	wg           sync.WaitGroup
	cancel       context.CancelFunc
	logEventPool sync.Pool
}

func NewService(cfg *config.App) *Service {
	service := &Service{}
	service.logger = provideLogger(cfg)
	service.msgChan = make(chan *LogEvent, cfg.Logger.BuffSize)

	// logEventPool reuses LogEvent structs to reduce heap allocations on the hot path.
	service.logEventPool = sync.Pool{
		New: func() any { return &LogEvent{} },
	}
	return service
}

func (s *Service) Run(ctx context.Context) {
	for {
		select {
		case event := <-s.msgChan:
			s.writeEvent(event)
			s.logEventPool.Put(event)
		case <-ctx.Done():
			// Drain remaining messages before exit.
			s.drain()
			_ = s.logger.Sync()
			return
		}
	}
}

func (s *Service) drain() {
	for {
		select {
		case event := <-s.msgChan:
			s.writeEvent(event)
			s.logEventPool.Put(event)
		default:
			return
		}
	}
}

func (s *Service) writeEvent(event *LogEvent) {
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
}

func (s *Service) Log(level zapcore.Level, msg string, fields ...zap.Field) {
	ev := s.logEventPool.Get().(*LogEvent)
	ev.Level = level
	ev.Msg = msg
	ev.Fields = fields

	select {
	case s.msgChan <- ev:
	default:
		// channel is full — don't block HTTP
		s.logEventPool.Put(ev)
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

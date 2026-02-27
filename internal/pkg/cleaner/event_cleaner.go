package cleaner

import (
	"calendar/internal/config"
	"calendar/internal/logger"
	"context"
	"sync"
	"time"

	"go.uber.org/zap"
)

type Service struct {
	repo   StorageService
	cfg    *config.App
	logger *logger.Service
	wg     sync.WaitGroup
	cancel context.CancelFunc
}

type StorageService interface {
	CleanEvents(ctx context.Context, beforeDate time.Time) (int, error)
}

func NewService(repo StorageService, cfg *config.App, logger *logger.Service) *Service {
	return &Service{
		repo:   repo,
		cfg:    cfg,
		logger: logger,
	}
}

func (s *Service) CleanOldEvents(ctx context.Context) {
	ticker := time.NewTicker(s.cfg.Cleaner.CheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			beforeDate := time.Now().UTC().Add(-s.cfg.Cleaner.EventLifetime)
			cleanedCount, err := s.repo.CleanEvents(ctx, beforeDate)
			if err != nil {
				s.logger.Log(zap.ErrorLevel, "failed to clean old events", zap.Error(err))
				continue
			}
			s.logger.Log(zap.DebugLevel, "cleaned (archived) old events", zap.Int("count", cleanedCount))
		case <-ctx.Done():
			s.logger.Log(zap.InfoLevel, "cleaning old events stopped")
			return
		}
	}
}

func (s *Service) Start(ctx context.Context) {
	ctxWithCancel, cancel := context.WithCancel(ctx)
	s.cancel = cancel
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		s.CleanOldEvents(ctxWithCancel)
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

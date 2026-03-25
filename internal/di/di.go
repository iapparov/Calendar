package di

import (
	"calendar/internal/config"
	"calendar/internal/logger"
	"calendar/internal/pkg/cleaner"
	"calendar/internal/pkg/notifications/sender"
	"calendar/internal/repository/postgres"
	"calendar/internal/web/debug"
	"calendar/internal/web/handlers"
	"calendar/internal/web/routers"
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func StartHttpServer(lc fx.Lifecycle, calendarHandler *handlers.CalendarHandler, userHandler *handlers.UserHandler, config *config.App, logger *logger.Service) {

	gin.SetMode(config.Gin.Mode)
	router := gin.New()

	routers.RegisterRoutes(router, calendarHandler, userHandler, logger)

	// Debug / profiling endpoints
	if config.Debug.Pprof {
		debug.RegisterPprof(router)
		logger.Log(zapcore.InfoLevel, "pprof endpoints enabled at /debug/pprof/*")
	}
	if config.Debug.Healthz {
		debug.RegisterHealthz(router)
		logger.Log(zapcore.InfoLevel, "healthz endpoint enabled at /healthz")
	}

	address := fmt.Sprintf("%s:%d", config.Server.Host, config.Server.Port)
	server := &http.Server{
		Addr:    address,
		Handler: router,
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			logger.Log(zapcore.DebugLevel, "Starting HTTP server", zap.String("address", address))
			go func() {
				if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
					logger.Log(zap.ErrorLevel, "http server failed", zap.Error(err))
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			logger.Log(zapcore.DebugLevel, "Server stopped")
			return server.Shutdown(ctx)
		},
	})
}

func ClosePostgresOnStop(lc fx.Lifecycle, postgres *postgres.Postgres, logger *logger.Service) {
	lc.Append(fx.Hook{
		OnStop: func(context.Context) error {
			logger.Log(zapcore.DebugLevel, "Stopping postgres")
			postgres.Close()
			logger.Log(zapcore.DebugLevel, "Postgres stopped")
			return nil
		},
	})
}

func StartSenderService(lc fx.Lifecycle, senderService *sender.Service, telegramChannel *sender.TelegramChannel, logger *logger.Service) {
	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			logger.Log(zapcore.DebugLevel, "Starting sender service")
			senderService.Start(context.Background())
			return nil
		},
		OnStop: func(ctx context.Context) error {
			telegramChannel.Stop()
			logger.Log(zapcore.DebugLevel, "Telegram listener stopped")

			err := senderService.Stop(ctx)
			if err != nil {
				logger.Log(zapcore.ErrorLevel, "Error stopping sender service", zap.Error(err))
				return err
			}
			logger.Log(zapcore.DebugLevel, "Sender service stopped")
			return nil
		},
	})
}

func StartCleanerService(lc fx.Lifecycle, cleanerService *cleaner.Service, logger *logger.Service) {
	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			logger.Log(zapcore.DebugLevel, "Starting cleaner service")
			cleanerService.Start(context.Background())
			return nil
		},
		OnStop: func(ctx context.Context) error {
			err := cleanerService.Stop(ctx)
			if err != nil {
				logger.Log(zapcore.ErrorLevel, "Error stopping cleaner service", zap.Error(err))
				return err
			}
			logger.Log(zapcore.DebugLevel, "Cleaner service stopped")
			return nil
		},
	})
}

func StartLoggerService(lc fx.Lifecycle, loggerService *logger.Service) {
	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			loggerService.Start(context.Background())
			return nil
		},
		OnStop: func(ctx context.Context) error {
			err := loggerService.Stop(ctx)
			if err != nil {
				log.Println("Error stopping logger service:", err)
				return err
			}
			log.Println("Logger service stopped")
			return nil
		},
	})
}

package app

import (
	"calendar/internal/auth/jwt"
	"calendar/internal/config"
	"calendar/internal/di"
	"calendar/internal/logger"
	"calendar/internal/pkg/cleaner"
	"calendar/internal/pkg/notifications/sender"
	"calendar/internal/repository/postgres"
	"calendar/internal/services/event"
	"calendar/internal/services/user"
	"calendar/internal/web/handlers"

	"go.uber.org/fx"
)

func NewApp() *fx.App {
	app := fx.New(
		fx.Provide(
			config.NewAppConfig,
			logger.NewService,
			postgres.NewPostgres,

			func(pg *postgres.Postgres) cleaner.StorageService {
				return pg
			},
			cleaner.NewService,

			sender.NewTelegramChannel,
			sender.NewEmailChannel,
			func(pg *postgres.Postgres) sender.StorageService {
				return pg
			},
			func(pg *postgres.Postgres) sender.ReminderStorageService {
				return pg
			},
			sender.NewService,

			jwt.NewService,

			func(pg *postgres.Postgres) event.StorageService {
				return pg
			},
			func(sender *sender.Service) event.NotificationService {
				return sender
			},
			event.NewEventService,

			func(pg *postgres.Postgres) user.StorageService {
				return pg
			},
			func(jwt *jwt.Service) user.JwtAuthService {
				return jwt
			},
			user.NewService,

			func(event *event.Service) handlers.EventService {
				return event
			},
			handlers.NewCalendarHandler,

			func(user *user.Service) handlers.AuthService {
				return user
			},
			handlers.NewUserHandler,
		),

		fx.Invoke(
			di.StartLoggerService,
			di.StartHttpServer,
			di.StartCleanerService,
			di.StartSenderService,
			di.ClosePostgresOnStop,
		),
	)
	return app
}

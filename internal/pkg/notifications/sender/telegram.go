package sender

import (
	"calendar/internal/config"
	"calendar/internal/logger"
	"context"
	"fmt"
	"strconv"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type TelegramChannel struct {
	bot    *tgbotapi.BotAPI
	logger *logger.Service
	cancel context.CancelFunc
}

func NewTelegramChannel(cfg *config.App, logger *logger.Service) (*TelegramChannel, error) {
	bot, err := tgbotapi.NewBotAPI(cfg.Telegram.BotToken)
	if err != nil {
		logger.Log(zapcore.ErrorLevel, "Failed to create Telegram bot", zap.Error(err))
		return nil, err
	}

	tc := &TelegramChannel{bot: bot, logger: logger}
	ctx, cancel := context.WithCancel(context.Background())
	tc.cancel = cancel
	go tc.listenForStartCommand(ctx)
	return tc, nil
}

// Send delivers a notification via Telegram.
func (t *TelegramChannel) Send(tg string, eventName string, eventText string, eventDate time.Time) error {
	chatId, err := strconv.Atoi(tg)
	if err != nil {
		return fmt.Errorf("invalid chat ID: %w", err)
	}
	msg := tgbotapi.NewMessage(int64(chatId), fmt.Sprintf("Event: %s with details: %s on %s", eventName, eventText, eventDate.Format(time.RFC1123)))
	_, err = t.bot.Send(msg)
	return err
}

func (t *TelegramChannel) Stop() {
	if t.cancel != nil {
		t.cancel()
	}
	t.bot.StopReceivingUpdates()
}

func (t *TelegramChannel) listenForStartCommand(ctx context.Context) {
	t.logger.Log(zapcore.DebugLevel, "Telegram listener started")
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := t.bot.GetUpdatesChan(u)

	for {
		select {
		case <-ctx.Done():
			t.logger.Log(zapcore.DebugLevel, "Telegram listener stopped")
			return
		case update := <-updates:
			if update.Message == nil {
				continue
			}

			if update.Message.Text == "/start" {
				chatID := update.Message.Chat.ID
				username := update.Message.From.UserName

				t.logger.Log(zapcore.DebugLevel, "User started bot",
					zap.String("username", username),
					zap.Int64("chat_id", chatID))

				msg := tgbotapi.NewMessage(chatID,
					fmt.Sprintf("👋 Привет, %s!\n\nТвой chat_id: `%d`\nОтправь его в приложение, чтобы получать уведомления.",
						username, chatID))

				if _, err := t.bot.Send(msg); err != nil {
					t.logger.Log(zapcore.ErrorLevel, "Failed to send Telegram message", zap.Error(err))
				}
			}

		}
	}
}

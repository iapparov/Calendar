package sender

import (
	"calendar/internal/config"
	"fmt"
	"net/smtp"
	"time"
)

type EmailChannel struct {
	smtpEmail string
	auth      smtp.Auth
	address   string
}

func NewEmailChannel(cfg *config.App) *EmailChannel {
	EmailChannel := &EmailChannel{
		smtpEmail: cfg.Mail.SMTPEmail,
		auth:      smtp.PlainAuth("", cfg.Mail.SMTPEmail, cfg.Mail.SMTPPassword, cfg.Mail.SMTPHost),
		address:   fmt.Sprintf("%s:%d", cfg.Mail.SMTPHost, cfg.Mail.SMTPPort),
	}
	return EmailChannel
}

func (s *EmailChannel) Send(email string, eventName string, eventText string, eventDate time.Time) error {
	to := []string{email}
	msg := []byte(fmt.Sprintf(
		"To: %s\r\nSubject: Event Reminder\r\n\r\nEvent: %s\nDetails: %s\nDate: %s\r\n",
		email,
		eventName,
		eventText,
		eventDate.Format(time.RFC1123),
	))

	err := smtp.SendMail(s.address, s.auth, s.smtpEmail, to, msg)
	if err != nil {
		return err
	}
	return nil
}

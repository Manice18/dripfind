package auth

import (
	"context"
	"fmt"
	"log/slog"
	"net/smtp"
	"strings"
)

type Mailer interface {
	SendOTP(ctx context.Context, to, code string) error
}

type LogMailer struct {
	Log *slog.Logger
}

func (m LogMailer) SendOTP(_ context.Context, to, code string) error {
	m.Log.Info("otp code (dev mailer)", "email", to, "code", code)
	return nil
}

type SMTPMailer struct {
	Host     string
	Port     string
	User     string
	Password string
	From     string
	Log      *slog.Logger
}

func (m SMTPMailer) SendOTP(_ context.Context, to, code string) error {
	addr := m.Host + ":" + m.Port
	subject := "Your LOOKBOOK login code"
	body := fmt.Sprintf("Your LOOKBOOK verification code is %s.\n\nIt expires in 10 minutes.\n", code)
	msg := strings.Join([]string{
		"From: " + m.From,
		"To: " + to,
		"Subject: " + subject,
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=UTF-8",
		"",
		body,
	}, "\r\n")

	var auth smtp.Auth
	if m.User != "" {
		auth = smtp.PlainAuth("", m.User, m.Password, m.Host)
	}
	if err := smtp.SendMail(addr, auth, m.From, []string{to}, []byte(msg)); err != nil {
		return err
	}
	m.Log.Info("otp email sent", "email", to)
	return nil
}

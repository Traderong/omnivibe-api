package email

import (
	"fmt"
	"net/smtp"
	"os"
)

type Service struct {
	host     string
	port     string
	username string
	password string
	from     string
}

func NewService() *Service {
	return &Service{
		host:     os.Getenv("SMTP_HOST"),
		port:     os.Getenv("SMTP_PORT"),
		username: os.Getenv("SMTP_USERNAME"),
		password: os.Getenv("SMTP_PASSWORD"),
		from:     os.Getenv("SMTP_FROM"),
	}
}

func (s *Service) send(
	to string,
	subject string,
	body string,
) error {
	if s.host == "" || s.port == "" || s.username == "" ||
		s.password == "" || s.from == "" {
		return fmt.Errorf("SMTP configuration is incomplete")
	}

	message := []byte(
		"From: " + s.from + "\r\n" +
			"To: " + to + "\r\n" +
			"Subject: " + subject + "\r\n" +
			"Content-Type: text/plain; charset=UTF-8\r\n" +
			"\r\n" +
			body,
	)

	auth := smtp.PlainAuth(
		"",
		s.username,
		s.password,
		s.host,
	)

	return smtp.SendMail(
		s.host+":"+s.port,
		auth,
		s.from,
		[]string{to},
		message,
	)
}

func (s *Service) SendVerificationEmail(
	to string,
	username string,
	token string,
) error {
	verifyURL := fmt.Sprintf(
		"%s/api/auth/verify-email?token=%s",
		os.Getenv("APP_URL"),
		token,
	)

	subject := "Verify your OmniVibe email"

	body := fmt.Sprintf(
		"Hello %s,\n\n"+
			"Please verify your OmniVibe email by opening this link:\n\n"+
			"%s\n\n"+
			"This verification link expires in 24 hours.\n\n"+
			"If you did not create an OmniVibe account, you can ignore this email.\n\n"+
			"OmniVibe",
		username,
		verifyURL,
	)

	return s.send(to, subject, body)
}

func (s *Service) SendPasswordResetEmail(
	to string,
	username string,
	token string,
) error {
	resetURL := fmt.Sprintf(
		"%s/reset-password?token=%s",
		os.Getenv("APP_URL"),
		token,
	)

	subject := "Reset your OmniVibe password"

	body := fmt.Sprintf(
		"Hello %s,\n\n"+
			"We received a request to reset your OmniVibe password.\n\n"+
			"Reset your password using this link:\n\n"+
			"%s\n\n"+
			"This reset link expires in 1 hour.\n\n"+
			"If you did not request a password reset, you can safely ignore this email.\n\n"+
			"OmniVibe",
		username,
		resetURL,
	)

	return s.send(to, subject, body)
}

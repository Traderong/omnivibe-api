package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	AppURL       string
	JWTSecret    string
	JWTIssuer    string
	JWTAudience  string
	DBHost       string
	DBPort       uint16
	DBUser       string
	DBPassword   string
	DBName       string
	SMTPHost     string
	SMTPPort     string
	SMTPUsername string
	SMTPPassword string
	SMTPFrom     string
}

func Load() (*Config, error) {
	cfg := &Config{
		AppURL:       strings.TrimSpace(os.Getenv("APP_URL")),
		JWTSecret:    os.Getenv("JWT_SECRET"),
		JWTIssuer:    strings.TrimSpace(os.Getenv("JWT_ISSUER")),
		JWTAudience:  strings.TrimSpace(os.Getenv("JWT_AUDIENCE")),
		DBHost:       strings.TrimSpace(os.Getenv("DB_HOST")),
		DBUser:       strings.TrimSpace(os.Getenv("DB_USER")),
		DBPassword:   os.Getenv("DB_PASSWORD"),
		DBName:       strings.TrimSpace(os.Getenv("DB_NAME")),
		SMTPHost:     strings.TrimSpace(os.Getenv("SMTP_HOST")),
		SMTPPort:     strings.TrimSpace(os.Getenv("SMTP_PORT")),
		SMTPUsername: strings.TrimSpace(os.Getenv("SMTP_USERNAME")),
		SMTPPassword: os.Getenv("SMTP_PASSWORD"),
		SMTPFrom:     strings.TrimSpace(os.Getenv("SMTP_FROM")),
	}

	if cfg.AppURL == "" {
		return nil, errors.New("APP_URL is required")
	}

	if cfg.JWTSecret == "" {
		return nil, errors.New("JWT_SECRET is required")
	}

	if len(cfg.JWTSecret) < 32 {
		return nil, errors.New("JWT_SECRET must be at least 32 characters")
	}

	if cfg.JWTIssuer == "" {
		cfg.JWTIssuer = "omnivibe-api"
	}

	if cfg.JWTAudience == "" {
		cfg.JWTAudience = "omnivibe"
	}

	if cfg.DBHost == "" {
		cfg.DBHost = "localhost"
	}

	if cfg.DBUser == "" {
		return nil, errors.New("DB_USER is required")
	}

	if cfg.DBPassword == "" {
		return nil, errors.New("DB_PASSWORD is required")
	}

	if cfg.DBName == "" {
		return nil, errors.New("DB_NAME is required")
	}

	cfg.DBPort = 5432

	if rawPort := strings.TrimSpace(os.Getenv("DB_PORT")); rawPort != "" {
		port, err := strconv.ParseUint(rawPort, 10, 16)
		if err != nil || port == 0 {
			return nil, fmt.Errorf("invalid DB_PORT: %q", rawPort)
		}

		cfg.DBPort = uint16(port)
	}

	requiredSMTP := map[string]string{
		"SMTP_HOST":     cfg.SMTPHost,
		"SMTP_PORT":     cfg.SMTPPort,
		"SMTP_USERNAME": cfg.SMTPUsername,
		"SMTP_PASSWORD": cfg.SMTPPassword,
		"SMTP_FROM":     cfg.SMTPFrom,
	}

	for name, value := range requiredSMTP {
		if strings.TrimSpace(value) == "" {
			return nil, fmt.Errorf("%s is required", name)
		}
	}

	return cfg, nil
}

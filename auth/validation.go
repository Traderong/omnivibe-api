package auth

import (
	"fmt"
	"net/mail"
	"strings"
	"unicode"
)

func ValidateRegisterRequest(req RegisterRequest) error {
	req.Username = strings.TrimSpace(req.Username)
	req.Email = strings.TrimSpace(req.Email)
	req.DisplayName = strings.TrimSpace(req.DisplayName)

	if req.Username == "" {
		return fmt.Errorf("username is required")
	}

	if len(req.Username) < 3 || len(req.Username) > 30 {
		return fmt.Errorf("username must be between 3 and 30 characters")
	}

	for _, r := range req.Username {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' {
			return fmt.Errorf("username can only contain letters, numbers, and underscores")
		}
	}

	if req.Email == "" {
		return fmt.Errorf("email is required")
	}

	if _, err := mail.ParseAddress(req.Email); err != nil {
		return fmt.Errorf("invalid email address")
	}

	if req.DisplayName == "" {
		return fmt.Errorf("display name is required")
	}

	if len(req.DisplayName) < 2 || len(req.DisplayName) > 100 {
		return fmt.Errorf("display name must be between 2 and 100 characters")
	}

	if len(req.Password) < 8 {
		return fmt.Errorf("password must be at least 8 characters")
	}

	return nil
}
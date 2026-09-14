package auth

import (
	"fmt"
	"net/mail"
	"strings"
	"unicode"
)

func ValidateRegisterRequest(req RegisterRequest) error {
	username := strings.TrimSpace(req.Username)
	email := strings.TrimSpace(req.Email)
	displayName := strings.TrimSpace(req.DisplayName)
	password := req.Password

	if username == "" {
		return fmt.Errorf("username is required")
	}

	if len(username) < 3 || len(username) > 30 {
		return fmt.Errorf("username must be between 3 and 30 characters")
	}

	for _, r := range username {
		if !unicode.IsLetter(r) &&
			!unicode.IsDigit(r) &&
			r != '_' {
			return fmt.Errorf(
				"username can only contain letters, numbers, and underscores",
			)
		}
	}

	if email == "" {
		return fmt.Errorf("email is required")
	}

	if len(email) > 255 {
		return fmt.Errorf("email must not exceed 255 characters")
	}

	parsedEmail, err := mail.ParseAddress(email)
	if err != nil || parsedEmail.Address != email {
		return fmt.Errorf("invalid email address")
	}

	if displayName == "" {
		return fmt.Errorf("display name is required")
	}

	if len(displayName) < 2 || len(displayName) > 100 {
		return fmt.Errorf("display name must be between 2 and 100 characters")
	}

	if password == "" {
		return fmt.Errorf("password is required")
	}

	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters")
	}

	if len(password) > 72 {
		return fmt.Errorf("password must not exceed 72 characters")
	}

	var hasUpper bool
	var hasLower bool
	var hasNumber bool
	var hasSpecial bool

	for _, r := range password {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsDigit(r):
			hasNumber = true
		case unicode.IsPunct(r) || unicode.IsSymbol(r):
			hasSpecial = true
		}
	}

	if !hasUpper {
		return fmt.Errorf("password must contain at least one uppercase letter")
	}

	if !hasLower {
		return fmt.Errorf("password must contain at least one lowercase letter")
	}

	if !hasNumber {
		return fmt.Errorf("password must contain at least one number")
	}

	if !hasSpecial {
		return fmt.Errorf("password must contain at least one special character")
	}

	return nil
}

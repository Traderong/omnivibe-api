package auth

import (
	"context"
	"errors"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

var ErrInvalidRegistration = errors.New("invalid registration data")

func RegisterUser(ctx context.Context, req RegisterRequest) (*User, error) {
	req.Username = strings.TrimSpace(req.Username)
	req.Email = strings.TrimSpace(req.Email)
	req.DisplayName = strings.TrimSpace(req.DisplayName)

	if err := ValidateRegisterRequest(req); err != nil {
		return nil, err
	}

	passwordHash, err := bcrypt.GenerateFromPassword(
		[]byte(req.Password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return nil, err
	}

	user, err := CreateUser(
		ctx,
		req.Username,
		req.Email,
		req.DisplayName,
		string(passwordHash),
	)
	if err != nil {
		return nil, err
	}

	verificationToken, err := GenerateVerificationToken()
	if err != nil {
		return nil, err
	}

	if err := SaveVerificationToken(
		ctx,
		user.ID.String(),
		verificationToken,
	); err != nil {
		return nil, err
	}

	return user, nil
}
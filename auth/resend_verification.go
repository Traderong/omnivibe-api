package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

type ResendVerificationRequest struct {
	Email string `json:"email"`
}

var ErrEmailAlreadyVerified = errors.New("email is already verified")

func ResendVerificationEmail(
	ctx context.Context,
	email string,
) (*User, string, error) {
	email = strings.TrimSpace(email)

	if email == "" {
		return nil, "", errors.New("email is required")
	}

	user, err := GetUserByEmail(ctx, email)
	if err != nil {
		return nil, "", errors.New("unable to process verification request")
	}

	if user.IsVerified {
		return nil, "", ErrEmailAlreadyVerified
	}

	token, err := GenerateVerificationToken()
	if err != nil {
		return nil, "", fmt.Errorf("generate verification token: %w", err)
	}

	if err := SaveVerificationToken(
		ctx,
		user.ID.String(),
		token,
	); err != nil {
		return nil, "", fmt.Errorf("save verification token: %w", err)
	}

	return user, token, nil
}
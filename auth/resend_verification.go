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
) error {
	email = strings.TrimSpace(email)

	if email == "" {
		return errors.New("email is required")
	}

	user, err := GetUserByEmail(ctx, email)
	if err != nil {
		return errors.New("unable to process verification request")
	}

	if user.IsVerified {
		return ErrEmailAlreadyVerified
	}

	token, err := GenerateVerificationToken()
	if err != nil {
		return fmt.Errorf("generate verification token: %w", err)
	}

	if err := SaveVerificationToken(
		ctx,
		user.ID.String(),
		token,
	); err != nil {
		return fmt.Errorf("save verification token: %w", err)
	}

	return nil
}
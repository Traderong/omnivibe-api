package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/Traderong/omnivibe-api/database"
)

type VerifyEmailRequest struct {
	Token string `json:"token"`
}

func GenerateVerificationToken() (string, error) {
	b := make([]byte, 32)

	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate verification token: %w", err)
	}

	return hex.EncodeToString(b), nil
}

func SaveVerificationToken(
	ctx context.Context,
	userID string,
	token string,
) error {
	expiresAt := time.Now().Add(24 * time.Hour)

	_, err := database.DB.Exec(
		ctx,
		`UPDATE users
		 SET email_verification_token = $1,
		     email_verification_expires_at = $2
		 WHERE id = $3`,
		token,
		expiresAt,
		userID,
	)

	if err != nil {
		return fmt.Errorf("save verification token: %w", err)
	}

	return nil
}

func VerifyEmail(ctx context.Context, token string) error {
	if token == "" {
		return errors.New("verification token is required")
	}

	result, err := database.DB.Exec(
		ctx,
		`UPDATE users
		 SET is_verified = TRUE,
		     email_verification_token = NULL,
		     email_verification_expires_at = NULL,
		     email_verified_at = NOW(),
		     updated_at = NOW()
		 WHERE email_verification_token = $1
		   AND email_verification_expires_at > NOW()`,
		token,
	)

	if err != nil {
		return fmt.Errorf("verify email: %w", err)
	}

	if result.RowsAffected() == 0 {
		return errors.New("invalid or expired verification token")
	}

	return nil
}
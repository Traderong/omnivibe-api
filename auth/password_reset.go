package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Traderong/omnivibe-api/database"
	"golang.org/x/crypto/bcrypt"
)

type ForgotPasswordRequest struct {
	Email string `json:"email"`
}

type ResetPasswordRequest struct {
	Token       string `json:"token"`
	NewPassword string `json:"new_password"`
}

func GeneratePasswordResetToken() (string, error) {
	bytes := make([]byte, 32)

	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("generate password reset token: %w", err)
	}

	return hex.EncodeToString(bytes), nil
}

func RequestPasswordReset(
	ctx context.Context,
	email string,
) (*User, string, error) {
	email = strings.TrimSpace(email)

	if email == "" {
		return nil, "", errors.New("email is required")
	}

	user, err := GetUserByEmail(ctx, email)
	if err != nil {
		// Avoid revealing whether an email exists.
		return nil, "", nil
	}

	token, err := GeneratePasswordResetToken()
	if err != nil {
		return nil, "", err
	}

	expiresAt := time.Now().Add(1 * time.Hour)

	_, err = database.DB.Exec(
		ctx,
		`UPDATE users
		 SET password_reset_token = $1,
		     password_reset_expires_at = $2,
		     updated_at = NOW()
		 WHERE id = $3`,
		token,
		expiresAt,
		user.ID,
	)
	if err != nil {
		return nil, "", fmt.Errorf("save password reset token: %w", err)
	}

	return user, token, nil
}

func ResetUserPassword(
	ctx context.Context,
	req ResetPasswordRequest,
) error {
	req.Token = strings.TrimSpace(req.Token)

	if req.Token == "" {
		return errors.New("reset token is required")
	}

	if len(req.NewPassword) < 8 {
		return errors.New("new password must be at least 8 characters")
	}

	var userID string

	err := database.DB.QueryRow(
		ctx,
		`SELECT id
		 FROM users
		 WHERE password_reset_token = $1
		   AND password_reset_expires_at > NOW()
		 LIMIT 1`,
		req.Token,
	).Scan(&userID)

	if err != nil {
		return errors.New("invalid or expired password reset token")
	}

	newHash, err := bcrypt.GenerateFromPassword(
		[]byte(req.NewPassword),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return fmt.Errorf("hash new password: %w", err)
	}

	_, err = database.DB.Exec(
		ctx,
		`UPDATE users
		 SET password_hash = $1,
		     password_reset_token = NULL,
		     password_reset_expires_at = NULL,
		     updated_at = NOW()
		 WHERE id = $2`,
		string(newHash),
		userID,
	)
	if err != nil {
		return fmt.Errorf("update password: %w", err)
	}

	return nil
}

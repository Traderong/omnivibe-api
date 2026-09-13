package auth

import (
	"context"
	"errors"
	"fmt"

	"github.com/Traderong/omnivibe-api/database"
	"golang.org/x/crypto/bcrypt"
)

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

func ChangeUserPassword(
	ctx context.Context,
	userID string,
	req ChangePasswordRequest,
) error {
	if req.CurrentPassword == "" {
		return errors.New("current password is required")
	}

	if len(req.NewPassword) < 8 {
		return errors.New("new password must be at least 8 characters")
	}

	var currentHash string

	err := database.DB.QueryRow(
		ctx,
		`SELECT password_hash FROM users WHERE id = $1 LIMIT 1`,
		userID,
	).Scan(&currentHash)

	if err != nil {
		return errors.New("user not found")
	}

	if err := bcrypt.CompareHashAndPassword(
		[]byte(currentHash),
		[]byte(req.CurrentPassword),
	); err != nil {
		return errors.New("current password is incorrect")
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
		`UPDATE users SET password_hash = $1, updated_at = NOW() WHERE id = $2`,
		string(newHash),
		userID,
	)

	if err != nil {
		return fmt.Errorf("update password: %w", err)
	}

	return nil
}
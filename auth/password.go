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

	if err := ValidateResetPassword(req.NewPassword); err != nil {
		return err
	}

	if req.CurrentPassword == req.NewPassword {
		return errors.New("new password must be different from current password")
	}

	var currentHash string

	err := database.DB.QueryRow(
		ctx,
		`SELECT password_hash
		 FROM users
		 WHERE id = $1
		 LIMIT 1`,
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

	tx, err := database.DB.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin password change transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(
		ctx,
		`UPDATE users
		 SET password_hash = $1,
		     updated_at = NOW()
		 WHERE id = $2`,
		string(newHash),
		userID,
	)
	if err != nil {
		return fmt.Errorf("update password: %w", err)
	}

	// Revoke every existing refresh-token session.
	_, err = tx.Exec(
		ctx,
		`UPDATE refresh_tokens
		 SET revoked_at = NOW()
		 WHERE user_id = $1
		   AND revoked_at IS NULL`,
		userID,
	)
	if err != nil {
		return fmt.Errorf("revoke refresh tokens: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit password change: %w", err)
	}

	return nil
}

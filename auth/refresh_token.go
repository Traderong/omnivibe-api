package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Traderong/omnivibe-api/database"
)

const refreshTokenLifetime = 30 * 24 * time.Hour

func GenerateRefreshToken() (string, error) {
	bytes := make([]byte, 32)

	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("generate refresh token: %w", err)
	}

	return base64.RawURLEncoding.EncodeToString(bytes), nil
}

func hashRefreshToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

func StoreRefreshToken(
	ctx context.Context,
	userID string,
	token string,
) error {
	tokenHash := hashRefreshToken(token)
	expiresAt := time.Now().Add(refreshTokenLifetime)

	_, err := database.DB.Exec(
		ctx,
		`INSERT INTO refresh_tokens
			(user_id, token_hash, expires_at)
		 VALUES ($1, $2, $3)`,
		userID,
		tokenHash,
		expiresAt,
	)
	if err != nil {
		return fmt.Errorf("store refresh token: %w", err)
	}

	return nil
}

func CreateRefreshToken(
	ctx context.Context,
	user *User,
) (string, error) {
	if user == nil {
		return "", errors.New("user is required")
	}

	token, err := GenerateRefreshToken()
	if err != nil {
		return "", err
	}

	if err := StoreRefreshToken(ctx, user.ID.String(), token); err != nil {
		return "", err
	}

	return token, nil
}

func RotateRefreshToken(
	ctx context.Context,
	token string,
) (*User, string, error) {
	token = strings.TrimSpace(token)

	if token == "" {
		return nil, "", errors.New("refresh token is required")
	}

	tokenHash := hashRefreshToken(token)

	tx, err := database.DB.Begin(ctx)
	if err != nil {
		return nil, "", fmt.Errorf("begin refresh token rotation: %w", err)
	}
	defer tx.Rollback(ctx)

	var (
		user       User
		oldTokenID string
	)

	err = tx.QueryRow(
		ctx,
		`SELECT
			rt.id::text,
			u.id,
			u.username,
			u.email,
			u.display_name,
			u.bio,
			u.avatar_url,
			u.is_verified,
			u.is_active,
			u.email_verified_at,
			u.created_at,
			u.updated_at
		 FROM refresh_tokens rt
		 JOIN users u ON u.id = rt.user_id
		 WHERE rt.token_hash = $1
		   AND rt.revoked_at IS NULL
		   AND rt.expires_at > NOW()
		 FOR UPDATE OF rt
		 LIMIT 1`,
		tokenHash,
	).Scan(
		&oldTokenID,
		&user.ID,
		&user.Username,
		&user.Email,
		&user.DisplayName,
		&user.Bio,
		&user.AvatarURL,
		&user.IsVerified,
		&user.IsActive,
		&user.EmailVerifiedAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return nil, "", errors.New("invalid or expired refresh token")
	}

	if !user.IsActive {
		return nil, "", errors.New("account is inactive")
	}

	result, err := tx.Exec(
		ctx,
		`UPDATE refresh_tokens
		 SET revoked_at = NOW()
		 WHERE id = $1
		   AND revoked_at IS NULL`,
		oldTokenID,
	)
	if err != nil {
		return nil, "", fmt.Errorf("revoke refresh token: %w", err)
	}

	if result.RowsAffected() != 1 {
		return nil, "", errors.New("invalid or expired refresh token")
	}

	newToken, err := GenerateRefreshToken()
	if err != nil {
		return nil, "", err
	}

	newTokenHash := hashRefreshToken(newToken)
	newExpiresAt := time.Now().Add(refreshTokenLifetime)

	_, err = tx.Exec(
		ctx,
		`INSERT INTO refresh_tokens
			(user_id, token_hash, expires_at)
		 VALUES ($1, $2, $3)`,
		user.ID,
		newTokenHash,
		newExpiresAt,
	)
	if err != nil {
		return nil, "", fmt.Errorf("store rotated refresh token: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, "", fmt.Errorf("commit refresh token rotation: %w", err)
	}

	return &user, newToken, nil
}

func RevokeRefreshToken(
	ctx context.Context,
	token string,
) error {
	token = strings.TrimSpace(token)

	if token == "" {
		return errors.New("refresh token is required")
	}

	tokenHash := hashRefreshToken(token)

	_, err := database.DB.Exec(
		ctx,
		`UPDATE refresh_tokens
		 SET revoked_at = NOW()
		 WHERE token_hash = $1
		   AND revoked_at IS NULL`,
		tokenHash,
	)
	if err != nil {
		return fmt.Errorf("revoke refresh token: %w", err)
	}

	return nil
}

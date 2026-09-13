
package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/Traderong/omnivibe-api/database"
	"golang.org/x/crypto/bcrypt"
)

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

var ErrInvalidCredentials = errors.New("invalid email or password")

func LoginUser(ctx context.Context, req LoginRequest) (*User, error) {
	req.Email = strings.TrimSpace(req.Email)

	if req.Email == "" || req.Password == "" {
		return nil, ErrInvalidCredentials
	}

	var (
		user         User
		passwordHash string
	)

	query := `
		SELECT
			id,
			username,
			email,
			password_hash,
			display_name,
			bio,
			avatar_url,
			is_verified,
			is_active,
			email_verified_at,
			created_at,
			updated_at
		FROM users
		WHERE email = $1
		LIMIT 1
	`

	err := database.DB.QueryRow(
		ctx,
		query,
		req.Email,
	).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&passwordHash,
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
		return nil, ErrInvalidCredentials
	}

	if !user.IsActive {
		return nil, fmt.Errorf("account is inactive")
	}

	if err := bcrypt.CompareHashAndPassword(
		[]byte(passwordHash),
		[]byte(req.Password),
	); err != nil {
		return nil, ErrInvalidCredentials
	}

	return &user, nil
}
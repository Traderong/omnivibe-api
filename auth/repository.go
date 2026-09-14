package auth

import (
	"context"
	"errors"
	"fmt"

	"github.com/Traderong/omnivibe-api/database"
)

var ErrUserExists = errors.New("username or email already exists")

func CreateUser(
	ctx context.Context,
	username string,
	email string,
	displayName string,
	passwordHash string,
) (*User, error) {
	query := `
		INSERT INTO users (
			username,
			email,
			display_name,
			password_hash
		)
		VALUES ($1, $2, $3, $4)
		RETURNING
			id,
			username,
			email,
			display_name,
			bio,
			avatar_url,
			is_verified,
			is_active,
			profile_private,
			email_verified_at,
			created_at,
			updated_at
	`

	user := &User{}

	err := database.DB.QueryRow(
		ctx,
		query,
		username,
		email,
		displayName,
		passwordHash,
	).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.DisplayName,
		&user.Bio,
		&user.AvatarURL,
		&user.IsVerified,
		&user.IsActive,
		&user.ProfilePrivate,
		&user.EmailVerifiedAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	return user, nil
}

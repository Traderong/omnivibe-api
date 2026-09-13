package auth

import (
	"context"
	"errors"

	"github.com/Traderong/omnivibe-api/database"
)

var ErrUserNotFound = errors.New("user not found")

func GetUserByID(ctx context.Context, userID string) (*User, error) {
	query := `
		SELECT
			id,
			username,
			email,
			display_name,
			bio,
			avatar_url,
			is_verified,
			is_active,
			email_verified_at,
			created_at,
			updated_at
		FROM users
		WHERE id = $1
		LIMIT 1
	`

	user := &User{}

	err := database.DB.QueryRow(ctx, query, userID).Scan(
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
		return nil, ErrUserNotFound
	}

	return user, nil
}
package auth

import (
	"context"
	"fmt"
	"strings"

	"github.com/Traderong/omnivibe-api/database"
)

type UpdateProfileRequest struct {
	DisplayName *string `json:"display_name"`
	Bio         *string `json:"bio"`
	AvatarURL   *string `json:"avatar_url"`
}

func UpdateUserProfile(
	ctx context.Context,
	userID string,
	req UpdateProfileRequest,
) (*User, error) {
	if req.DisplayName != nil {
		value := strings.TrimSpace(*req.DisplayName)

		if value == "" {
			return nil, fmt.Errorf("display name cannot be empty")
		}

		if len(value) < 2 || len(value) > 100 {
			return nil, fmt.Errorf("display name must be between 2 and 100 characters")
		}

		req.DisplayName = &value
	}

	if req.Bio != nil {
		value := strings.TrimSpace(*req.Bio)

		if len(value) > 500 {
			return nil, fmt.Errorf("bio cannot exceed 500 characters")
		}

		req.Bio = &value
	}

	if req.AvatarURL != nil {
		value := strings.TrimSpace(*req.AvatarURL)

		if len(value) > 2048 {
			return nil, fmt.Errorf("avatar URL is too long")
		}

		req.AvatarURL = &value
	}

	query := `
		UPDATE users
		SET
			display_name = COALESCE($1, display_name),
			bio = COALESCE($2, bio),
			avatar_url = COALESCE($3, avatar_url),
			updated_at = NOW()
		WHERE id = $4
		RETURNING
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
	`

	user := &User{}

	err := database.DB.QueryRow(
		ctx,
		query,
		req.DisplayName,
		req.Bio,
		req.AvatarURL,
		userID,
	).Scan(
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
		return nil, fmt.Errorf("update profile: %w", err)
	}

	return user, nil
}
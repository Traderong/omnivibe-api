package auth

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"unicode"

	"github.com/Traderong/omnivibe-api/database"
)

type UpdateProfileRequest struct {
	Username       *string `json:"username"`
	DisplayName    *string `json:"display_name"`
	Bio            *string `json:"bio"`
	AvatarURL      *string `json:"avatar_url"`
	ProfilePrivate *bool   `json:"profile_private"`
}

func ValidateAvatarURL(value string) error {
	value = strings.TrimSpace(value)

	if value == "" {
		return nil
	}

	if len(value) > 2048 {
		return fmt.Errorf("avatar URL is too long")
	}

	parsed, err := url.ParseRequestURI(value)
	if err != nil {
		return fmt.Errorf("invalid avatar URL")
	}

	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("avatar URL must use http or https")
	}

	if parsed.Host == "" {
		return fmt.Errorf("avatar URL must include a valid host")
	}

	return nil
}

func validateUsername(value string) error {
	value = strings.TrimSpace(value)

	if value == "" {
		return fmt.Errorf("username cannot be empty")
	}

	if len(value) < 3 || len(value) > 30 {
		return fmt.Errorf("username must be between 3 and 30 characters")
	}

	for _, r := range value {
		if !unicode.IsLetter(r) &&
			!unicode.IsDigit(r) &&
			r != '_' {
			return fmt.Errorf(
				"username can only contain letters, numbers, and underscores",
			)
		}
	}

	return nil
}

func validateDisplayName(value string) error {
	value = strings.TrimSpace(value)

	if value == "" {
		return fmt.Errorf("display name cannot be empty")
	}

	if len(value) < 2 || len(value) > 100 {
		return fmt.Errorf(
			"display name must be between 2 and 100 characters",
		)
	}

	return nil
}

func validateBio(value string) error {
	if len(value) > 500 {
		return fmt.Errorf("bio cannot exceed 500 characters")
	}

	return nil
}

func UpdateUserProfile(
	ctx context.Context,
	userID string,
	req UpdateProfileRequest,
) (*User, error) {
	if req.Username != nil {
		value := strings.TrimSpace(*req.Username)

		if err := validateUsername(value); err != nil {
			return nil, err
		}

		req.Username = &value
	}

	if req.DisplayName != nil {
		value := strings.TrimSpace(*req.DisplayName)

		if err := validateDisplayName(value); err != nil {
			return nil, err
		}

		req.DisplayName = &value
	}

	if req.Bio != nil {
		value := strings.TrimSpace(*req.Bio)

		if err := validateBio(value); err != nil {
			return nil, err
		}

		req.Bio = &value
	}

	if req.AvatarURL != nil {
		value := strings.TrimSpace(*req.AvatarURL)

		if err := ValidateAvatarURL(value); err != nil {
			return nil, err
		}

		req.AvatarURL = &value
	}

	query := `
		UPDATE users
		SET
			username = COALESCE($1, username),
			display_name = COALESCE($2, display_name),
			bio = COALESCE($3, bio),
			avatar_url = COALESCE($4, avatar_url),
			profile_private = COALESCE($5, profile_private),
			updated_at = NOW()
		WHERE id = $6
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
		req.Username,
		req.DisplayName,
		req.Bio,
		req.AvatarURL,
		req.ProfilePrivate,
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
		&user.ProfilePrivate,
		&user.EmailVerifiedAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		errText := strings.ToLower(err.Error())

		if strings.Contains(errText, "duplicate key") &&
			strings.Contains(errText, "username") {
			return nil, fmt.Errorf("username is already taken")
		}

		return nil, fmt.Errorf("update profile: %w", err)
	}

	return user, nil
}

package auth

import (
	"context"
	"errors"
	"strings"

	"github.com/Traderong/omnivibe-api/database"
)

var ErrPublicProfileNotFound = errors.New("profile not found")

type PublicProfile struct {
	Username    string  `json:"username"`
	DisplayName string  `json:"display_name"`
	Bio         *string `json:"bio,omitempty"`
	AvatarURL   *string `json:"avatar_url,omitempty"`
	IsVerified  bool    `json:"is_verified"`
}

func GetPublicProfileByUsername(
	ctx context.Context,
	username string,
) (*PublicProfile, error) {
	username = strings.TrimSpace(username)

	if username == "" {
		return nil, ErrPublicProfileNotFound
	}

	const query = `
		SELECT
			username,
			display_name,
			bio,
			avatar_url,
			is_verified
		FROM users
		WHERE username = $1
		  AND is_active = TRUE
		  AND profile_private = FALSE
		LIMIT 1
	`

	profile := &PublicProfile{}

	err := database.DB.QueryRow(
		ctx,
		query,
		username,
	).Scan(
		&profile.Username,
		&profile.DisplayName,
		&profile.Bio,
		&profile.AvatarURL,
		&profile.IsVerified,
	)

	if err != nil {
		return nil, ErrPublicProfileNotFound
	}

	return profile, nil
}

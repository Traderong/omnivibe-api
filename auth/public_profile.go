package auth

import (
	"context"
	"errors"
	"strings"

	"github.com/Traderong/omnivibe-api/database"
)

var ErrPublicProfileNotFound = errors.New("profile not found")

type PublicProfile struct {
	Username       string  `json:"username"`
	DisplayName    string  `json:"display_name"`
	Bio            *string `json:"bio,omitempty"`
	AvatarURL      *string `json:"avatar_url,omitempty"`
	IsVerified     bool    `json:"is_verified"`
	FollowersCount int64   `json:"followers_count"`
	FollowingCount int64   `json:"following_count"`
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
			u.username,
			u.display_name,
			u.bio,
			u.avatar_url,
			u.is_verified,
			COUNT(DISTINCT f_in.follower_id) AS followers_count,
			COUNT(DISTINCT f_out.following_id) AS following_count
		FROM users u
		LEFT JOIN follows f_in
			ON f_in.following_id = u.id
		LEFT JOIN follows f_out
			ON f_out.follower_id = u.id
		WHERE u.username = $1
		  AND u.is_active = TRUE
		  AND u.profile_private = FALSE
		GROUP BY
			u.id,
			u.username,
			u.display_name,
			u.bio,
			u.avatar_url,
			u.is_verified
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
		&profile.FollowersCount,
		&profile.FollowingCount,
	)

	if err != nil {
		return nil, ErrPublicProfileNotFound
	}

	return profile, nil
}

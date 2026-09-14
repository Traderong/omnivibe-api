package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/Traderong/omnivibe-api/database"
)

var (
	ErrFollowSelf         = errors.New("you cannot follow yourself")
	ErrAlreadyFollowing   = errors.New("already following user")
	ErrNotFollowing       = errors.New("not following user")
	ErrFollowUserNotFound = errors.New("user not found")
)

type FollowStatus struct {
	Following bool `json:"following"`
}

type FollowUserProfile struct {
	Username    string  `json:"username"`
	DisplayName string  `json:"display_name"`
	AvatarURL   *string `json:"avatar_url,omitempty"`
	IsVerified  bool    `json:"is_verified"`
}

func GetUserIDByUsername(ctx context.Context, username string) (string, error) {
	username = strings.TrimSpace(username)

	if username == "" {
		return "", ErrFollowUserNotFound
	}

	var userID string

	err := database.DB.QueryRow(
		ctx,
		`
		SELECT id
		FROM users
		WHERE username = $1
		  AND is_active = TRUE
		LIMIT 1
		`,
		username,
	).Scan(&userID)

	if err != nil {
		return "", ErrFollowUserNotFound
	}

	return userID, nil
}

func FollowUser(
	ctx context.Context,
	followerID string,
	targetUsername string,
) error {
	targetID, err := GetUserIDByUsername(ctx, targetUsername)
	if err != nil {
		return err
	}

	if followerID == targetID {
		return ErrFollowSelf
	}

	_, err = database.DB.Exec(
		ctx,
		`
		INSERT INTO follows (
			follower_id,
			following_id
		)
		VALUES ($1, $2)
		`,
		followerID,
		targetID,
	)

	if err != nil {
		if strings.Contains(
			strings.ToLower(err.Error()),
			"duplicate key",
		) {
			return ErrAlreadyFollowing
		}

		return fmt.Errorf("follow user: %w", err)
	}

	return nil
}

func UnfollowUser(
	ctx context.Context,
	followerID string,
	targetUsername string,
) error {
	targetID, err := GetUserIDByUsername(ctx, targetUsername)
	if err != nil {
		return err
	}

	result, err := database.DB.Exec(
		ctx,
		`
		DELETE FROM follows
		WHERE follower_id = $1
		  AND following_id = $2
		`,
		followerID,
		targetID,
	)

	if err != nil {
		return fmt.Errorf("unfollow user: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrNotFollowing
	}

	return nil
}

func IsFollowing(
	ctx context.Context,
	followerID string,
	targetUsername string,
) (bool, error) {
	targetID, err := GetUserIDByUsername(ctx, targetUsername)
	if err != nil {
		return false, err
	}

	var exists bool

	err = database.DB.QueryRow(
		ctx,
		`
		SELECT EXISTS (
			SELECT 1
			FROM follows
			WHERE follower_id = $1
			  AND following_id = $2
		)
		`,
		followerID,
		targetID,
	).Scan(&exists)

	if err != nil {
		return false, fmt.Errorf("check follow status: %w", err)
	}

	return exists, nil
}

func GetFollowers(
	ctx context.Context,
	username string,
) ([]FollowUserProfile, error) {
	targetID, err := GetUserIDByUsername(ctx, username)
	if err != nil {
		return nil, err
	}

	rows, err := database.DB.Query(
		ctx,
		`
		SELECT
			u.username,
			u.display_name,
			u.avatar_url,
			u.is_verified
		FROM follows f
		INNER JOIN users u
			ON u.id = f.follower_id
		WHERE f.following_id = $1
		  AND u.is_active = TRUE
		ORDER BY f.created_at DESC
		`,
		targetID,
	)
	if err != nil {
		return nil, fmt.Errorf("get followers: %w", err)
	}
	defer rows.Close()

	users := make([]FollowUserProfile, 0)

	for rows.Next() {
		var user FollowUserProfile

		if err := rows.Scan(
			&user.Username,
			&user.DisplayName,
			&user.AvatarURL,
			&user.IsVerified,
		); err != nil {
			return nil, fmt.Errorf("scan follower: %w", err)
		}

		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate followers: %w", err)
	}

	return users, nil
}

func GetFollowing(
	ctx context.Context,
	username string,
) ([]FollowUserProfile, error) {
	targetID, err := GetUserIDByUsername(ctx, username)
	if err != nil {
		return nil, err
	}

	rows, err := database.DB.Query(
		ctx,
		`
		SELECT
			u.username,
			u.display_name,
			u.avatar_url,
			u.is_verified
		FROM follows f
		INNER JOIN users u
			ON u.id = f.following_id
		WHERE f.follower_id = $1
		  AND u.is_active = TRUE
		ORDER BY f.created_at DESC
		`,
		targetID,
	)
	if err != nil {
		return nil, fmt.Errorf("get following: %w", err)
	}
	defer rows.Close()

	users := make([]FollowUserProfile, 0)

	for rows.Next() {
		var user FollowUserProfile

		if err := rows.Scan(
			&user.Username,
			&user.DisplayName,
			&user.AvatarURL,
			&user.IsVerified,
		); err != nil {
			return nil, fmt.Errorf("scan following: %w", err)
		}

		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate following: %w", err)
	}

	return users, nil
}

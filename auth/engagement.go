package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/Traderong/omnivibe-api/database"
	"github.com/google/uuid"
)

var (
	ErrEngagementPostNotFound = errors.New("post not found")
	ErrAlreadyLiked           = errors.New("post already liked")
	ErrNotLiked               = errors.New("post not liked")
	ErrAlreadyFavorited       = errors.New("post already favorited")
	ErrNotFavorited           = errors.New("post not favorited")
)

type PostEngagement struct {
	LikesCount     int64 `json:"likes_count"`
	FavoritesCount int64 `json:"favorites_count"`
	CommentsCount  int64 `json:"comments_count"`
	Liked          bool  `json:"liked"`
	Favorited      bool  `json:"favorited"`
}

func parsePostID(postID string) (uuid.UUID, error) {
	id, err := uuid.Parse(strings.TrimSpace(postID))
	if err != nil {
		return uuid.Nil, ErrEngagementPostNotFound
	}

	return id, nil
}

func ensurePostExists(ctx context.Context, postID string) error {
	id, err := parsePostID(postID)
	if err != nil {
		return err
	}

	var exists bool

	err = database.DB.QueryRow(
		ctx,
		`
		SELECT EXISTS (
			SELECT 1
			FROM posts
			WHERE id = $1
		)
		`,
		id,
	).Scan(&exists)

	if err != nil {
		return fmt.Errorf("check post: %w", err)
	}

	if !exists {
		return ErrEngagementPostNotFound
	}

	return nil
}

func LikePost(
	ctx context.Context,
	userID string,
	postID string,
) error {
	id, err := parsePostID(postID)
	if err != nil {
		return err
	}

	if err := ensurePostExists(ctx, postID); err != nil {
		return err
	}

	_, err = database.DB.Exec(
		ctx,
		`
		INSERT INTO post_likes (
			post_id,
			user_id
		)
		VALUES ($1, $2)
		`,
		id,
		userID,
	)

	if err != nil {
		if isUniqueViolation(err) {
			return ErrAlreadyLiked
		}

		return fmt.Errorf("like post: %w", err)
	}

	// Find the owner of the post.
	var postOwnerID string

	err = database.DB.QueryRow(
		ctx,
		`
		SELECT user_id::text
		FROM posts
		WHERE id = $1
		`,
		id,
	).Scan(&postOwnerID)

	if err != nil {
		return fmt.Errorf("get post owner: %w", err)
	}

	// Find the username of the user who liked the post.
	var actorUsername string

	err = database.DB.QueryRow(
		ctx,
		`
		SELECT username
		FROM users
		WHERE id = $1
		`,
		userID,
	).Scan(&actorUsername)

	if err != nil {
		return fmt.Errorf("get actor username: %w", err)
	}

	// Create a notification for the post owner.
	// CreateNotification automatically skips self-notifications.
	if err := CreateNotification(
		ctx,
		postOwnerID,
		userID,
		"like",
		postID,
		"",
		fmt.Sprintf("%s liked your post", actorUsername),
	); err != nil {
		return fmt.Errorf("create like notification: %w", err)
	}

	return nil
}

func UnlikePost(
	ctx context.Context,
	userID string,
	postID string,
) error {
	id, err := parsePostID(postID)
	if err != nil {
		return err
	}

	result, err := database.DB.Exec(
		ctx,
		`
		DELETE FROM post_likes
		WHERE post_id = $1
		  AND user_id = $2
		`,
		id,
		userID,
	)

	if err != nil {
		return fmt.Errorf("unlike post: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrNotLiked
	}

	return nil
}

func FavoritePost(
	ctx context.Context,
	userID string,
	postID string,
) error {
	id, err := parsePostID(postID)
	if err != nil {
		return err
	}

	if err := ensurePostExists(ctx, postID); err != nil {
		return err
	}

	_, err = database.DB.Exec(
		ctx,
		`
		INSERT INTO post_favorites (
			post_id,
			user_id
		)
		VALUES ($1, $2)
		`,
		id,
		userID,
	)

	if err != nil {
		if isUniqueViolation(err) {
			return ErrAlreadyFavorited
		}

		return fmt.Errorf("favorite post: %w", err)
	}

	return nil
}

func UnfavoritePost(
	ctx context.Context,
	userID string,
	postID string,
) error {
	id, err := parsePostID(postID)
	if err != nil {
		return err
	}

	result, err := database.DB.Exec(
		ctx,
		`
		DELETE FROM post_favorites
		WHERE post_id = $1
		  AND user_id = $2
		`,
		id,
		userID,
	)

	if err != nil {
		return fmt.Errorf("unfavorite post: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrNotFavorited
	}

	return nil
}

func GetPostEngagement(
	ctx context.Context,
	userID string,
	postID string,
) (*PostEngagement, error) {
	id, err := parsePostID(postID)
	if err != nil {
		return nil, err
	}

	if err := ensurePostExists(ctx, postID); err != nil {
		return nil, err
	}

	engagement := &PostEngagement{}

	err = database.DB.QueryRow(
		ctx,
		`
		SELECT
			(
				SELECT COUNT(*)
				FROM post_likes
				WHERE post_id = $1
			),
			(
				SELECT COUNT(*)
				FROM post_favorites
				WHERE post_id = $1
			),
			(
				SELECT COUNT(*)
				FROM post_comments
				WHERE post_id = $1
			),
			EXISTS (
				SELECT 1
				FROM post_likes
				WHERE post_id = $1
				  AND user_id = $2
			),
			EXISTS (
				SELECT 1
				FROM post_favorites
				WHERE post_id = $1
				  AND user_id = $2
			)
		`,
		id,
		userID,
	).Scan(
		&engagement.LikesCount,
		&engagement.FavoritesCount,
		&engagement.CommentsCount,
		&engagement.Liked,
		&engagement.Favorited,
	)

	if err != nil {
		return nil, fmt.Errorf("get post engagement: %w", err)
	}

	return engagement, nil
}

func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}

	return strings.Contains(
		strings.ToLower(err.Error()),
		"duplicate key",
	)
}

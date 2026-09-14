package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/Traderong/omnivibe-api/database"
)

var (
	ErrPostNotFound = errors.New("post not found")
	ErrInvalidPost  = errors.New("invalid post")
	ErrNotPostOwner = errors.New("you are not the owner of this post")
)

type Post struct {
	ID          uuid.UUID `json:"id"`
	UserID      uuid.UUID `json:"user_id"`
	Username    string    `json:"username"`
	DisplayName string    `json:"display_name"`
	AvatarURL   *string   `json:"avatar_url,omitempty"`
	Content     string    `json:"content"`
	MediaURL    *string   `json:"media_url,omitempty"`
	Visibility  string    `json:"visibility"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CreatePostRequest struct {
	Content    string  `json:"content"`
	MediaURL   *string `json:"media_url"`
	Visibility string  `json:"visibility"`
}

func validatePost(req CreatePostRequest) error {
	content := strings.TrimSpace(req.Content)

	if content == "" {
		return fmt.Errorf("post content is required")
	}

	if len(content) > 5000 {
		return fmt.Errorf("post content cannot exceed 5000 characters")
	}

	visibility := strings.TrimSpace(req.Visibility)

	if visibility == "" {
		visibility = "public"
	}

	switch visibility {
	case "public", "followers", "private":
	default:
		return fmt.Errorf(
			"visibility must be public, followers, or private",
		)
	}

	if req.MediaURL != nil {
		mediaURL := strings.TrimSpace(*req.MediaURL)

		if mediaURL != "" && len(mediaURL) > 2048 {
			return fmt.Errorf("media URL is too long")
		}

		if mediaURL != "" {
			req.MediaURL = &mediaURL
		}
	}

	return nil
}

func CreatePost(
	ctx context.Context,
	userID string,
	req CreatePostRequest,
) (*Post, error) {
	if err := validatePost(req); err != nil {
		return nil, err
	}

	content := strings.TrimSpace(req.Content)

	visibility := strings.TrimSpace(req.Visibility)
	if visibility == "" {
		visibility = "public"
	}

	var mediaURL *string

	if req.MediaURL != nil {
		value := strings.TrimSpace(*req.MediaURL)

		if value != "" {
			mediaURL = &value
		}
	}

	const query = `
		INSERT INTO posts (
			user_id,
			content,
			media_url,
			visibility
		)
		VALUES ($1, $2, $3, $4)
		RETURNING id, user_id, content, media_url, visibility, created_at, updated_at
	`

	post := &Post{}

	err := database.DB.QueryRow(
		ctx,
		query,
		userID,
		content,
		mediaURL,
		visibility,
	).Scan(
		&post.ID,
		&post.UserID,
		&post.Content,
		&post.MediaURL,
		&post.Visibility,
		&post.CreatedAt,
		&post.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("create post: %w", err)
	}

	return GetPostByID(ctx, post.ID.String())
}

func GetPostByID(
	ctx context.Context,
	postID string,
) (*Post, error) {
	const query = `
		SELECT
			p.id,
			p.user_id,
			u.username,
			u.display_name,
			u.avatar_url,
			p.content,
			p.media_url,
			p.visibility,
			p.created_at,
			p.updated_at
		FROM posts p
		INNER JOIN users u
			ON u.id = p.user_id
		WHERE p.id = $1
		  AND u.is_active = TRUE
		LIMIT 1
	`

	post := &Post{}

	err := database.DB.QueryRow(
		ctx,
		query,
		postID,
	).Scan(
		&post.ID,
		&post.UserID,
		&post.Username,
		&post.DisplayName,
		&post.AvatarURL,
		&post.Content,
		&post.MediaURL,
		&post.Visibility,
		&post.CreatedAt,
		&post.UpdatedAt,
	)

	if err != nil {
		return nil, ErrPostNotFound
	}

	return post, nil
}

func DeletePost(
	ctx context.Context,
	userID string,
	postID string,
) error {
	result, err := database.DB.Exec(
		ctx,
		`
		DELETE FROM posts
		WHERE id = $1
		  AND user_id = $2
		`,
		postID,
		userID,
	)

	if err != nil {
		return fmt.Errorf("delete post: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrPostNotFound
	}

	return nil
}

func GetUserPosts(
	ctx context.Context,
	username string,
) ([]Post, error) {
	const query = `
		SELECT
			p.id,
			p.user_id,
			u.username,
			u.display_name,
			u.avatar_url,
			p.content,
			p.media_url,
			p.visibility,
			p.created_at,
			p.updated_at
		FROM posts p
		INNER JOIN users u
			ON u.id = p.user_id
		WHERE u.username = $1
		  AND u.is_active = TRUE
		ORDER BY p.created_at DESC
	`

	rows, err := database.DB.Query(
		ctx,
		query,
		username,
	)
	if err != nil {
		return nil, fmt.Errorf("get user posts: %w", err)
	}
	defer rows.Close()

	posts := make([]Post, 0)

	for rows.Next() {
		var post Post

		if err := rows.Scan(
			&post.ID,
			&post.UserID,
			&post.Username,
			&post.DisplayName,
			&post.AvatarURL,
			&post.Content,
			&post.MediaURL,
			&post.Visibility,
			&post.CreatedAt,
			&post.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan post: %w", err)
		}

		posts = append(posts, post)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate posts: %w", err)
	}

	return posts, nil
}

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
	ErrCommentPostNotFound = errors.New("post not found")
	ErrCommentNotFound     = errors.New("comment not found")
)

type Comment struct {
	ID          string `json:"id"`
	PostID      string `json:"post_id"`
	UserID      string `json:"user_id"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	AvatarURL   string `json:"avatar_url"`
	Content     string `json:"content"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

type CreateCommentRequest struct {
	Content string `json:"content"`
}

func ValidateCreateCommentRequest(req CreateCommentRequest) error {
	content := strings.TrimSpace(req.Content)

	if content == "" {
		return errors.New("comment content is required")
	}

	if len([]rune(content)) > 2000 {
		return errors.New("comment content must not exceed 2000 characters")
	}

	return nil
}

func CreateComment(
	ctx context.Context,
	userID string,
	postID string,
	req CreateCommentRequest,
) (*Comment, error) {
	if err := ValidateCreateCommentRequest(req); err != nil {
		return nil, err
	}

	postUUID, err := uuid.Parse(strings.TrimSpace(postID))
	if err != nil {
		return nil, ErrCommentPostNotFound
	}

	userUUID, err := uuid.Parse(strings.TrimSpace(userID))
	if err != nil {
		return nil, errors.New("invalid user id")
	}

	// Check that the post exists.
	var postOwnerID string

	err = database.DB.QueryRow(
		ctx,
		`
		SELECT user_id::text
		FROM posts
		WHERE id = $1
		`,
		postUUID,
	).Scan(&postOwnerID)

	if err != nil {
		return nil, ErrCommentPostNotFound
	}

	content := strings.TrimSpace(req.Content)

	comment := &Comment{}

	err = database.DB.QueryRow(
		ctx,
		`
		INSERT INTO post_comments (
			post_id,
			user_id,
			content
		)
		VALUES ($1, $2, $3)
		RETURNING
			id::text,
			post_id::text,
			user_id::text,
			content,
			created_at::text,
			updated_at::text
		`,
		postUUID,
		userUUID,
		content,
	).Scan(
		&comment.ID,
		&comment.PostID,
		&comment.UserID,
		&comment.Content,
		&comment.CreatedAt,
		&comment.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("create comment: %w", err)
	}

	// Get the comment author's profile information.
	err = database.DB.QueryRow(
		ctx,
		`
		SELECT
			username,
			COALESCE(display_name, ''),
			COALESCE(avatar_url, '')
		FROM users
		WHERE id = $1
		`,
		userUUID,
	).Scan(
		&comment.Username,
		&comment.DisplayName,
		&comment.AvatarURL,
	)

	if err != nil {
		return nil, fmt.Errorf("get comment user: %w", err)
	}

	// Create a notification for the post owner.
	// CreateNotification automatically skips notifications
	// when the commenter is the post owner.
	if err := CreateNotification(
		ctx,
		postOwnerID,
		userID,
		"comment",
		postID,
		comment.ID,
		fmt.Sprintf("%s commented on your post", comment.Username),
	); err != nil {
		return nil, fmt.Errorf("create comment notification: %w", err)
	}

	return comment, nil
}

func GetPostComments(
	ctx context.Context,
	postID string,
) ([]*Comment, error) {
	postUUID, err := uuid.Parse(strings.TrimSpace(postID))
	if err != nil {
		return nil, ErrCommentPostNotFound
	}

	var exists bool

	if err := database.DB.QueryRow(
		ctx,
		`
		SELECT EXISTS (
			SELECT 1
			FROM posts
			WHERE id = $1
		)
		`,
		postUUID,
	).Scan(&exists); err != nil {
		return nil, fmt.Errorf("check post: %w", err)
	}

	if !exists {
		return nil, ErrCommentPostNotFound
	}

	rows, err := database.DB.Query(
		ctx,
		`
		SELECT
			pc.id::text,
			pc.post_id::text,
			pc.user_id::text,
			u.username,
			COALESCE(u.display_name, ''),
			COALESCE(u.avatar_url, ''),
			pc.content,
			pc.created_at::text,
			pc.updated_at::text
		FROM post_comments pc
		INNER JOIN users u
			ON u.id = pc.user_id
		WHERE pc.post_id = $1
		ORDER BY pc.created_at ASC
		`,
		postUUID,
	)
	if err != nil {
		return nil, fmt.Errorf("get post comments: %w", err)
	}
	defer rows.Close()

	comments := make([]*Comment, 0)

	for rows.Next() {
		comment := &Comment{}

		if err := rows.Scan(
			&comment.ID,
			&comment.PostID,
			&comment.UserID,
			&comment.Username,
			&comment.DisplayName,
			&comment.AvatarURL,
			&comment.Content,
			&comment.CreatedAt,
			&comment.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan comment: %w", err)
		}

		comments = append(comments, comment)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate comments: %w", err)
	}

	return comments, nil
}

func DeleteComment(
	ctx context.Context,
	userID string,
	commentID string,
) error {
	commentUUID, err := uuid.Parse(strings.TrimSpace(commentID))
	if err != nil {
		return ErrCommentNotFound
	}

	userUUID, err := uuid.Parse(strings.TrimSpace(userID))
	if err != nil {
		return ErrCommentNotFound
	}

	result, err := database.DB.Exec(
		ctx,
		`
		DELETE FROM post_comments
		WHERE id = $1
		  AND user_id = $2
		`,
		commentUUID,
		userUUID,
	)
	if err != nil {
		return fmt.Errorf("delete comment: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrCommentNotFound
	}

	return nil
}

func GetUserComments(
	ctx context.Context,
	userID string,
) ([]*Comment, error) {
	userUUID, err := uuid.Parse(strings.TrimSpace(userID))
	if err != nil {
		return nil, errors.New("invalid user id")
	}

	rows, err := database.DB.Query(
		ctx,
		`
		SELECT
			pc.id::text,
			pc.post_id::text,
			pc.user_id::text,
			u.username,
			COALESCE(u.display_name, ''),
			COALESCE(u.avatar_url, ''),
			pc.content,
			pc.created_at::text,
			pc.updated_at::text
		FROM post_comments pc
		INNER JOIN users u
			ON u.id = pc.user_id
		WHERE pc.user_id = $1
		ORDER BY pc.created_at DESC
		`,
		userUUID,
	)
	if err != nil {
		return nil, fmt.Errorf("get user comments: %w", err)
	}
	defer rows.Close()

	comments := make([]*Comment, 0)

	for rows.Next() {
		comment := &Comment{}

		if err := rows.Scan(
			&comment.ID,
			&comment.PostID,
			&comment.UserID,
			&comment.Username,
			&comment.DisplayName,
			&comment.AvatarURL,
			&comment.Content,
			&comment.CreatedAt,
			&comment.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan user comment: %w", err)
		}

		comments = append(comments, comment)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate user comments: %w", err)
	}

	return comments, nil
}

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
	ErrNotificationNotFound = errors.New("notification not found")
)

type Notification struct {
	ID          string  `json:"id"`
	RecipientID string  `json:"recipient_id"`
	ActorID     *string `json:"actor_id,omitempty"`
	ActorName   string  `json:"actor_name,omitempty"`
	ActorAvatar string  `json:"actor_avatar,omitempty"`
	Type        string  `json:"type"`
	PostID      *string `json:"post_id,omitempty"`
	CommentID   *string `json:"comment_id,omitempty"`
	Message     string  `json:"message"`
	IsRead      bool    `json:"is_read"`
	CreatedAt   string  `json:"created_at"`
}

func CreateNotification(
	ctx context.Context,
	recipientID string,
	actorID string,
	notificationType string,
	postID string,
	commentID string,
	message string,
) error {
	recipientUUID, err := uuid.Parse(strings.TrimSpace(recipientID))
	if err != nil {
		return errors.New("invalid recipient id")
	}

	actorUUID, err := uuid.Parse(strings.TrimSpace(actorID))
	if err != nil {
		return errors.New("invalid actor id")
	}

	if recipientUUID == actorUUID {
		// Never notify a user about their own action.
		return nil
	}

	if strings.TrimSpace(message) == "" {
		return errors.New("notification message is required")
	}

	var postUUID *uuid.UUID
	if strings.TrimSpace(postID) != "" {
		id, err := uuid.Parse(strings.TrimSpace(postID))
		if err != nil {
			return errors.New("invalid post id")
		}
		postUUID = &id
	}

	var commentUUID *uuid.UUID
	if strings.TrimSpace(commentID) != "" {
		id, err := uuid.Parse(strings.TrimSpace(commentID))
		if err != nil {
			return errors.New("invalid comment id")
		}
		commentUUID = &id
	}

	_, err = database.DB.Exec(
		ctx,
		`
		INSERT INTO notifications (
			recipient_id,
			actor_id,
			type,
			post_id,
			comment_id,
			message
		)
		VALUES ($1, $2, $3, $4, $5, $6)
		`,
		recipientUUID,
		actorUUID,
		notificationType,
		postUUID,
		commentUUID,
		strings.TrimSpace(message),
	)
	if err != nil {
		return fmt.Errorf("create notification: %w", err)
	}

	return nil
}

func GetNotifications(
	ctx context.Context,
	userID string,
) ([]*Notification, error) {
	userUUID, err := uuid.Parse(strings.TrimSpace(userID))
	if err != nil {
		return nil, errors.New("invalid user id")
	}

	rows, err := database.DB.Query(
		ctx,
		`
		SELECT
			n.id::text,
			n.recipient_id::text,
			n.actor_id::text,
			COALESCE(
				NULLIF(u.display_name, ''),
				u.username,
				''
			),
			COALESCE(u.avatar_url, ''),
			n.type,
			n.post_id::text,
			n.comment_id::text,
			n.message,
			n.is_read,
			n.created_at::text
		FROM notifications n
		LEFT JOIN users u
			ON u.id = n.actor_id
		WHERE n.recipient_id = $1
		ORDER BY n.created_at DESC
		`,
		userUUID,
	)
	if err != nil {
		return nil, fmt.Errorf("get notifications: %w", err)
	}
	defer rows.Close()

	notifications := make([]*Notification, 0)

	for rows.Next() {
		notification := &Notification{}

		var actorID *string
		var postID *string
		var commentID *string

		if err := rows.Scan(
			&notification.ID,
			&notification.RecipientID,
			&actorID,
			&notification.ActorName,
			&notification.ActorAvatar,
			&notification.Type,
			&postID,
			&commentID,
			&notification.Message,
			&notification.IsRead,
			&notification.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan notification: %w", err)
		}

		notification.ActorID = actorID
		notification.PostID = postID
		notification.CommentID = commentID

		notifications = append(notifications, notification)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate notifications: %w", err)
	}

	return notifications, nil
}

func GetUnreadNotificationCount(
	ctx context.Context,
	userID string,
) (int64, error) {
	userUUID, err := uuid.Parse(strings.TrimSpace(userID))
	if err != nil {
		return 0, errors.New("invalid user id")
	}

	var count int64

	err = database.DB.QueryRow(
		ctx,
		`
		SELECT COUNT(*)
		FROM notifications
		WHERE recipient_id = $1
		  AND is_read = FALSE
		`,
		userUUID,
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("get unread notification count: %w", err)
	}

	return count, nil
}

func MarkNotificationRead(
	ctx context.Context,
	userID string,
	notificationID string,
) error {
	userUUID, err := uuid.Parse(strings.TrimSpace(userID))
	if err != nil {
		return ErrNotificationNotFound
	}

	notificationUUID, err := uuid.Parse(strings.TrimSpace(notificationID))
	if err != nil {
		return ErrNotificationNotFound
	}

	result, err := database.DB.Exec(
		ctx,
		`
		UPDATE notifications
		SET is_read = TRUE
		WHERE id = $1
		  AND recipient_id = $2
		`,
		notificationUUID,
		userUUID,
	)
	if err != nil {
		return fmt.Errorf("mark notification read: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrNotificationNotFound
	}

	return nil
}

func MarkAllNotificationsRead(
	ctx context.Context,
	userID string,
) error {
	userUUID, err := uuid.Parse(strings.TrimSpace(userID))
	if err != nil {
		return errors.New("invalid user id")
	}

	_, err = database.DB.Exec(
		ctx,
		`
		UPDATE notifications
		SET is_read = TRUE
		WHERE recipient_id = $1
		  AND is_read = FALSE
		`,
		userUUID,
	)
	if err != nil {
		return fmt.Errorf("mark all notifications read: %w", err)
	}

	return nil
}

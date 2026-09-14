package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/Traderong/omnivibe-api/auth"
)

func notificationIDFromPath(r *http.Request) string {
	const prefix = "/api/notifications/"

	path := strings.TrimPrefix(r.URL.Path, prefix)

	return strings.Trim(path, "/")
}

// GET /api/notifications
func GetNotifications(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}

	notifications, err := auth.GetNotifications(
		r.Context(),
		userID,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	_ = json.NewEncoder(w).Encode(notifications)
}

// GET /api/notifications/unread-count
func GetUnreadNotificationCount(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}

	count, err := auth.GetUnreadNotificationCount(
		r.Context(),
		userID,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]int64{
		"unread_count": count,
	})
}

// PATCH /api/notifications/{id}/read
func MarkNotificationRead(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}

	notificationID := notificationIDFromPath(r)

	notificationID = strings.TrimSuffix(notificationID, "/read")
	notificationID = strings.Trim(notificationID, "/")

	if err := auth.MarkNotificationRead(
		r.Context(),
		userID,
		notificationID,
	); err != nil {
		switch {
		case errors.Is(err, auth.ErrNotificationNotFound):
			http.Error(w, "notification not found", http.StatusNotFound)

		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		return
	}

	_ = json.NewEncoder(w).Encode(map[string]string{
		"message": "notification marked as read",
	})
}

// POST /api/notifications/read-all
func MarkAllNotificationsRead(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}

	if err := auth.MarkAllNotificationsRead(
		r.Context(),
		userID,
	); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]string{
		"message": "all notifications marked as read",
	})
}

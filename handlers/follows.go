package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/Traderong/omnivibe-api/auth"
)

func extractUsername(path, prefix string) string {
	username := strings.TrimPrefix(path, prefix)
	username = strings.TrimSpace(username)
	username = strings.Trim(username, "/")

	return username
}

func FollowUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	followerID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}

	username := extractUsername(
		r.URL.Path,
		"/api/users/",
	)

	username = strings.TrimSuffix(username, "/follow")

	if username == "" || strings.Contains(username, "/") {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	// Create the follow relationship.
	if err := auth.FollowUser(
		r.Context(),
		followerID,
		username,
	); err != nil {
		switch {
		case errors.Is(err, auth.ErrFollowUserNotFound):
			http.Error(w, "user not found", http.StatusNotFound)

		case errors.Is(err, auth.ErrFollowSelf):
			http.Error(w, "you cannot follow yourself", http.StatusBadRequest)

		case errors.Is(err, auth.ErrAlreadyFollowing):
			http.Error(w, "already following user", http.StatusConflict)

		default:
			http.Error(
				w,
				"failed to follow user",
				http.StatusInternalServerError,
			)
		}

		return
	}

	// Find the user who was followed.
	targetUserID, err := auth.GetUserIDByUsername(
		r.Context(),
		username,
	)
	if err != nil {
		http.Error(
			w,
			"failed to create follow notification",
			http.StatusInternalServerError,
		)
		return
	}

	// Create a notification for the followed user.
	if err := auth.CreateNotification(
		r.Context(),
		targetUserID,
		followerID,
		"follow",
		"",
		"",
		"started following you",
	); err != nil {
		http.Error(
			w,
			"failed to create follow notification",
			http.StatusInternalServerError,
		)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"message":   "user followed successfully",
		"following": true,
	})
}

func UnfollowUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	followerID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}

	username := extractUsername(
		r.URL.Path,
		"/api/users/",
	)

	username = strings.TrimSuffix(username, "/follow")

	if username == "" || strings.Contains(username, "/") {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	if err := auth.UnfollowUser(
		r.Context(),
		followerID,
		username,
	); err != nil {
		switch {
		case errors.Is(err, auth.ErrFollowUserNotFound):
			http.Error(w, "user not found", http.StatusNotFound)

		case errors.Is(err, auth.ErrNotFollowing):
			http.Error(w, "not following user", http.StatusConflict)

		default:
			http.Error(
				w,
				"failed to unfollow user",
				http.StatusInternalServerError,
			)
		}

		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"message":   "user unfollowed successfully",
		"following": false,
	})
}

func FollowStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	followerID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}

	username := extractUsername(
		r.URL.Path,
		"/api/users/",
	)

	username = strings.TrimSuffix(username, "/follow-status")

	if username == "" || strings.Contains(username, "/") {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	following, err := auth.IsFollowing(
		r.Context(),
		followerID,
		username,
	)
	if err != nil {
		if errors.Is(err, auth.ErrFollowUserNotFound) {
			http.Error(w, "user not found", http.StatusNotFound)
			return
		}

		http.Error(
			w,
			"failed to check follow status",
			http.StatusInternalServerError,
		)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	_ = json.NewEncoder(w).Encode(auth.FollowStatus{
		Following: following,
	})
}

func Followers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	username := extractUsername(
		r.URL.Path,
		"/api/users/",
	)

	username = strings.TrimSuffix(username, "/followers")

	if username == "" || strings.Contains(username, "/") {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	users, err := auth.GetFollowers(
		r.Context(),
		username,
	)
	if err != nil {
		if errors.Is(err, auth.ErrFollowUserNotFound) {
			http.Error(w, "user not found", http.StatusNotFound)
			return
		}

		http.Error(
			w,
			"failed to get followers",
			http.StatusInternalServerError,
		)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	_ = json.NewEncoder(w).Encode(users)
}

func Following(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	username := extractUsername(
		r.URL.Path,
		"/api/users/",
	)

	username = strings.TrimSuffix(username, "/following")

	if username == "" || strings.Contains(username, "/") {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	users, err := auth.GetFollowing(
		r.Context(),
		username,
	)
	if err != nil {
		if errors.Is(err, auth.ErrFollowUserNotFound) {
			http.Error(w, "user not found", http.StatusNotFound)
			return
		}

		http.Error(
			w,
			"failed to get following",
			http.StatusInternalServerError,
		)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	_ = json.NewEncoder(w).Encode(users)
}

func FollowStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	username := extractUsername(
		r.URL.Path,
		"/api/users/",
	)

	username = strings.TrimSuffix(username, "/stats")

	if username == "" || strings.Contains(username, "/") {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	stats, err := auth.GetFollowStats(
		r.Context(),
		username,
	)
	if err != nil {
		if errors.Is(err, auth.ErrFollowUserNotFound) {
			http.Error(w, "user not found", http.StatusNotFound)
			return
		}

		http.Error(
			w,
			"failed to get profile stats",
			http.StatusInternalServerError,
		)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	_ = json.NewEncoder(w).Encode(stats)
}

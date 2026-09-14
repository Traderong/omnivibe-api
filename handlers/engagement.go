package handlers

import (
	"net/http"
	"strings"

	"github.com/Traderong/omnivibe-api/auth"
)

func postIDFromPath(r *http.Request) string {
	const prefix = "/api/posts/"

	path := strings.TrimPrefix(r.URL.Path, prefix)

	path = strings.TrimSuffix(path, "/engagement")
	path = strings.TrimSuffix(path, "/like")
	path = strings.TrimSuffix(path, "/favorite")

	return strings.Trim(path, "/")
}

func LikePost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}

	postID := postIDFromPath(r)

	if err := auth.LikePost(r.Context(), userID, postID); err != nil {
		switch err {
		case auth.ErrEngagementPostNotFound:
			http.Error(w, "post not found", http.StatusNotFound)
		case auth.ErrAlreadyLiked:
			http.Error(w, "post already liked", http.StatusConflict)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	_ = WriteJSON(w, http.StatusOK, map[string]string{
		"message": "post liked successfully",
	})
}

func UnlikePost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}

	postID := postIDFromPath(r)

	if err := auth.UnlikePost(r.Context(), userID, postID); err != nil {
		switch err {
		case auth.ErrNotLiked:
			http.Error(w, "post not liked", http.StatusConflict)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	_ = WriteJSON(w, http.StatusOK, map[string]string{
		"message": "post unliked successfully",
	})
}

func FavoritePost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}

	postID := postIDFromPath(r)

	if err := auth.FavoritePost(r.Context(), userID, postID); err != nil {
		switch err {
		case auth.ErrEngagementPostNotFound:
			http.Error(w, "post not found", http.StatusNotFound)
		case auth.ErrAlreadyFavorited:
			http.Error(w, "post already favorited", http.StatusConflict)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	_ = WriteJSON(w, http.StatusOK, map[string]string{
		"message": "post favorited successfully",
	})
}

func UnfavoritePost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}

	postID := postIDFromPath(r)

	if err := auth.UnfavoritePost(r.Context(), userID, postID); err != nil {
		switch err {
		case auth.ErrNotFavorited:
			http.Error(w, "post not favorited", http.StatusConflict)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	_ = WriteJSON(w, http.StatusOK, map[string]string{
		"message": "post unfavorited successfully",
	})
}

func PostEngagement(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}

	postID := postIDFromPath(r)

	engagement, err := auth.GetPostEngagement(
		r.Context(),
		userID,
		postID,
	)
	if err != nil {
		switch err {
		case auth.ErrEngagementPostNotFound:
			http.Error(w, "post not found", http.StatusNotFound)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	_ = WriteJSON(w, http.StatusOK, engagement)
}

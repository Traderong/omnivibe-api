package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/Traderong/omnivibe-api/auth"
)

func CreatePost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}

	var req auth.CreatePostRequest

	if err := DecodeJSON(w, r, &req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	post, err := auth.CreatePost(
		r.Context(),
		userID,
		req,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	_ = json.NewEncoder(w).Encode(post)
}

func GetPost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	postID := strings.TrimPrefix(
		r.URL.Path,
		"/api/posts/",
	)
	postID = strings.Trim(postID, "/")

	if postID == "" || strings.Contains(postID, "/") {
		http.Error(w, "post not found", http.StatusNotFound)
		return
	}

	post, err := auth.GetPostByID(
		r.Context(),
		postID,
	)
	if err != nil {
		http.Error(w, "post not found", http.StatusNotFound)
		return
	}

	_ = json.NewEncoder(w).Encode(post)
}

func DeletePost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}

	postID := strings.TrimPrefix(
		r.URL.Path,
		"/api/posts/",
	)
	postID = strings.Trim(postID, "/")

	if postID == "" || strings.Contains(postID, "/") {
		http.Error(w, "post not found", http.StatusNotFound)
		return
	}

	err := auth.DeletePost(
		r.Context(),
		userID,
		postID,
	)
	if err != nil {
		if errors.Is(err, auth.ErrPostNotFound) {
			http.Error(w, "post not found", http.StatusNotFound)
			return
		}

		http.Error(
			w,
			"failed to delete post",
			http.StatusInternalServerError,
		)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	_ = json.NewEncoder(w).Encode(map[string]string{
		"message": "post deleted successfully",
	})
}

func GetUserPosts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	username := strings.TrimPrefix(
		r.URL.Path,
		"/api/users/",
	)
	username = strings.TrimSuffix(username, "/posts")
	username = strings.Trim(username, "/")

	if username == "" || strings.Contains(username, "/") {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	posts, err := auth.GetUserPosts(
		r.Context(),
		username,
	)
	if err != nil {
		http.Error(
			w,
			"failed to get user posts",
			http.StatusInternalServerError,
		)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	_ = json.NewEncoder(w).Encode(posts)
}

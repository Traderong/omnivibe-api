package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/Traderong/omnivibe-api/auth"
)

func commentPostIDFromPath(r *http.Request) string {
	const prefix = "/api/posts/"

	path := strings.TrimPrefix(r.URL.Path, prefix)
	path = strings.TrimSuffix(path, "/comments")

	return strings.Trim(path, "/")
}

func commentIDFromPath(r *http.Request) string {
	const prefix = "/api/comments/"

	path := strings.TrimPrefix(r.URL.Path, prefix)

	return strings.Trim(path, "/")
}

func CreateComment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}

	postID := commentPostIDFromPath(r)

	var req auth.CreateCommentRequest

	if err := DecodeJSON(w, r, &req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	comment, err := auth.CreateComment(
		r.Context(),
		userID,
		postID,
		req,
	)
	if err != nil {
		switch {
		case errors.Is(err, auth.ErrCommentPostNotFound):
			http.Error(w, "post not found", http.StatusNotFound)
		default:
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	_ = json.NewEncoder(w).Encode(comment)
}

func GetPostComments(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	postID := commentPostIDFromPath(r)

	comments, err := auth.GetPostComments(
		r.Context(),
		postID,
	)
	if err != nil {
		switch {
		case errors.Is(err, auth.ErrCommentPostNotFound):
			http.Error(w, "post not found", http.StatusNotFound)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	_ = json.NewEncoder(w).Encode(comments)
}

func DeleteComment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}

	commentID := commentIDFromPath(r)

	err := auth.DeleteComment(
		r.Context(),
		userID,
		commentID,
	)
	if err != nil {
		switch {
		case errors.Is(err, auth.ErrCommentNotFound):
			http.Error(w, "comment not found", http.StatusNotFound)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	_ = json.NewEncoder(w).Encode(map[string]string{
		"message": "comment deleted successfully",
	})
}

func GetUserComments(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}

	comments, err := auth.GetUserComments(
		r.Context(),
		userID,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	_ = json.NewEncoder(w).Encode(comments)
}

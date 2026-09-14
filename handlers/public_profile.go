package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/Traderong/omnivibe-api/auth"
)

func PublicProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(
			w,
			"method not allowed",
			http.StatusMethodNotAllowed,
		)
		return
	}

	username := strings.TrimPrefix(
		r.URL.Path,
		"/api/users/",
	)

	username = strings.TrimSpace(username)

	if username == "" || strings.Contains(username, "/") {
		http.Error(
			w,
			"profile not found",
			http.StatusNotFound,
		)
		return
	}

	profile, err := auth.GetPublicProfileByUsername(
		r.Context(),
		username,
	)
	if err != nil {
		http.Error(
			w,
			"profile not found",
			http.StatusNotFound,
		)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	_ = json.NewEncoder(w).Encode(profile)
}

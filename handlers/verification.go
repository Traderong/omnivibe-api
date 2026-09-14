package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/Traderong/omnivibe-api/auth"
)

func VerifyEmail(w http.ResponseWriter, r *http.Request) {
	var token string

	switch r.Method {
	case http.MethodGet:
		token = r.URL.Query().Get("token")

	case http.MethodPost:
		var req auth.VerifyEmailRequest

		if err := DecodeJSON(w, r, &req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		token = req.Token

	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if token == "" {
		http.Error(
			w,
			"verification token is required",
			http.StatusBadRequest,
		)
		return
	}

	if err := auth.VerifyEmail(r.Context(), token); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	_ = json.NewEncoder(w).Encode(map[string]string{
		"message": "email verified successfully",
	})
}

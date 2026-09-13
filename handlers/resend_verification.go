package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Traderong/omnivibe-api/auth"
)

func ResendVerification(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req auth.ResendVerificationRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	err := auth.ResendVerificationEmail(r.Context(), req.Email)
	if err != nil {
		if errors.Is(err, auth.ErrEmailAlreadyVerified) {
			http.Error(w, "email is already verified", http.StatusConflict)
			return
		}

		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	_ = json.NewEncoder(w).Encode(map[string]string{
		"message": "verification email request processed successfully",
	})
}
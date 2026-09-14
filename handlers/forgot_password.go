package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/Traderong/omnivibe-api/auth"
	"github.com/Traderong/omnivibe-api/email"
)

func ForgotPassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req auth.ForgotPasswordRequest

	if err := DecodeJSON(w, r, &req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	user, resetToken, err := auth.RequestPasswordReset(
		r.Context(),
		req.Email,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Do not reveal whether the email exists.
	if user != nil && resetToken != "" {
		emailService := email.NewService()

		if err := emailService.SendPasswordResetEmail(
			user.Email,
			user.Username,
			resetToken,
		); err != nil {
			log.Printf("Password reset email delivery failed: %v", err)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	_ = json.NewEncoder(w).Encode(map[string]string{
		"message": "if the email exists, a password reset link has been sent",
	})
}

package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/Traderong/omnivibe-api/auth"
	"github.com/Traderong/omnivibe-api/email"
)

func ResendVerification(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req auth.ResendVerificationRequest

	if err := DecodeJSON(w, r, &req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	user, verificationToken, err := auth.ResendVerificationEmail(
		r.Context(),
		req.Email,
	)
	if err != nil {
		if errors.Is(err, auth.ErrEmailAlreadyVerified) {
			http.Error(
				w,
				"email is already verified",
				http.StatusConflict,
			)
			return
		}

		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	emailService := email.NewService()

	if err := emailService.SendVerificationEmail(
		user.Email,
		user.Username,
		verificationToken,
	); err != nil {
		log.Printf("SMTP email delivery failed: %v", err)

		http.Error(
			w,
			"verification token generated but email could not be sent",
			http.StatusInternalServerError,
		)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	_ = json.NewEncoder(w).Encode(map[string]string{
		"message": "verification email sent successfully",
	})
}

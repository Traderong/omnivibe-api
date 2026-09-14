package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/Traderong/omnivibe-api/auth"
	"github.com/Traderong/omnivibe-api/email"
)

func Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req auth.RegisterRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	user, verificationToken, err := auth.RegisterUser(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	emailService := email.NewService()

	if err := emailService.SendVerificationEmail(
		user.Email,
		user.Username,
		verificationToken,
	); err != nil {
		http.Error(w, "account created but verification email could not be sent", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	_ = json.NewEncoder(w).Encode(user)
}
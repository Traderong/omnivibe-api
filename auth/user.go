package auth

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID              uuid.UUID  `json:"id"`
	Username        string     `json:"username"`
	Email           string     `json:"email"`
	DisplayName     string     `json:"display_name"`
	Bio             *string    `json:"bio,omitempty"`
	AvatarURL       *string    `json:"avatar_url,omitempty"`
	IsVerified      bool       `json:"is_verified"`
	IsActive        bool       `json:"is_active"`
	ProfilePrivate  bool       `json:"profile_private"`
	EmailVerifiedAt *time.Time `json:"email_verified_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

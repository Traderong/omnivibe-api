package auth

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var ErrJWTSecretMissing = errors.New("JWT_SECRET is not configured")

const (
	DefaultJWTIssuer   = "omnivibe-api"
	DefaultJWTAudience = "omnivibe"
)

func jwtIssuer() string {
	if value := os.Getenv("JWT_ISSUER"); value != "" {
		return value
	}

	return DefaultJWTIssuer
}

func jwtAudience() string {
	if value := os.Getenv("JWT_AUDIENCE"); value != "" {
		return value
	}

	return DefaultJWTAudience
}

func GenerateAccessToken(user *User) (string, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return "", ErrJWTSecretMissing
	}

	now := time.Now()

	claims := jwt.RegisteredClaims{
		Subject:   user.ID.String(),
		Issuer:    jwtIssuer(),
		Audience:  jwt.ClaimStrings{jwtAudience()},
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(15 * time.Minute)),
	}

	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		claims,
	)

	return token.SignedString([]byte(secret))
}

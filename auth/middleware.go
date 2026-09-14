package auth

import (
	"context"
	"errors"
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const userIDContextKey contextKey = "user_id"

var ErrMissingToken = errors.New("authentication token required")

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")

		if authHeader == "" {
			http.Error(
				w,
				"authentication required",
				http.StatusUnauthorized,
			)
			return
		}

		parts := strings.Fields(authHeader)

		if len(parts) != 2 ||
			!strings.EqualFold(parts[0], "Bearer") ||
			parts[1] == "" {
			http.Error(
				w,
				"invalid authorization header",
				http.StatusUnauthorized,
			)
			return
		}

		tokenString := parts[1]

		secret := os.Getenv("JWT_SECRET")
		if secret == "" {
			http.Error(
				w,
				"server authentication configuration error",
				http.StatusInternalServerError,
			)
			return
		}

		claims := &jwt.RegisteredClaims{}

		token, err := jwt.ParseWithClaims(
			tokenString,
			claims,
			func(token *jwt.Token) (interface{}, error) {
				if token.Method != jwt.SigningMethodHS256 {
					return nil, errors.New("unexpected signing method")
				}

				return []byte(secret), nil
			},
			jwt.WithIssuer(jwtIssuer()),
			jwt.WithAudience(jwtAudience()),
		)

		if err != nil || !token.Valid {
			http.Error(
				w,
				"invalid or expired token",
				http.StatusUnauthorized,
			)
			return
		}

		if claims.Subject == "" {
			http.Error(
				w,
				"invalid token subject",
				http.StatusUnauthorized,
			)
			return
		}

		ctx := context.WithValue(
			r.Context(),
			userIDContextKey,
			claims.Subject,
		)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func UserIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(userIDContextKey).(string)

	return userID, ok
}

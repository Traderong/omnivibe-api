package auth

import (
	"context"
	"fmt"

	"github.com/Traderong/omnivibe-api/database"
)

func CleanupRefreshTokens(ctx context.Context) error {
	_, err := database.DB.Exec(
		ctx,
		`DELETE FROM refresh_tokens
		 WHERE expires_at <= NOW()
		    OR revoked_at <= NOW() - INTERVAL '30 days'`,
	)
	if err != nil {
		return fmt.Errorf("cleanup refresh tokens: %w", err)
	}

	return nil
}

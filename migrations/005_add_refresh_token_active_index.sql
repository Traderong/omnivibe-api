CREATE INDEX IF NOT EXISTS idx_refresh_tokens_active_user
ON refresh_tokens(user_id)
WHERE revoked_at IS NULL;
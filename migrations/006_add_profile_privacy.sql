ALTER TABLE users
ADD COLUMN IF NOT EXISTS profile_private BOOLEAN NOT NULL DEFAULT FALSE;

CREATE INDEX IF NOT EXISTS idx_users_profile_private
ON users(profile_private);
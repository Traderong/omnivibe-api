CREATE TABLE IF NOT EXISTS follows (
    follower_id UUID NOT NULL,
    following_id UUID NOT NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    PRIMARY KEY (follower_id, following_id),

    CONSTRAINT fk_follows_follower
        FOREIGN KEY (follower_id)
        REFERENCES users(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_follows_following
        FOREIGN KEY (following_id)
        REFERENCES users(id)
        ON DELETE CASCADE,

    CONSTRAINT chk_follows_not_self
        CHECK (follower_id <> following_id)
);

CREATE INDEX IF NOT EXISTS idx_follows_following_id
    ON follows(following_id);

CREATE INDEX IF NOT EXISTS idx_follows_follower_id
    ON follows(follower_id);
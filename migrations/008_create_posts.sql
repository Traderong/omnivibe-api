CREATE TABLE IF NOT EXISTS posts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    user_id UUID NOT NULL,

    content TEXT NOT NULL,

    media_url TEXT,

    visibility VARCHAR(20) NOT NULL DEFAULT 'public',

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_posts_user
        FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE CASCADE,

    CONSTRAINT chk_posts_content_length
        CHECK (
            char_length(trim(content)) > 0
            AND char_length(content) <= 5000
        ),

    CONSTRAINT chk_posts_visibility
        CHECK (
            visibility IN ('public', 'followers', 'private')
        )
);

CREATE INDEX IF NOT EXISTS idx_posts_user_id_created_at
    ON posts(user_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_posts_visibility_created_at
    ON posts(visibility, created_at DESC);
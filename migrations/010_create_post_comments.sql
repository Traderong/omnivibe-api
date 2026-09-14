CREATE TABLE IF NOT EXISTS post_comments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    post_id UUID NOT NULL
        REFERENCES posts(id)
        ON DELETE CASCADE,

    user_id UUID NOT NULL
        REFERENCES users(id)
        ON DELETE CASCADE,

    content TEXT NOT NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT post_comments_content_not_empty
        CHECK (char_length(trim(content)) > 0),

    CONSTRAINT post_comments_content_max_length
        CHECK (char_length(content) <= 2000)
);

CREATE INDEX IF NOT EXISTS idx_post_comments_post_created
    ON post_comments(post_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_post_comments_user_created
    ON post_comments(user_id, created_at DESC);

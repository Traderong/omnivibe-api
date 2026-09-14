CREATE TABLE IF NOT EXISTS post_likes (
    post_id UUID NOT NULL,
    user_id UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    PRIMARY KEY (post_id, user_id),

    CONSTRAINT fk_post_likes_post
        FOREIGN KEY (post_id)
        REFERENCES posts(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_post_likes_user
        FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_post_likes_user_id
    ON post_likes(user_id);

CREATE INDEX IF NOT EXISTS idx_post_likes_post_id
    ON post_likes(post_id);


CREATE TABLE IF NOT EXISTS post_favorites (
    post_id UUID NOT NULL,
    user_id UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    PRIMARY KEY (post_id, user_id),

    CONSTRAINT fk_post_favorites_post
        FOREIGN KEY (post_id)
        REFERENCES posts(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_post_favorites_user
        FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_post_favorites_user_id
    ON post_favorites(user_id);

CREATE INDEX IF NOT EXISTS idx_post_favorites_post_id
    ON post_favorites(post_id);
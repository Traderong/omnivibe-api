CREATE TABLE IF NOT EXISTS notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    recipient_id UUID NOT NULL
        REFERENCES users(id)
        ON DELETE CASCADE,

    actor_id UUID
        REFERENCES users(id)
        ON DELETE SET NULL,

    type VARCHAR(30) NOT NULL,

    post_id UUID
        REFERENCES posts(id)
        ON DELETE CASCADE,

    comment_id UUID
        REFERENCES post_comments(id)
        ON DELETE CASCADE,

    message TEXT NOT NULL,

    is_read BOOLEAN NOT NULL DEFAULT FALSE,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT notifications_type_check
        CHECK (
            type IN (
                'follow',
                'like',
                'comment'
            )
        ),

    CONSTRAINT notifications_message_not_empty
        CHECK (char_length(trim(message)) > 0)
);

CREATE INDEX IF NOT EXISTS idx_notifications_recipient_created
    ON notifications(recipient_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_notifications_recipient_unread
    ON notifications(recipient_id, is_read, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_notifications_post
    ON notifications(post_id);

CREATE INDEX IF NOT EXISTS idx_notifications_actor
    ON notifications(actor_id);
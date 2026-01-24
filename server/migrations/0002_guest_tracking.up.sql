-- GUEST USER TRACKING
CREATE TABLE IF NOT EXISTS guests (
    user_id UUID REFERENCES users (user_id) ON DELETE CASCADE,
    added_by UUID REFERENCES users (user_id) ON DELETE SET NULL,
    added_at TIMESTAMPTZ DEFAULT now(),
    PRIMARY KEY (user_id)
);

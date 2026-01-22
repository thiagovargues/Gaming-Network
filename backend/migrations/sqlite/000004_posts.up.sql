CREATE INDEX IF NOT EXISTS idx_posts_visibility ON posts(visibility);
CREATE INDEX IF NOT EXISTS idx_post_allowed_post_id ON post_allowed(post_id);
CREATE INDEX IF NOT EXISTS idx_post_allowed_user_id ON post_allowed(user_id);

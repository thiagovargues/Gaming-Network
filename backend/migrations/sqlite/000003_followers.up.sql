CREATE INDEX IF NOT EXISTS idx_follow_requests_from ON follow_requests(from_user_id);
CREATE INDEX IF NOT EXISTS idx_follow_requests_to ON follow_requests(to_user_id);
CREATE INDEX IF NOT EXISTS idx_follows_follower ON follows(follower_id);
CREATE INDEX IF NOT EXISTS idx_follows_followee ON follows(followee_id);

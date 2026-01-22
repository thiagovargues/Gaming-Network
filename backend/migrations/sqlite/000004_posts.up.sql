CREATE TABLE IF NOT EXISTS posts (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	author_id INTEGER NOT NULL,
	content TEXT NOT NULL,
	visibility TEXT NOT NULL CHECK (visibility IN ('public', 'followers', 'selected')),
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY (author_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS post_allowed_viewers (
	post_id INTEGER NOT NULL,
	user_id INTEGER NOT NULL,
	UNIQUE(post_id, user_id),
	FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE,
	FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_posts_author_id ON posts(author_id);
CREATE INDEX IF NOT EXISTS idx_posts_visibility ON posts(visibility);
CREATE INDEX IF NOT EXISTS idx_posts_created_at ON posts(created_at);
CREATE INDEX IF NOT EXISTS idx_post_allowed_viewers_post_id ON post_allowed_viewers(post_id);
CREATE INDEX IF NOT EXISTS idx_post_allowed_viewers_user_id ON post_allowed_viewers(user_id);

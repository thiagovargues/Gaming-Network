ALTER TABLE users RENAME TO users_old;

CREATE TABLE IF NOT EXISTS users (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	email TEXT NOT NULL UNIQUE,
	password_hash TEXT,
	first_name TEXT NOT NULL,
	last_name TEXT NOT NULL,
	dob TEXT,
	avatar_path TEXT,
	nickname TEXT,
	about TEXT,
	is_public INTEGER NOT NULL DEFAULT 1,
	created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO users (id, email, password_hash, first_name, last_name, dob, avatar_path, nickname, about, is_public, created_at)
SELECT id, email, password_hash, first_name, last_name, dob, avatar_path, nickname, about, is_public, created_at
FROM users_old;

DROP TABLE users_old;

CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);

CREATE TABLE IF NOT EXISTS oauth_accounts (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	user_id INTEGER NOT NULL,
	provider TEXT NOT NULL,
	subject TEXT NOT NULL,
	email TEXT,
	created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
	UNIQUE (provider, subject),
	UNIQUE (user_id, provider),
	FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_oauth_accounts_user_id ON oauth_accounts(user_id);

PRAGMA foreign_keys = OFF;
BEGIN;

DROP TABLE IF EXISTS oauth_accounts;

ALTER TABLE users RENAME TO users_old;

CREATE TABLE IF NOT EXISTS users (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	email TEXT NOT NULL UNIQUE,
	password_hash TEXT NOT NULL,
	first_name TEXT NOT NULL,
	last_name TEXT NOT NULL,
	dob TEXT NOT NULL,
	avatar_path TEXT,
	nickname TEXT,
	about TEXT,
	is_public INTEGER NOT NULL DEFAULT 1,
	created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO users (id, email, password_hash, first_name, last_name, dob, avatar_path, nickname, about, is_public, created_at)
SELECT id,
	email,
	COALESCE(password_hash, ''),
	first_name,
	last_name,
	COALESCE(dob, ''),
	avatar_path,
	nickname,
	about,
	is_public,
	created_at
FROM users_old;

DROP TABLE users_old;

CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);

COMMIT;
PRAGMA foreign_keys = ON;

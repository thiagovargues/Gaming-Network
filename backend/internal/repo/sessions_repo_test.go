package repo

import (
	"context"
	"database/sql"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func TestCreateSessionExpirationComparable(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		t.Fatalf("pragma: %v", err)
	}

	if _, err := db.Exec(`CREATE TABLE users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		email TEXT NOT NULL UNIQUE,
		password_hash TEXT NOT NULL,
		first_name TEXT NOT NULL,
		last_name TEXT NOT NULL,
		dob TEXT NOT NULL,
		created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`); err != nil {
		t.Fatalf("create users: %v", err)
	}

	if _, err := db.Exec(`CREATE TABLE sessions (
		token TEXT PRIMARY KEY,
		user_id INTEGER NOT NULL,
		expires_at TEXT NOT NULL,
		created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	)`); err != nil {
		t.Fatalf("create sessions: %v", err)
	}

	result, err := db.Exec(
		"INSERT INTO users (email, password_hash, first_name, last_name, dob) VALUES (?, ?, ?, ?, ?)",
		"user@example.com",
		"hash",
		"Test",
		"User",
		"2000-01-01",
	)
	if err != nil {
		t.Fatalf("insert user: %v", err)
	}

	userID, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("user id: %v", err)
	}

	token, _, err := CreateSession(context.Background(), db, userID, 7*24*time.Hour)
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}

	var found string
	err = db.QueryRow(
		"SELECT token FROM sessions WHERE token = ? AND expires_at > CURRENT_TIMESTAMP",
		token,
	).Scan(&found)
	if err != nil {
		t.Fatalf("session lookup: %v", err)
	}
	if found == "" {
		t.Fatalf("expected session row to be visible")
	}
}

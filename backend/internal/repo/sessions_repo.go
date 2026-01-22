package repo

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"time"
)

func CreateSession(ctx context.Context, db *sql.DB, userID int64, ttl time.Duration) (string, time.Time, error) {
	token, err := randomToken()
	if err != nil {
		return "", time.Time{}, err
	}
	expiresAt := time.Now().Add(ttl).UTC()
	_, err = db.ExecContext(
		ctx,
		"INSERT INTO sessions (token, user_id, expires_at) VALUES (?, ?, ?)",
		token,
		userID,
		expiresAt,
	)
	if err != nil {
		return "", time.Time{}, err
	}
	return token, expiresAt, nil
}

func DeleteSession(ctx context.Context, db *sql.DB, token string) error {
	_, err := db.ExecContext(ctx, "DELETE FROM sessions WHERE token = ?", token)
	return err
}

func GetSessionUser(ctx context.Context, db *sql.DB, token string) (UserProfile, bool, error) {
	var profile UserProfile
	var avatar sql.NullString
	var nickname sql.NullString
	var about sql.NullString
	var isPublic int

	row := db.QueryRowContext(ctx, `SELECT users.id, users.email, users.first_name, users.last_name, users.dob,
		users.avatar_path, users.nickname, users.about, users.is_public, users.created_at
		FROM sessions
		JOIN users ON users.id = sessions.user_id
		WHERE sessions.token = ? AND sessions.expires_at > CURRENT_TIMESTAMP`, token)
	if err := row.Scan(
		&profile.ID,
		&profile.Email,
		&profile.FirstName,
		&profile.LastName,
		&profile.DOB,
		&avatar,
		&nickname,
		&about,
		&isPublic,
		&profile.CreatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return UserProfile{}, false, nil
		}
		return UserProfile{}, false, err
	}

	profile.Avatar = nullableStringPtr(avatar)
	profile.Nickname = nullableStringPtr(nickname)
	profile.About = nullableStringPtr(about)
	profile.IsPublic = isPublic == 1
	providers, err := ListOAuthProvidersByUserID(ctx, db, profile.ID)
	if err != nil {
		return UserProfile{}, false, err
	}
	if len(providers) > 0 {
		profile.AuthProviders = providers
	}
	return profile, true, nil
}

func randomToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

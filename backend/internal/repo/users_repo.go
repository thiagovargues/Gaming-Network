package repo

import (
	"context"
	"database/sql"
	"strings"
)

func CreateUser(ctx context.Context, db *sql.DB, email, passwordHash, firstName, lastName, dob string, avatar, nickname, about *string) (int64, error) {
	result, err := db.ExecContext(
		ctx,
		`INSERT INTO users (email, password_hash, first_name, last_name, dob, avatar_path, nickname, about, is_public)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, 1)`,
		email,
		passwordHash,
		firstName,
		lastName,
		dob,
		nullableString(avatar),
		nullableString(nickname),
		nullableString(about),
	)
	if err != nil {
		return 0, err
	}
	id, _ := result.LastInsertId()
	return id, nil
}

func GetUserByEmail(ctx context.Context, db *sql.DB, email string) (int64, string, error) {
	var id int64
	var hash string
	row := db.QueryRowContext(ctx, "SELECT id, password_hash FROM users WHERE email = ?", email)
	if err := row.Scan(&id, &hash); err != nil {
		return 0, "", err
	}
	return id, hash, nil
}

func GetUserByID(ctx context.Context, db *sql.DB, id int64) (UserProfile, bool, error) {
	var profile UserProfile
	var avatar sql.NullString
	var nickname sql.NullString
	var about sql.NullString
	var isPublic int

	row := db.QueryRowContext(ctx, `SELECT id, email, first_name, last_name, dob, avatar_path, nickname, about, is_public, created_at
		FROM users WHERE id = ?`, id)
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
	return profile, true, nil
}

func UpdateMe(ctx context.Context, db *sql.DB, userID int64, isPublic *bool, avatar, nickname, about *string) (UserProfile, error) {
	if isPublic != nil {
		value := 0
		if *isPublic {
			value = 1
		}
		if _, err := db.ExecContext(ctx, "UPDATE users SET is_public = ? WHERE id = ?", value, userID); err != nil {
			return UserProfile{}, err
		}
	}

	if avatar != nil {
		if _, err := db.ExecContext(ctx, "UPDATE users SET avatar_path = ? WHERE id = ?", nullableString(avatar), userID); err != nil {
			return UserProfile{}, err
		}
	}
	if nickname != nil {
		if _, err := db.ExecContext(ctx, "UPDATE users SET nickname = ? WHERE id = ?", nullableString(nickname), userID); err != nil {
			return UserProfile{}, err
		}
	}
	if about != nil {
		if _, err := db.ExecContext(ctx, "UPDATE users SET about = ? WHERE id = ?", nullableString(about), userID); err != nil {
			return UserProfile{}, err
		}
	}

	profile, _, err := GetUserByID(ctx, db, userID)
	return profile, err
}

func IsUserPublic(ctx context.Context, db *sql.DB, userID int64) (bool, bool, error) {
	var isPublic int
	row := db.QueryRowContext(ctx, "SELECT is_public FROM users WHERE id = ?", userID)
	if err := row.Scan(&isPublic); err != nil {
		if err == sql.ErrNoRows {
			return false, false, nil
		}
		return false, false, err
	}
	return isPublic == 1, true, nil
}

func nullableString(value *string) sql.NullString {
	if value == nil {
		return sql.NullString{}
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: trimmed, Valid: true}
}

func nullableStringPtr(value sql.NullString) *string {
	if !value.Valid {
		return nil
	}
	val := strings.TrimSpace(value.String)
	if val == "" {
		return nil
	}
	return &val
}

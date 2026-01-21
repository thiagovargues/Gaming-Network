package repo

import (
	"context"
	"database/sql"
)

func GetUserByID(ctx context.Context, db *sql.DB, userID int64) (UserProfile, bool, error) {
	var profile UserProfile
	var nickname sql.NullString
	var aboutMe sql.NullString
	var avatarPath sql.NullString
	var isPrivate int

	row := db.QueryRowContext(
		ctx,
		`SELECT id, email, first_name, last_name, birth_date, nickname, about_me, avatar_path,
			is_private, created_at
		FROM users
		WHERE id = ?`,
		userID,
	)
	if err := row.Scan(
		&profile.ID,
		&profile.Email,
		&profile.FirstName,
		&profile.LastName,
		&profile.BirthDate,
		&nickname,
		&aboutMe,
		&avatarPath,
		&isPrivate,
		&profile.CreatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return UserProfile{}, false, nil
		}
		return UserProfile{}, false, err
	}

	profile.IsPrivate = isPrivate != 0
	profile.Nickname = nullableStringPtr(nickname)
	profile.AboutMe = nullableStringPtr(aboutMe)
	profile.AvatarPath = nullableStringPtr(avatarPath)

	return profile, true, nil
}

func UpdateUserPrivacy(ctx context.Context, db *sql.DB, userID int64, isPrivate bool) (UserProfile, bool, error) {
	value := 0
	if isPrivate {
		value = 1
	}

	if _, err := db.ExecContext(
		ctx,
		"UPDATE users SET is_private = ? WHERE id = ?",
		value,
		userID,
	); err != nil {
		return UserProfile{}, false, err
	}

	profile, found, err := GetUserByID(ctx, db, userID)
	if err != nil {
		return UserProfile{}, false, err
	}
	return profile, found, nil
}

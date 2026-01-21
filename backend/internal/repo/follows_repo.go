package repo

import (
	"context"
	"database/sql"
)

type UserProfile struct {
	ID         int64
	Email      string
	FirstName  string
	LastName   string
	BirthDate  string
	Nickname   *string
	AboutMe    *string
	AvatarPath *string
	IsPrivate  bool
	CreatedAt  string
}

func GetUserPrivacy(ctx context.Context, db *sql.DB, userID int64) (bool, bool, error) {
	var isPrivate int
	err := db.QueryRowContext(ctx, "SELECT is_private FROM users WHERE id = ?", userID).Scan(&isPrivate)
	if err == sql.ErrNoRows {
		return false, false, nil
	}
	if err != nil {
		return false, false, err
	}
	return isPrivate != 0, true, nil
}

func IsFollowing(ctx context.Context, db *sql.DB, followerID, followedID int64) (bool, error) {
	var exists int
	err := db.QueryRowContext(
		ctx,
		"SELECT 1 FROM follows WHERE follower_id = ? AND followed_id = ? LIMIT 1",
		followerID,
		followedID,
	).Scan(&exists)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func CreateFollowRequest(ctx context.Context, db *sql.DB, requesterID, targetID int64) (bool, error) {
	result, err := db.ExecContext(
		ctx,
		"INSERT OR IGNORE INTO follow_requests (requester_id, target_id) VALUES (?, ?)",
		requesterID,
		targetID,
	)
	if err != nil {
		return false, err
	}
	rows, _ := result.RowsAffected()
	return rows > 0, nil
}

func CreateFollow(ctx context.Context, db *sql.DB, followerID, followedID int64) (bool, error) {
	result, err := db.ExecContext(
		ctx,
		"INSERT OR IGNORE INTO follows (follower_id, followed_id) VALUES (?, ?)",
		followerID,
		followedID,
	)
	if err != nil {
		return false, err
	}
	rows, _ := result.RowsAffected()
	return rows > 0, nil
}

func DeleteFollowRequest(ctx context.Context, db *sql.DB, requesterID, targetID int64) (bool, error) {
	result, err := db.ExecContext(
		ctx,
		"DELETE FROM follow_requests WHERE requester_id = ? AND target_id = ?",
		requesterID,
		targetID,
	)
	if err != nil {
		return false, err
	}
	rows, _ := result.RowsAffected()
	return rows > 0, nil
}

func DeleteFollow(ctx context.Context, db *sql.DB, followerID, followedID int64) (bool, error) {
	result, err := db.ExecContext(
		ctx,
		"DELETE FROM follows WHERE follower_id = ? AND followed_id = ?",
		followerID,
		followedID,
	)
	if err != nil {
		return false, err
	}
	rows, _ := result.RowsAffected()
	return rows > 0, nil
}

func AcceptFollowRequest(ctx context.Context, db *sql.DB, requesterID, targetID int64) (bool, error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return false, err
	}

	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	result, err := tx.ExecContext(
		ctx,
		"DELETE FROM follow_requests WHERE requester_id = ? AND target_id = ?",
		requesterID,
		targetID,
	)
	if err != nil {
		return false, err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return false, nil
	}

	if _, err := tx.ExecContext(
		ctx,
		"INSERT OR IGNORE INTO follows (follower_id, followed_id) VALUES (?, ?)",
		requesterID,
		targetID,
	); err != nil {
		return false, err
	}

	if err := tx.Commit(); err != nil {
		return false, err
	}
	committed = true
	return true, nil
}

func ListFollowers(ctx context.Context, db *sql.DB, userID int64) ([]UserProfile, error) {
	query := `SELECT users.id, users.email, users.first_name, users.last_name, users.birth_date,
		users.nickname, users.about_me, users.avatar_path, users.is_private, users.created_at
		FROM follows
		JOIN users ON users.id = follows.follower_id
		WHERE follows.followed_id = ?
		ORDER BY follows.created_at DESC`
	return queryUserProfiles(ctx, db, query, userID)
}

func ListFollowing(ctx context.Context, db *sql.DB, userID int64) ([]UserProfile, error) {
	query := `SELECT users.id, users.email, users.first_name, users.last_name, users.birth_date,
		users.nickname, users.about_me, users.avatar_path, users.is_private, users.created_at
		FROM follows
		JOIN users ON users.id = follows.followed_id
		WHERE follows.follower_id = ?
		ORDER BY follows.created_at DESC`
	return queryUserProfiles(ctx, db, query, userID)
}

func queryUserProfiles(ctx context.Context, db *sql.DB, query string, args ...any) ([]UserProfile, error) {
	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var profiles []UserProfile
	for rows.Next() {
		var profile UserProfile
		var nickname sql.NullString
		var aboutMe sql.NullString
		var avatarPath sql.NullString
		var isPrivate int

		if err := rows.Scan(
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
			return nil, err
		}

		profile.IsPrivate = isPrivate != 0
		profile.Nickname = nullableStringPtr(nickname)
		profile.AboutMe = nullableStringPtr(aboutMe)
		profile.AvatarPath = nullableStringPtr(avatarPath)
		profiles = append(profiles, profile)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return profiles, nil
}

func nullableStringPtr(value sql.NullString) *string {
	if !value.Valid {
		return nil
	}
	val := value.String
	return &val
}

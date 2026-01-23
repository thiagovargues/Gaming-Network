package repo

import (
	"context"
	"database/sql"
)

func CreateFollowRequest(ctx context.Context, db *sql.DB, fromID, toID int64) (int64, error) {
	result, err := db.ExecContext(
		ctx,
		"INSERT INTO follow_requests (from_user_id, to_user_id, status) VALUES (?, ?, 'pending')",
		fromID,
		toID,
	)
	if err != nil {
		return 0, err
	}
	id, _ := result.LastInsertId()
	return id, nil
}

func UpdateFollowRequestStatus(ctx context.Context, db *sql.DB, requestID int64, status string) (int64, int64, bool, error) {
	var fromID int64
	var toID int64
	row := db.QueryRowContext(ctx, "SELECT from_user_id, to_user_id FROM follow_requests WHERE id = ? AND status = 'pending'", requestID)
	if err := row.Scan(&fromID, &toID); err != nil {
		if err == sql.ErrNoRows {
			return 0, 0, false, nil
		}
		return 0, 0, false, err
	}

	if _, err := db.ExecContext(ctx, "UPDATE follow_requests SET status = ? WHERE id = ?", status, requestID); err != nil {
		return 0, 0, false, err
	}
	return fromID, toID, true, nil
}

func CreateFollow(ctx context.Context, db *sql.DB, followerID, followeeID int64) error {
	_, err := db.ExecContext(ctx, "INSERT OR IGNORE INTO follows (follower_id, followee_id) VALUES (?, ?)", followerID, followeeID)
	return err
}

func DeleteFollow(ctx context.Context, db *sql.DB, followerID, followeeID int64) error {
	_, err := db.ExecContext(ctx, "DELETE FROM follows WHERE follower_id = ? AND followee_id = ?", followerID, followeeID)
	return err
}

func IsFollowing(ctx context.Context, db *sql.DB, followerID, followeeID int64) (bool, error) {
	var exists int
	row := db.QueryRowContext(ctx, "SELECT 1 FROM follows WHERE follower_id = ? AND followee_id = ?", followerID, followeeID)
	if err := row.Scan(&exists); err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func ListFollowers(ctx context.Context, db *sql.DB, userID int64) ([]UserProfile, error) {
	query := `SELECT users.id, users.email, users.first_name, users.last_name, users.dob, users.avatar_path,
		users.nickname, users.about, users.sex, users.age, users.show_first_name, users.show_last_name, users.show_age, users.show_sex, users.show_nickname, users.show_about, users.is_public, users.created_at
		FROM follows
		JOIN users ON users.id = follows.follower_id
		WHERE follows.followee_id = ?
		ORDER BY follows.created_at DESC`
	return queryProfiles(ctx, db, query, userID)
}

func ListFollowing(ctx context.Context, db *sql.DB, userID int64) ([]UserProfile, error) {
	query := `SELECT users.id, users.email, users.first_name, users.last_name, users.dob, users.avatar_path,
		users.nickname, users.about, users.sex, users.age, users.show_first_name, users.show_last_name, users.show_age, users.show_sex, users.show_nickname, users.show_about, users.is_public, users.created_at
		FROM follows
		JOIN users ON users.id = follows.followee_id
		WHERE follows.follower_id = ?
		ORDER BY follows.created_at DESC`
	return queryProfiles(ctx, db, query, userID)
}

func ListIncomingFollowRequests(ctx context.Context, db *sql.DB, userID int64) ([]FollowRequest, error) {
	rows, err := db.QueryContext(ctx, `SELECT id, from_user_id, to_user_id, status, created_at
		FROM follow_requests WHERE to_user_id = ? AND status = 'pending' ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []FollowRequest
	for rows.Next() {
		var fr FollowRequest
		if err := rows.Scan(&fr.ID, &fr.FromUserID, &fr.ToUserID, &fr.Status, &fr.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, fr)
	}
	return list, rows.Err()
}

func ListOutgoingFollowRequests(ctx context.Context, db *sql.DB, userID int64) ([]FollowRequest, error) {
	rows, err := db.QueryContext(ctx, `SELECT id, from_user_id, to_user_id, status, created_at
		FROM follow_requests WHERE from_user_id = ? AND status = 'pending' ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []FollowRequest
	for rows.Next() {
		var fr FollowRequest
		if err := rows.Scan(&fr.ID, &fr.FromUserID, &fr.ToUserID, &fr.Status, &fr.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, fr)
	}
	return list, rows.Err()
}

type FollowRequest struct {
	ID        int64  `json:"id"`
	FromUserID int64 `json:"from_user_id"`
	ToUserID  int64  `json:"to_user_id"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
}

func queryProfiles(ctx context.Context, db *sql.DB, query string, args ...any) ([]UserProfile, error) {
	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var profiles []UserProfile
	for rows.Next() {
		var profile UserProfile
		var avatar sql.NullString
		var nickname sql.NullString
		var about sql.NullString
		var sex sql.NullString
		var age sql.NullInt64
		var isPublic int
		var showFirst int
		var showLast int
		var showAge int
		var showSex int
		var showNickname int
		var showAbout int
		if err := rows.Scan(&profile.ID, &profile.Email, &profile.FirstName, &profile.LastName, &profile.DOB,
			&avatar, &nickname, &about, &sex, &age, &showFirst, &showLast, &showAge, &showSex, &showNickname, &showAbout, &isPublic, &profile.CreatedAt); err != nil {
			return nil, err
		}
		profile.Avatar = nullableStringPtr(avatar)
		profile.Nickname = nullableStringPtr(nickname)
		profile.About = nullableStringPtr(about)
		profile.Sex = nullableStringPtr(sex)
		if age.Valid {
			v := int(age.Int64)
			profile.Age = &v
		}
		profile.ShowFirstName = showFirst == 1
		profile.ShowLastName = showLast == 1
		profile.ShowAge = showAge == 1
		profile.ShowSex = showSex == 1
		profile.ShowNickname = showNickname == 1
		profile.ShowAbout = showAbout == 1
		profile.IsPublic = isPublic == 1
		profiles = append(profiles, profile)
	}
	return profiles, rows.Err()
}

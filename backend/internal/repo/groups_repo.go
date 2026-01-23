package repo

import (
	"context"
	"database/sql"
)

func CreateGroup(ctx context.Context, db *sql.DB, creatorID int64, title, description string) (Group, error) {
	result, err := db.ExecContext(ctx, "INSERT INTO groups (creator_id, title, description) VALUES (?, ?, ?)", creatorID, title, description)
	if err != nil {
		return Group{}, err
	}
	id, _ := result.LastInsertId()
	_, err = db.ExecContext(ctx, "INSERT INTO group_members (group_id, user_id, role) VALUES (?, ?, 'creator')", id, creatorID)
	if err != nil {
		return Group{}, err
	}
	return GetGroup(ctx, db, id)
}

func GetGroup(ctx context.Context, db *sql.DB, groupID int64) (Group, error) {
	var group Group
	row := db.QueryRowContext(ctx, "SELECT id, creator_id, title, description, created_at FROM groups WHERE id = ?", groupID)
	if err := row.Scan(&group.ID, &group.CreatorID, &group.Title, &group.Description, &group.CreatedAt); err != nil {
		return Group{}, err
	}
	return group, nil
}

func ListGroups(ctx context.Context, db *sql.DB, query string, limit, offset int) ([]Group, error) {
	rows, err := db.QueryContext(ctx, "SELECT id, creator_id, title, description, created_at FROM groups WHERE title LIKE ? ORDER BY created_at DESC LIMIT ? OFFSET ?", "%"+query+"%", limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var groups []Group
	for rows.Next() {
		var group Group
		if err := rows.Scan(&group.ID, &group.CreatorID, &group.Title, &group.Description, &group.CreatedAt); err != nil {
			return nil, err
		}
		groups = append(groups, group)
	}
	return groups, rows.Err()
}

func IsGroupMember(ctx context.Context, db *sql.DB, groupID, userID int64) (bool, error) {
	var exists int
	row := db.QueryRowContext(ctx, "SELECT 1 FROM group_members WHERE group_id = ? AND user_id = ?", groupID, userID)
	if err := row.Scan(&exists); err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func AddGroupMember(ctx context.Context, db *sql.DB, groupID, userID int64, role string) error {
	_, err := db.ExecContext(ctx, "INSERT OR IGNORE INTO group_members (group_id, user_id, role) VALUES (?, ?, ?)", groupID, userID, role)
	return err
}

func ListGroupMembers(ctx context.Context, db *sql.DB, groupID int64) ([]UserProfile, error) {
	query := `SELECT users.id, users.email, users.first_name, users.last_name, users.dob, users.avatar_path,
		users.nickname, users.about, users.sex, users.age, users.show_first_name, users.show_last_name, users.show_age, users.show_sex, users.show_nickname, users.show_about, users.is_public, users.created_at
		FROM group_members
		JOIN users ON users.id = group_members.user_id
		WHERE group_members.group_id = ?
		ORDER BY group_members.created_at ASC`
	return queryProfiles(ctx, db, query, groupID)
}

func CreateGroupInvite(ctx context.Context, db *sql.DB, groupID, fromID, toID int64) (int64, error) {
	result, err := db.ExecContext(ctx, "INSERT INTO group_invites (group_id, from_user_id, to_user_id, status) VALUES (?, ?, ?, 'pending')", groupID, fromID, toID)
	if err != nil {
		return 0, err
	}
	id, _ := result.LastInsertId()
	return id, nil
}

func UpdateGroupInviteStatus(ctx context.Context, db *sql.DB, inviteID int64, status string) (int64, int64, bool, error) {
	var groupID int64
	var toID int64
	row := db.QueryRowContext(ctx, "SELECT group_id, to_user_id FROM group_invites WHERE id = ? AND status = 'pending'", inviteID)
	if err := row.Scan(&groupID, &toID); err != nil {
		if err == sql.ErrNoRows {
			return 0, 0, false, nil
		}
		return 0, 0, false, err
	}
	if _, err := db.ExecContext(ctx, "UPDATE group_invites SET status = ? WHERE id = ?", status, inviteID); err != nil {
		return 0, 0, false, err
	}
	return groupID, toID, true, nil
}

func CreateJoinRequest(ctx context.Context, db *sql.DB, groupID, userID int64) (int64, error) {
	result, err := db.ExecContext(ctx, "INSERT INTO group_join_requests (group_id, user_id, status) VALUES (?, ?, 'pending')", groupID, userID)
	if err != nil {
		return 0, err
	}
	id, _ := result.LastInsertId()
	return id, nil
}

func UpdateJoinRequestStatus(ctx context.Context, db *sql.DB, requestID int64, status string) (int64, int64, bool, error) {
	var groupID int64
	var userID int64
	row := db.QueryRowContext(ctx, "SELECT group_id, user_id FROM group_join_requests WHERE id = ? AND status = 'pending'", requestID)
	if err := row.Scan(&groupID, &userID); err != nil {
		if err == sql.ErrNoRows {
			return 0, 0, false, nil
		}
		return 0, 0, false, err
	}
	if _, err := db.ExecContext(ctx, "UPDATE group_join_requests SET status = ? WHERE id = ?", status, requestID); err != nil {
		return 0, 0, false, err
	}
	return groupID, userID, true, nil
}

func GroupCreator(ctx context.Context, db *sql.DB, groupID int64) (int64, error) {
	var creatorID int64
	row := db.QueryRowContext(ctx, "SELECT creator_id FROM groups WHERE id = ?", groupID)
	if err := row.Scan(&creatorID); err != nil {
		return 0, err
	}
	return creatorID, nil
}

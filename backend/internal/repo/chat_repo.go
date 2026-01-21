package repo

import (
	"context"
	"database/sql"
)

func SaveDM(ctx context.Context, db *sql.DB, fromID, toID int64, text string) (Message, error) {
	result, err := db.ExecContext(ctx, "INSERT INTO dm_messages (from_user_id, to_user_id, text) VALUES (?, ?, ?)", fromID, toID, text)
	if err != nil {
		return Message{}, err
	}
	id, _ := result.LastInsertId()
	return GetDM(ctx, db, id)
}

func GetDM(ctx context.Context, db *sql.DB, id int64) (Message, error) {
	var msg Message
	row := db.QueryRowContext(ctx, "SELECT id, from_user_id, to_user_id, text, created_at FROM dm_messages WHERE id = ?", id)
	if err := row.Scan(&msg.ID, &msg.FromID, &msg.ToID, &msg.Text, &msg.CreatedAt); err != nil {
		return Message{}, err
	}
	return msg, nil
}

func ListDMs(ctx context.Context, db *sql.DB, userA, userB int64, limit, offset int) ([]Message, error) {
	rows, err := db.QueryContext(ctx, `SELECT id, from_user_id, to_user_id, text, created_at
		FROM dm_messages
		WHERE (from_user_id = ? AND to_user_id = ?) OR (from_user_id = ? AND to_user_id = ?)
		ORDER BY created_at DESC LIMIT ? OFFSET ?`, userA, userB, userB, userA, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var msgs []Message
	for rows.Next() {
		var msg Message
		if err := rows.Scan(&msg.ID, &msg.FromID, &msg.ToID, &msg.Text, &msg.CreatedAt); err != nil {
			return nil, err
		}
		msgs = append(msgs, msg)
	}
	return msgs, rows.Err()
}

func SaveGroupMessage(ctx context.Context, db *sql.DB, groupID, fromID int64, text string) (Message, error) {
	result, err := db.ExecContext(ctx, "INSERT INTO group_messages (group_id, from_user_id, text) VALUES (?, ?, ?)", groupID, fromID, text)
	if err != nil {
		return Message{}, err
	}
	id, _ := result.LastInsertId()
	return GetGroupMessage(ctx, db, id)
}

func GetGroupMessage(ctx context.Context, db *sql.DB, id int64) (Message, error) {
	var msg Message
	row := db.QueryRowContext(ctx, "SELECT id, group_id, from_user_id, text, created_at FROM group_messages WHERE id = ?", id)
	if err := row.Scan(&msg.ID, &msg.GroupID, &msg.FromID, &msg.Text, &msg.CreatedAt); err != nil {
		return Message{}, err
	}
	return msg, nil
}

func ListGroupMessages(ctx context.Context, db *sql.DB, groupID int64, limit, offset int) ([]Message, error) {
	rows, err := db.QueryContext(ctx, `SELECT id, group_id, from_user_id, text, created_at
		FROM group_messages WHERE group_id = ? ORDER BY created_at DESC LIMIT ? OFFSET ?`, groupID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var msgs []Message
	for rows.Next() {
		var msg Message
		if err := rows.Scan(&msg.ID, &msg.GroupID, &msg.FromID, &msg.Text, &msg.CreatedAt); err != nil {
			return nil, err
		}
		msgs = append(msgs, msg)
	}
	return msgs, rows.Err()
}

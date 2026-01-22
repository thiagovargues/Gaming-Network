package repo

import (
	"context"
	"database/sql"
)

func CreateNotification(ctx context.Context, db *sql.DB, userID int64, ntype string, payload string) error {
	_, err := db.ExecContext(ctx, "INSERT INTO notifications (user_id, type, payload_json, is_read) VALUES (?, ?, ?, 0)", userID, ntype, payload)
	return err
}

func ListNotifications(ctx context.Context, db *sql.DB, userID int64, limit, offset int) ([]Notification, error) {
	rows, err := db.QueryContext(ctx, "SELECT id, user_id, type, payload_json, is_read, created_at FROM notifications WHERE user_id = ? ORDER BY created_at DESC LIMIT ? OFFSET ?", userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []Notification
	for rows.Next() {
		var n Notification
		var isRead int
		if err := rows.Scan(&n.ID, &n.UserID, &n.Type, &n.Payload, &isRead, &n.CreatedAt); err != nil {
			return nil, err
		}
		n.IsRead = isRead == 1
		list = append(list, n)
	}
	return list, rows.Err()
}

func MarkNotificationRead(ctx context.Context, db *sql.DB, id int64, userID int64) error {
	_, err := db.ExecContext(ctx, "UPDATE notifications SET is_read = 1 WHERE id = ? AND user_id = ?", id, userID)
	return err
}

func MarkAllNotificationsRead(ctx context.Context, db *sql.DB, userID int64) error {
	_, err := db.ExecContext(ctx, "UPDATE notifications SET is_read = 1 WHERE user_id = ?", userID)
	return err
}

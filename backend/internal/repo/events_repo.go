package repo

import (
	"context"
	"database/sql"
)

func CreateEvent(ctx context.Context, db *sql.DB, groupID, creatorID int64, title, description, datetime string) (int64, error) {
	result, err := db.ExecContext(ctx, "INSERT INTO events (group_id, creator_id, title, description, datetime) VALUES (?, ?, ?, ?, ?)", groupID, creatorID, title, description, datetime)
	if err != nil {
		return 0, err
	}
	id, _ := result.LastInsertId()
	return id, nil
}

func ListEvents(ctx context.Context, db *sql.DB, groupID int64) ([]Event, error) {
	rows, err := db.QueryContext(ctx, "SELECT id, group_id, creator_id, title, description, datetime, created_at FROM events WHERE group_id = ? ORDER BY datetime ASC", groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var events []Event
	for rows.Next() {
		var ev Event
		if err := rows.Scan(&ev.ID, &ev.GroupID, &ev.CreatorID, &ev.Title, &ev.Description, &ev.Datetime, &ev.CreatedAt); err != nil {
			return nil, err
		}
		events = append(events, ev)
	}
	return events, rows.Err()
}

func RespondEvent(ctx context.Context, db *sql.DB, eventID, userID int64, status string) error {
	_, err := db.ExecContext(ctx, "INSERT OR REPLACE INTO event_responses (event_id, user_id, status) VALUES (?, ?, ?)", eventID, userID, status)
	return err
}

type Event struct {
	ID        int64  `json:"id"`
	GroupID   int64  `json:"group_id"`
	CreatorID int64  `json:"creator_id"`
	Title     string `json:"title"`
	Description string `json:"description"`
	Datetime  string `json:"datetime"`
	CreatedAt string `json:"created_at"`
}

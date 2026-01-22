package repo

import (
	"context"
	"database/sql"
)

func CreateMedia(ctx context.Context, db *sql.DB, ownerID int64, path, mime string) (int64, error) {
	result, err := db.ExecContext(ctx, "INSERT INTO media (owner_user_id, path, mime) VALUES (?, ?, ?)", ownerID, path, mime)
	if err != nil {
		return 0, err
	}
	id, _ := result.LastInsertId()
	return id, nil
}

package repo

import (
	"context"
	"database/sql"
)

func CreateComment(ctx context.Context, db *sql.DB, postID, userID int64, text string, mediaPath *string) (Comment, error) {
	result, err := db.ExecContext(ctx,
		"INSERT INTO comments (post_id, user_id, text, media_path) VALUES (?, ?, ?, ?)",
		postID,
		userID,
		text,
		nullableString(mediaPath),
	)
	if err != nil {
		return Comment{}, err
	}
	id, _ := result.LastInsertId()
	return GetCommentByID(ctx, db, id)
}

func GetCommentByID(ctx context.Context, db *sql.DB, id int64) (Comment, error) {
	var comment Comment
	var media sql.NullString
	row := db.QueryRowContext(ctx, "SELECT id, post_id, user_id, text, media_path, created_at FROM comments WHERE id = ?", id)
	if err := row.Scan(&comment.ID, &comment.PostID, &comment.UserID, &comment.Text, &media, &comment.CreatedAt); err != nil {
		return Comment{}, err
	}
	comment.MediaPath = nullableStringPtr(media)
	return comment, nil
}

func ListComments(ctx context.Context, db *sql.DB, postID int64, limit, offset int) ([]Comment, error) {
	rows, err := db.QueryContext(ctx, "SELECT id, post_id, user_id, text, media_path, created_at FROM comments WHERE post_id = ? ORDER BY created_at ASC LIMIT ? OFFSET ?", postID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var comments []Comment
	for rows.Next() {
		var comment Comment
		var media sql.NullString
		if err := rows.Scan(&comment.ID, &comment.PostID, &comment.UserID, &comment.Text, &media, &comment.CreatedAt); err != nil {
			return nil, err
		}
		comment.MediaPath = nullableStringPtr(media)
		comments = append(comments, comment)
	}
	return comments, rows.Err()
}

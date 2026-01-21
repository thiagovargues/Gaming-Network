package db

import (
	"context"
	"database/sql"
)

type Comment struct {
	ID        int64
	PostID    int64
	AuthorID  int64
	Content   string
	CreatedAt string
}

func CanUserViewPost(ctx context.Context, db *sql.DB, viewerID, postID int64) (bool, error) {
	query := `SELECT 1
		FROM posts
		LEFT JOIN follows
			ON follows.followed_id = posts.author_id AND follows.follower_id = ?
		LEFT JOIN post_allowed_viewers
			ON post_allowed_viewers.post_id = posts.id AND post_allowed_viewers.user_id = ?
		WHERE posts.id = ?
			AND (
				posts.author_id = ?
				OR posts.visibility = 'public'
				OR (posts.visibility = 'followers' AND follows.follower_id IS NOT NULL)
				OR (posts.visibility = 'selected' AND post_allowed_viewers.user_id IS NOT NULL)
			)
		LIMIT 1`

	var exists int
	err := db.QueryRowContext(ctx, query, viewerID, viewerID, postID, viewerID).Scan(&exists)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func CreateComment(ctx context.Context, db *sql.DB, postID, authorID int64, content string) (Comment, error) {
	result, err := db.ExecContext(
		ctx,
		"INSERT INTO comments (post_id, author_id, content) VALUES (?, ?, ?)",
		postID,
		authorID,
		content,
	)
	if err != nil {
		return Comment{}, err
	}

	commentID, err := result.LastInsertId()
	if err != nil {
		return Comment{}, err
	}

	var createdAt string
	if err := db.QueryRowContext(
		ctx,
		"SELECT created_at FROM comments WHERE id = ?",
		commentID,
	).Scan(&createdAt); err != nil {
		return Comment{}, err
	}

	return Comment{
		ID:        commentID,
		PostID:    postID,
		AuthorID:  authorID,
		Content:   content,
		CreatedAt: createdAt,
	}, nil
}

func ListComments(ctx context.Context, db *sql.DB, postID int64, limit, offset int) ([]Comment, error) {
	query := `SELECT id, post_id, author_id, content, created_at
		FROM comments
		WHERE post_id = ?
		ORDER BY created_at ASC
		LIMIT ? OFFSET ?`

	rows, err := db.QueryContext(ctx, query, postID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []Comment
	for rows.Next() {
		var comment Comment
		if err := rows.Scan(
			&comment.ID,
			&comment.PostID,
			&comment.AuthorID,
			&comment.Content,
			&comment.CreatedAt,
		); err != nil {
			return nil, err
		}
		comments = append(comments, comment)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return comments, nil
}

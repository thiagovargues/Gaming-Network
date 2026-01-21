package repo

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

type Post struct {
	ID         int64
	AuthorID   int64
	Content    string
	Visibility string
	CreatedAt  string
}

func CreatePost(ctx context.Context, db *sql.DB, authorID int64, content, visibility string, allowedUserIDs []int64) (Post, error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return Post{}, err
	}

	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	result, err := tx.ExecContext(
		ctx,
		"INSERT INTO posts (author_id, content, visibility) VALUES (?, ?, ?)",
		authorID,
		content,
		visibility,
	)
	if err != nil {
		return Post{}, err
	}

	postID, err := result.LastInsertId()
	if err != nil {
		return Post{}, err
	}

	if visibility == "selected" && len(allowedUserIDs) > 0 {
		stmt, err := tx.PrepareContext(
			ctx,
			"INSERT OR IGNORE INTO post_allowed_viewers (post_id, user_id) VALUES (?, ?)",
		)
		if err != nil {
			return Post{}, err
		}
		defer stmt.Close()

		for _, userID := range allowedUserIDs {
			if userID <= 0 {
				continue
			}
			if _, err := stmt.ExecContext(ctx, postID, userID); err != nil {
				return Post{}, err
			}
		}
	}

	var createdAt string
	if err := tx.QueryRowContext(ctx, "SELECT created_at FROM posts WHERE id = ?", postID).Scan(&createdAt); err != nil {
		return Post{}, err
	}

	if err := tx.Commit(); err != nil {
		return Post{}, err
	}
	committed = true

	return Post{
		ID:         postID,
		AuthorID:   authorID,
		Content:    content,
		Visibility: visibility,
		CreatedAt:  createdAt,
	}, nil
}

func AreAllowedViewersFollowers(ctx context.Context, db *sql.DB, authorID int64, allowedUserIDs []int64) (bool, error) {
	if len(allowedUserIDs) == 0 {
		return true, nil
	}

	placeholders := make([]string, len(allowedUserIDs))
	args := make([]any, 0, len(allowedUserIDs)+1)
	args = append(args, authorID)
	for i, userID := range allowedUserIDs {
		placeholders[i] = "?"
		args = append(args, userID)
	}

	query := fmt.Sprintf(
		"SELECT COUNT(DISTINCT follower_id) FROM follows WHERE followed_id = ? AND follower_id IN (%s)",
		strings.Join(placeholders, ","),
	)

	var count int
	if err := db.QueryRowContext(ctx, query, args...).Scan(&count); err != nil {
		return false, err
	}

	return count == len(allowedUserIDs), nil
}

func ListFeed(ctx context.Context, db *sql.DB, userID int64, limit, offset int) ([]Post, error) {
	query := `SELECT posts.id, posts.author_id, posts.content, posts.visibility, posts.created_at
		FROM posts
		LEFT JOIN follows
			ON follows.followed_id = posts.author_id AND follows.follower_id = ?
		LEFT JOIN post_allowed_viewers
			ON post_allowed_viewers.post_id = posts.id AND post_allowed_viewers.user_id = ?
		WHERE posts.author_id = ?
			OR posts.visibility = 'public'
			OR (posts.visibility = 'followers' AND follows.follower_id IS NOT NULL)
			OR (posts.visibility = 'selected' AND post_allowed_viewers.user_id IS NOT NULL)
		ORDER BY posts.created_at DESC
		LIMIT ? OFFSET ?`

	rows, err := db.QueryContext(ctx, query, userID, userID, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var post Post
		if err := rows.Scan(
			&post.ID,
			&post.AuthorID,
			&post.Content,
			&post.Visibility,
			&post.CreatedAt,
		); err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return posts, nil
}

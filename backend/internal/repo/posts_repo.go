package repo

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

func CreatePost(ctx context.Context, db *sql.DB, userID int64, text string, visibility string, mediaPath *string, allowedIDs []int64, groupID *int64) (Post, error) {
	result, err := db.ExecContext(
		ctx,
		"INSERT INTO posts (user_id, group_id, text, visibility, media_path) VALUES (?, ?, ?, ?, ?)",
		userID,
		groupID,
		text,
		visibility,
		nullableString(mediaPath),
	)
	if err != nil {
		return Post{}, err
	}
	postID, _ := result.LastInsertId()

	if visibility == "private" && len(allowedIDs) > 0 {
		stmt, err := db.PrepareContext(ctx, "INSERT OR IGNORE INTO post_allowed (user_id, post_id) VALUES (?, ?)")
		if err != nil {
			return Post{}, err
		}
		defer stmt.Close()
		for _, id := range allowedIDs {
			if id <= 0 {
				continue
			}
			if _, err := stmt.ExecContext(ctx, id, postID); err != nil {
				return Post{}, err
			}
		}
	}

	return GetPostByID(ctx, db, postID)
}

func GetPostByID(ctx context.Context, db *sql.DB, postID int64) (Post, error) {
	var post Post
	var groupID sql.NullInt64
	var media sql.NullString

	row := db.QueryRowContext(ctx, "SELECT id, user_id, group_id, text, visibility, media_path, created_at FROM posts WHERE id = ?", postID)
	if err := row.Scan(&post.ID, &post.UserID, &groupID, &post.Text, &post.Visibility, &media, &post.CreatedAt); err != nil {
		return Post{}, err
	}
	if groupID.Valid {
		post.GroupID = &groupID.Int64
	}
	post.MediaPath = nullableStringPtr(media)
	return post, nil
}

func Feed(ctx context.Context, db *sql.DB, userID int64, limit, offset int) ([]Post, error) {
	query := `SELECT posts.id, posts.user_id, posts.group_id, posts.text, posts.visibility, posts.media_path, posts.created_at
		FROM posts
		LEFT JOIN follows ON follows.followee_id = posts.user_id AND follows.follower_id = ?
		LEFT JOIN post_allowed ON post_allowed.post_id = posts.id AND post_allowed.user_id = ?
		LEFT JOIN group_members ON group_members.group_id = posts.group_id AND group_members.user_id = ?
		WHERE posts.group_id IS NULL AND (
			posts.user_id = ?
			OR posts.visibility = 'public'
			OR (posts.visibility = 'followers' AND follows.follower_id IS NOT NULL)
			OR (posts.visibility = 'private' AND post_allowed.user_id IS NOT NULL)
		)
		OR (posts.group_id IS NOT NULL AND group_members.user_id IS NOT NULL)
		ORDER BY posts.created_at DESC
		LIMIT ? OFFSET ?`
	rows, err := db.QueryContext(ctx, query, userID, userID, userID, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanPosts(rows)
}

func UserPosts(ctx context.Context, db *sql.DB, viewerID, userID int64, limit, offset int) ([]Post, error) {
	query := `SELECT posts.id, posts.user_id, posts.group_id, posts.text, posts.visibility, posts.media_path, posts.created_at
		FROM posts
		LEFT JOIN follows ON follows.followee_id = posts.user_id AND follows.follower_id = ?
		LEFT JOIN post_allowed ON post_allowed.post_id = posts.id AND post_allowed.user_id = ?
		WHERE posts.user_id = ? AND posts.group_id IS NULL AND (
			posts.user_id = ?
			OR posts.visibility = 'public'
			OR (posts.visibility = 'followers' AND follows.follower_id IS NOT NULL)
			OR (posts.visibility = 'private' AND post_allowed.user_id IS NOT NULL)
		)
		ORDER BY posts.created_at DESC
		LIMIT ? OFFSET ?`
	rows, err := db.QueryContext(ctx, query, viewerID, viewerID, userID, viewerID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanPosts(rows)
}

func GroupPosts(ctx context.Context, db *sql.DB, userID, groupID int64, limit, offset int) ([]Post, error) {
	query := `SELECT posts.id, posts.user_id, posts.group_id, posts.text, posts.visibility, posts.media_path, posts.created_at
		FROM posts
		JOIN group_members ON group_members.group_id = posts.group_id AND group_members.user_id = ?
		WHERE posts.group_id = ?
		ORDER BY posts.created_at DESC
		LIMIT ? OFFSET ?`
	rows, err := db.QueryContext(ctx, query, userID, groupID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanPosts(rows)
}

func CanViewPost(ctx context.Context, db *sql.DB, viewerID, postID int64) (bool, error) {
	query := `SELECT posts.user_id, posts.group_id, posts.visibility,
		(SELECT 1 FROM follows WHERE follower_id = ? AND followee_id = posts.user_id LIMIT 1) AS follows,
		(SELECT 1 FROM post_allowed WHERE user_id = ? AND post_id = posts.id LIMIT 1) AS allowed,
		(SELECT 1 FROM group_members WHERE user_id = ? AND group_id = posts.group_id LIMIT 1) AS member
		FROM posts WHERE posts.id = ?`
	var authorID int64
	var groupID sql.NullInt64
	var visibility string
	var follows sql.NullInt64
	var allowed sql.NullInt64
	var member sql.NullInt64
	row := db.QueryRowContext(ctx, query, viewerID, viewerID, viewerID, postID)
	if err := row.Scan(&authorID, &groupID, &visibility, &follows, &allowed, &member); err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}

	if groupID.Valid {
		return member.Valid, nil
	}
	if authorID == viewerID {
		return true, nil
	}
	switch visibility {
	case "public":
		return true, nil
	case "followers":
		return follows.Valid, nil
	case "private":
		return allowed.Valid, nil
	default:
		return false, nil
	}
}

func EnsureAllowedFollowers(ctx context.Context, db *sql.DB, authorID int64, allowedIDs []int64) (bool, error) {
	if len(allowedIDs) == 0 {
		return true, nil
	}
	placeholders := make([]string, len(allowedIDs))
	args := make([]any, 0, len(allowedIDs)+1)
	args = append(args, authorID)
	for i, id := range allowedIDs {
		placeholders[i] = "?"
		args = append(args, id)
	}
	query := fmt.Sprintf("SELECT COUNT(DISTINCT follower_id) FROM follows WHERE followee_id = ? AND follower_id IN (%s)", strings.Join(placeholders, ","))
	var count int
	if err := db.QueryRowContext(ctx, query, args...).Scan(&count); err != nil {
		return false, err
	}
	return count == len(allowedIDs), nil
}

func scanPosts(rows *sql.Rows) ([]Post, error) {
	var posts []Post
	for rows.Next() {
		var post Post
		var groupID sql.NullInt64
		var media sql.NullString
		if err := rows.Scan(&post.ID, &post.UserID, &groupID, &post.Text, &post.Visibility, &media, &post.CreatedAt); err != nil {
			return nil, err
		}
		if groupID.Valid {
			post.GroupID = &groupID.Int64
		}
		post.MediaPath = nullableStringPtr(media)
		posts = append(posts, post)
	}
	return posts, rows.Err()
}

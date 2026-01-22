package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"

	"backend/internal/http/middleware"
	"backend/internal/repo"
)

type postResponse struct {
	ID        int64   `json:"id"`
	UserID    int64   `json:"user_id"`
	GroupID   *int64  `json:"group_id,omitempty"`
	Text      string  `json:"text"`
	Visibility string `json:"visibility"`
	MediaPath *string `json:"media_path,omitempty"`
	CreatedAt string  `json:"created_at"`
}

func CreatePost(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		current, ok := middleware.CurrentUser(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
			return
		}

		var req struct {
			Text              string  `json:"text"`
			Visibility        string  `json:"visibility"`
			AllowedFollowerIDs []int64 `json:"allowed_follower_ids"`
			MediaPath         *string `json:"media_path"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid json"})
			return
		}
		text := strings.TrimSpace(req.Text)
		if text == "" {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "text required"})
			return
		}
		if req.Visibility != "public" && req.Visibility != "followers" && req.Visibility != "private" {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid visibility"})
			return
		}
		allowed := uniquePositiveIDs(req.AllowedFollowerIDs)
		allowed = removeID(allowed, current.ID)

		if req.Visibility == "private" {
			if len(allowed) == 0 {
				writeJSON(w, http.StatusBadRequest, errorResponse{Error: "allowed_follower_ids required"})
				return
			}
			ok, err := repo.EnsureAllowedFollowers(r.Context(), db, current.ID, allowed)
			if err != nil || !ok {
				writeJSON(w, http.StatusBadRequest, errorResponse{Error: "allowed_follower_ids must be followers"})
				return
			}
		} else if len(allowed) > 0 {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "allowed_follower_ids only for private"})
			return
		}

		post, err := repo.CreatePost(r.Context(), db, current.ID, text, req.Visibility, req.MediaPath, allowed, nil)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "create failed"})
			return
		}
		writeJSON(w, http.StatusCreated, toPostResponse(post))
	}
}

func Feed(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		current, ok := middleware.CurrentUser(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
			return
		}
		limit, offset, err := parsePagination(r)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
			return
		}
		posts, err := repo.Feed(r.Context(), db, current.ID, limit, offset)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "feed failed"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"posts": toPostResponses(posts), "limit": limit, "offset": offset})
	}
}

func UserPosts(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		current, ok := middleware.CurrentUser(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
			return
		}
		userID, ok := parseIDParam(r, "id")
		if !ok {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid id"})
			return
		}
		limit, offset, err := parsePagination(r)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
			return
		}
		posts, err := repo.UserPosts(r.Context(), db, current.ID, userID, limit, offset)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "posts failed"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"posts": toPostResponses(posts), "limit": limit, "offset": offset})
	}
}

func CreateComment(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		current, ok := middleware.CurrentUser(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
			return
		}
		postID, ok := parseIDParam(r, "id")
		if !ok {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid id"})
			return
		}
		canView, err := repo.CanViewPost(r.Context(), db, current.ID, postID)
		if err != nil || !canView {
			writeJSON(w, http.StatusForbidden, errorResponse{Error: "forbidden"})
			return
		}
		var req struct {
			Text      string  `json:"text"`
			MediaPath *string `json:"media_path"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid json"})
			return
		}
		text := strings.TrimSpace(req.Text)
		if text == "" {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "text required"})
			return
		}
		comment, err := repo.CreateComment(r.Context(), db, postID, current.ID, text, req.MediaPath)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "comment failed"})
			return
		}
		writeJSON(w, http.StatusCreated, toCommentResponse(comment))
	}
}

func ListComments(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		current, ok := middleware.CurrentUser(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
			return
		}
		postID, ok := parseIDParam(r, "id")
		if !ok {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid id"})
			return
		}
		canView, err := repo.CanViewPost(r.Context(), db, current.ID, postID)
		if err != nil || !canView {
			writeJSON(w, http.StatusForbidden, errorResponse{Error: "forbidden"})
			return
		}
		limit, offset, err := parsePagination(r)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
			return
		}
		comments, err := repo.ListComments(r.Context(), db, postID, limit, offset)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "comments failed"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"comments": toCommentResponses(comments), "limit": limit, "offset": offset})
	}
}

func toPostResponse(post repo.Post) postResponse {
	return postResponse{
		ID: post.ID,
		UserID: post.UserID,
		GroupID: post.GroupID,
		Text: post.Text,
		Visibility: post.Visibility,
		MediaPath: post.MediaPath,
		CreatedAt: post.CreatedAt,
	}
}

func toPostResponses(posts []repo.Post) []postResponse {
	result := make([]postResponse, 0, len(posts))
	for _, post := range posts {
		result = append(result, toPostResponse(post))
	}
	return result
}

func toCommentResponse(comment repo.Comment) map[string]any {
	return map[string]any{
		"id": comment.ID,
		"post_id": comment.PostID,
		"user_id": comment.UserID,
		"text": comment.Text,
		"media_path": comment.MediaPath,
		"created_at": comment.CreatedAt,
	}
}

func toCommentResponses(comments []repo.Comment) []map[string]any {
	result := make([]map[string]any, 0, len(comments))
	for _, comment := range comments {
		result = append(result, toCommentResponse(comment))
	}
	return result
}

func uniquePositiveIDs(ids []int64) []int64 {
	seen := make(map[int64]struct{}, len(ids))
	result := make([]int64, 0, len(ids))
	for _, id := range ids {
		if id <= 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}
	return result
}

func removeID(ids []int64, target int64) []int64 {
	result := make([]int64, 0, len(ids))
	for _, id := range ids {
		if id == target {
			continue
		}
		result = append(result, id)
	}
	return result
}

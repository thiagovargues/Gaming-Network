package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"backend/internal/http/middleware"
	store "backend/internal/repo"
)

type commentResponse struct {
	ID        int64  `json:"id"`
	PostID    int64  `json:"post_id"`
	AuthorID  int64  `json:"author_id"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
}

type commentsResponse struct {
	Comments []commentResponse `json:"comments"`
	Limit    int               `json:"limit"`
	Offset   int               `json:"offset"`
}

func PostComments(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		postID, ok := parsePostCommentsPath(r.URL.Path)
		if !ok {
			http.NotFound(w, r)
			return
		}

		currentUser, ok := middleware.CurrentUser(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
			return
		}

		canView, err := store.CanUserViewPost(r.Context(), db, currentUser.ID, postID)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "comments failed"})
			return
		}
		if !canView {
			writeJSON(w, http.StatusForbidden, errorResponse{Error: "forbidden"})
			return
		}

		switch r.Method {
		case http.MethodPost:
			handleCreateComment(w, r, db, currentUser.ID, postID)
		case http.MethodGet:
			handleListComments(w, r, db, postID)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
}

func handleCreateComment(w http.ResponseWriter, r *http.Request, db *sql.DB, authorID, postID int64) {
	type request struct {
		Content string `json:"content"`
	}

	var req request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid json"})
		return
	}

	content := strings.TrimSpace(req.Content)
	if content == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "content required"})
		return
	}

	comment, err := store.CreateComment(r.Context(), db, postID, authorID, content)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "comments failed"})
		return
	}

	writeJSON(w, http.StatusCreated, toCommentResponse(comment))
}

func handleListComments(w http.ResponseWriter, r *http.Request, db *sql.DB, postID int64) {
	limit, offset, err := parsePagination(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}

	comments, err := store.ListComments(r.Context(), db, postID, limit, offset)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "comments failed"})
		return
	}

	writeJSON(w, http.StatusOK, commentsResponse{
		Comments: toCommentResponses(comments),
		Limit:    limit,
		Offset:   offset,
	})
}

func parsePostCommentsPath(path string) (int64, bool) {
	const prefix = "/api/posts/"
	if !strings.HasPrefix(path, prefix) {
		return 0, false
	}
	rest := strings.TrimPrefix(path, prefix)
	parts := strings.Split(rest, "/")
	if len(parts) != 2 || parts[1] != "comments" {
		return 0, false
	}
	id, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil || id <= 0 {
		return 0, false
	}
	return id, true
}

func toCommentResponse(comment store.Comment) commentResponse {
	return commentResponse{
		ID:        comment.ID,
		PostID:    comment.PostID,
		AuthorID:  comment.AuthorID,
		Content:   comment.Content,
		CreatedAt: comment.CreatedAt,
	}
}

func toCommentResponses(comments []store.Comment) []commentResponse {
	result := make([]commentResponse, 0, len(comments))
	for _, comment := range comments {
		result = append(result, toCommentResponse(comment))
	}
	return result
}

package httpapi

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	store "backend/internal/db"
)

type postResponse struct {
	ID         int64  `json:"id"`
	AuthorID   int64  `json:"author_id"`
	Content    string `json:"content"`
	Visibility string `json:"visibility"`
	CreatedAt  string `json:"created_at"`
}

type feedResponse struct {
	Posts  []postResponse `json:"posts"`
	Limit  int            `json:"limit"`
	Offset int            `json:"offset"`
}

func handleCreatePost(db *sql.DB) http.Handler {
	type request struct {
		Content        string  `json:"content"`
		Visibility     string  `json:"visibility"`
		AllowedUserIDs []int64 `json:"allowed_user_ids"`
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		currentUser, ok := currentUserFromContext(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
			return
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

		visibility := strings.TrimSpace(req.Visibility)
		if visibility != "public" && visibility != "followers" && visibility != "selected" {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid visibility"})
			return
		}

		allowedUserIDs := uniquePositiveIDs(req.AllowedUserIDs)
		allowedUserIDs = removeID(allowedUserIDs, currentUser.ID)
		if visibility == "selected" {
			if len(allowedUserIDs) == 0 {
				writeJSON(w, http.StatusBadRequest, errorResponse{Error: "allowed_user_ids required for selected visibility"})
				return
			}
			ok, err := store.AreAllowedViewersFollowers(r.Context(), db, currentUser.ID, allowedUserIDs)
			if err != nil {
				writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "create failed"})
				return
			}
			if !ok {
				writeJSON(w, http.StatusBadRequest, errorResponse{Error: "allowed_user_ids must be followers"})
				return
			}
		} else if len(allowedUserIDs) > 0 {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "allowed_user_ids only valid for selected visibility"})
			return
		}

		post, err := store.CreatePost(r.Context(), db, currentUser.ID, content, visibility, allowedUserIDs)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "create failed"})
			return
		}

		writeJSON(w, http.StatusCreated, toPostResponse(post))
	})
}

func handleFeed(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		currentUser, ok := currentUserFromContext(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
			return
		}

		limit, offset, err := parsePagination(r)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
			return
		}

		posts, err := store.ListFeed(r.Context(), db, currentUser.ID, limit, offset)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "feed failed"})
			return
		}

		writeJSON(w, http.StatusOK, feedResponse{Posts: toPostResponses(posts), Limit: limit, Offset: offset})
	})
}

func toPostResponse(post store.Post) postResponse {
	return postResponse{
		ID:         post.ID,
		AuthorID:   post.AuthorID,
		Content:    post.Content,
		Visibility: post.Visibility,
		CreatedAt:  post.CreatedAt,
	}
}

func toPostResponses(posts []store.Post) []postResponse {
	result := make([]postResponse, 0, len(posts))
	for _, post := range posts {
		result = append(result, toPostResponse(post))
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
	if len(ids) == 0 {
		return ids
	}
	result := make([]int64, 0, len(ids))
	for _, id := range ids {
		if id == target {
			continue
		}
		result = append(result, id)
	}
	return result
}

func parsePagination(r *http.Request) (int, int, error) {
	limit := 20
	offset := 0

	if raw := strings.TrimSpace(r.URL.Query().Get("limit")); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil || value <= 0 {
			return 0, 0, errInvalidPagination("limit")
		}
		limit = value
	}

	if raw := strings.TrimSpace(r.URL.Query().Get("offset")); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil || value < 0 {
			return 0, 0, errInvalidPagination("offset")
		}
		offset = value
	}

	if limit > 50 {
		limit = 50
	}

	return limit, offset, nil
}

func errInvalidPagination(field string) error {
	return fmt.Errorf("invalid %s", field)
}

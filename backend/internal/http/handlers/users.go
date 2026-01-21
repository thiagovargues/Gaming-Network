package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"backend/internal/http/middleware"
	"backend/internal/repo"
)

func GetUser(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		current, ok := middleware.CurrentUser(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
			return
		}
		id, ok := parseIDParam(r, "id")
		if !ok {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid id"})
			return
		}
		profile, found, err := repo.GetUserByID(r.Context(), db, id)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "profile failed"})
			return
		}
		if !found {
			writeJSON(w, http.StatusNotFound, errorResponse{Error: "user not found"})
			return
		}
		if !profile.IsPublic && profile.ID != current.ID {
			follows, err := repo.IsFollowing(r.Context(), db, current.ID, profile.ID)
			if err != nil {
				writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "profile failed"})
				return
			}
			if !follows {
				writeJSON(w, http.StatusForbidden, errorResponse{Error: "forbidden"})
				return
			}
		}
		writeJSON(w, http.StatusOK, profile)
	}
}

func UpdateMe(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		current, ok := middleware.CurrentUser(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
			return
		}
		var req struct {
			IsPublic *bool   `json:"is_public"`
			Avatar   *string `json:"avatar"`
			Nickname *string `json:"nickname"`
			About    *string `json:"about"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid json"})
			return
		}
		profile, err := repo.UpdateMe(r.Context(), db, current.ID, req.IsPublic, req.Avatar, req.Nickname, req.About)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "update failed"})
			return
		}
		writeJSON(w, http.StatusOK, profile)
	}
}

func ListFollowers(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := parseIDParam(r, "id")
		if !ok {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid id"})
			return
		}
		followers, err := repo.ListFollowers(r.Context(), db, id)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "followers failed"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"users": followers})
	}
}

func ListFollowing(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := parseIDParam(r, "id")
		if !ok {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid id"})
			return
		}
		following, err := repo.ListFollowing(r.Context(), db, id)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "following failed"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"users": following})
	}
}

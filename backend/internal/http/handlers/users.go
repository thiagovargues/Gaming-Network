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
		IsPublic  *bool   `json:"is_public"`
		Avatar    *string `json:"avatar"`
		Nickname  *string `json:"nickname"`
		About     *string `json:"about"`
		FirstName *string `json:"first_name"`
		LastName  *string `json:"last_name"`
		DOB       *string `json:"dob"`
		Sex       *string `json:"sex"`
		Age       *int    `json:"age"`
		ShowFirst *bool   `json:"show_first_name"`
		ShowLast  *bool   `json:"show_last_name"`
		ShowAge   *bool   `json:"show_age"`
		ShowSex   *bool   `json:"show_sex"`
		ShowNick  *bool   `json:"show_nickname"`
		ShowAbout *bool   `json:"show_about"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid json"})
		return
	}
	if req.Age != nil && *req.Age < 0 {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid age"})
		return
	}
	if req.Age != nil && *req.Age > 130 {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid age"})
		return
	}
	if req.FirstName != nil && len(*req.FirstName) > 30 {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "first_name too long"})
		return
	}
	if req.LastName != nil && len(*req.LastName) > 30 {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "last_name too long"})
		return
	}
	if req.Nickname != nil && len(*req.Nickname) > 30 {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "nickname too long"})
		return
	}
	if req.About != nil && len(*req.About) > 300 {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "bio too long"})
		return
	}
	profile, err := repo.UpdateMe(r.Context(), db, current.ID, req.IsPublic, req.Avatar, req.Nickname, req.About, req.FirstName, req.LastName, req.DOB, req.Sex, req.Age, req.ShowFirst, req.ShowLast, req.ShowAge, req.ShowSex, req.ShowNick, req.ShowAbout)
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

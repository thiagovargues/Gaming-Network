package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"

	"backend/internal/http/middleware"
	"backend/internal/repo"
)

func CreateEvent(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		current, ok := middleware.CurrentUser(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
			return
		}
		groupID, ok := parseIDParam(r, "id")
		if !ok {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid id"})
			return
		}
		member, err := repo.IsGroupMember(r.Context(), db, groupID, current.ID)
		if err != nil || !member {
			writeJSON(w, http.StatusForbidden, errorResponse{Error: "forbidden"})
			return
		}
		var req struct {
			Title       string `json:"title"`
			Description string `json:"description"`
			Datetime    string `json:"datetime"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid json"})
			return
		}
		if strings.TrimSpace(req.Title) == "" || strings.TrimSpace(req.Datetime) == "" {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "title and datetime required"})
			return
		}
		id, err := repo.CreateEvent(r.Context(), db, groupID, current.ID, req.Title, req.Description, req.Datetime)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "create failed"})
			return
		}
		members, err := repo.ListGroupMembers(r.Context(), db, groupID)
		if err == nil {
			for _, member := range members {
				if member.ID == current.ID {
					continue
				}
				payload := "{\"event_id\":" + intToString(id) + ",\"group_id\":" + intToString(groupID) + "}"
				_ = repo.CreateNotification(r.Context(), db, member.ID, "group_event", payload)
			}
		}
		writeJSON(w, http.StatusCreated, map[string]any{"id": id})
	}
}

func ListEvents(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		current, ok := middleware.CurrentUser(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
			return
		}
		groupID, ok := parseIDParam(r, "id")
		if !ok {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid id"})
			return
		}
		member, err := repo.IsGroupMember(r.Context(), db, groupID, current.ID)
		if err != nil || !member {
			writeJSON(w, http.StatusForbidden, errorResponse{Error: "forbidden"})
			return
		}
		events, err := repo.ListEvents(r.Context(), db, groupID)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "list failed"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"events": events})
	}
}

func RespondEvent(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		current, ok := middleware.CurrentUser(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
			return
		}
		eventID, ok := parseIDParam(r, "id")
		if !ok {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid id"})
			return
		}
		var req struct {
			Status string `json:"status"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid json"})
			return
		}
		if req.Status != "going" && req.Status != "not_going" {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid status"})
			return
		}
		if err := repo.RespondEvent(r.Context(), db, eventID, current.ID, req.Status); err != nil {
			writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "respond failed"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	}
}

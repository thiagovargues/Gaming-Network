package handlers

import (
	"database/sql"
	"net/http"

	"backend/internal/http/middleware"
	"backend/internal/repo"
)

func ListNotifications(db *sql.DB) http.HandlerFunc {
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
		items, err := repo.ListNotifications(r.Context(), db, current.ID, limit, offset)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "notifications failed"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"notifications": items, "limit": limit, "offset": offset})
	}
}

func MarkNotificationRead(db *sql.DB) http.HandlerFunc {
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
		if err := repo.MarkNotificationRead(r.Context(), db, id, current.ID); err != nil {
			writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "mark failed"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	}
}

func MarkAllNotificationsRead(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		current, ok := middleware.CurrentUser(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
			return
		}
		if err := repo.MarkAllNotificationsRead(r.Context(), db, current.ID); err != nil {
			writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "mark failed"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	}
}

package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"backend/internal/http/middleware"
	"backend/internal/repo"
)

func FollowRequest(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		current, ok := middleware.CurrentUser(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
			return
		}
		var req struct {
			ToUserID int64 `json:"to_user_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid json"})
			return
		}
		if req.ToUserID <= 0 || req.ToUserID == current.ID {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid to_user_id"})
			return
		}

		isPublic, found, err := repo.IsUserPublic(r.Context(), db, req.ToUserID)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "follow failed"})
			return
		}
		if !found {
			writeJSON(w, http.StatusNotFound, errorResponse{Error: "user not found"})
			return
		}

		if isPublic {
			if err := repo.CreateFollow(r.Context(), db, current.ID, req.ToUserID); err != nil {
				writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "follow failed"})
				return
			}
			writeJSON(w, http.StatusOK, map[string]string{"status": "following"})
			return
		}

		id, err := repo.CreateFollowRequest(r.Context(), db, current.ID, req.ToUserID)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "follow failed"})
			return
		}
		payload := "{\"request_id\":" + intToString(id) + ",\"from_user_id\":" + intToString(current.ID) + "}"
		_ = repo.CreateNotification(r.Context(), db, req.ToUserID, "follow_request", payload)
		writeJSON(w, http.StatusOK, map[string]string{"status": "requested"})
	}
}

func AcceptFollow(db *sql.DB) http.HandlerFunc {
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

		fromID, toID, found, err := repo.UpdateFollowRequestStatus(r.Context(), db, id, "accepted")
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "accept failed"})
			return
		}
		if !found || toID != current.ID {
			writeJSON(w, http.StatusNotFound, errorResponse{Error: "request not found"})
			return
		}
		if err := repo.CreateFollow(r.Context(), db, fromID, toID); err != nil {
			writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "accept failed"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "accepted"})
	}
}

func RefuseFollow(db *sql.DB) http.HandlerFunc {
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

		_, toID, found, err := repo.UpdateFollowRequestStatus(r.Context(), db, id, "refused")
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "refuse failed"})
			return
		}
		if !found || toID != current.ID {
			writeJSON(w, http.StatusNotFound, errorResponse{Error: "request not found"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "refused"})
	}
}

func Unfollow(db *sql.DB) http.HandlerFunc {
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
		if err := repo.DeleteFollow(r.Context(), db, current.ID, id); err != nil {
			writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "unfollow failed"})
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func ListIncomingFollowRequests(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		current, ok := middleware.CurrentUser(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
			return
		}
		list, err := repo.ListIncomingFollowRequests(r.Context(), db, current.ID)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "requests failed"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"requests": list})
	}
}

func ListOutgoingFollowRequests(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		current, ok := middleware.CurrentUser(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
			return
		}
		list, err := repo.ListOutgoingFollowRequests(r.Context(), db, current.ID)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "requests failed"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"requests": list})
	}
}

func intToString(id int64) string {
	return strconv.FormatInt(id, 10)
}

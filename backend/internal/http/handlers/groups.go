package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"

	"backend/internal/http/middleware"
	"backend/internal/repo"
)

func CreateGroup(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		current, ok := middleware.CurrentUser(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
			return
		}
		var req struct {
			Title       string `json:"title"`
			Description string `json:"description"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid json"})
			return
		}
		if strings.TrimSpace(req.Title) == "" {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "title required"})
			return
		}
		group, err := repo.CreateGroup(r.Context(), db, current.ID, req.Title, req.Description)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "create failed"})
			return
		}
		writeJSON(w, http.StatusCreated, group)
	}
}

func ListGroups(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := strings.TrimSpace(r.URL.Query().Get("query"))
		limit, offset, err := parsePagination(r)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
			return
		}
		groups, err := repo.ListGroups(r.Context(), db, query, limit, offset)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "list failed"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"groups": groups, "limit": limit, "offset": offset})
	}
}

func GetGroup(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := parseIDParam(r, "id")
		if !ok {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid id"})
			return
		}
		group, err := repo.GetGroup(r.Context(), db, id)
		if err != nil {
			writeJSON(w, http.StatusNotFound, errorResponse{Error: "group not found"})
			return
		}
		writeJSON(w, http.StatusOK, group)
	}
}

func InviteToGroup(db *sql.DB) http.HandlerFunc {
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
			UserID int64 `json:"user_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid json"})
			return
		}
		if req.UserID <= 0 {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "user_id required"})
			return
		}
		inviteID, err := repo.CreateGroupInvite(r.Context(), db, groupID, current.ID, req.UserID)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "invite failed"})
			return
		}
		payload := "{\"invite_id\":" + intToString(inviteID) + ",\"group_id\":" + intToString(groupID) + "}"
		_ = repo.CreateNotification(r.Context(), db, req.UserID, "group_invite", payload)
		writeJSON(w, http.StatusOK, map[string]string{"status": "invited"})
	}
}

func AcceptGroupInvite(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		current, ok := middleware.CurrentUser(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
			return
		}
		inviteID, ok := parseIDParam(r, "id")
		if !ok {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid id"})
			return
		}
		groupID, toID, found, err := repo.UpdateGroupInviteStatus(r.Context(), db, inviteID, "accepted")
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "accept failed"})
			return
		}
		if !found || toID != current.ID {
			writeJSON(w, http.StatusNotFound, errorResponse{Error: "invite not found"})
			return
		}
		if err := repo.AddGroupMember(r.Context(), db, groupID, current.ID, "member"); err != nil {
			writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "accept failed"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "accepted"})
	}
}

func RefuseGroupInvite(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		current, ok := middleware.CurrentUser(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
			return
		}
		inviteID, ok := parseIDParam(r, "id")
		if !ok {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid id"})
			return
		}
		_, toID, found, err := repo.UpdateGroupInviteStatus(r.Context(), db, inviteID, "refused")
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "refuse failed"})
			return
		}
		if !found || toID != current.ID {
			writeJSON(w, http.StatusNotFound, errorResponse{Error: "invite not found"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "refused"})
	}
}

func RequestJoinGroup(db *sql.DB) http.HandlerFunc {
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
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "join failed"})
			return
		}
		if member {
			writeJSON(w, http.StatusOK, map[string]string{"status": "already_member"})
			return
		}
		requestID, err := repo.CreateJoinRequest(r.Context(), db, groupID, current.ID)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "join failed"})
			return
		}
		creatorID, err := repo.GroupCreator(r.Context(), db, groupID)
		if err == nil {
			payload := "{\"join_request_id\":" + intToString(requestID) + ",\"group_id\":" + intToString(groupID) + "}"
			_ = repo.CreateNotification(r.Context(), db, creatorID, "group_join_request", payload)
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "requested"})
	}
}

func AcceptJoinRequest(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		current, ok := middleware.CurrentUser(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
			return
		}
		requestID, ok := parseIDParam(r, "id")
		if !ok {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid id"})
			return
		}
		groupID, userID, found, err := repo.UpdateJoinRequestStatus(r.Context(), db, requestID, "accepted")
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "accept failed"})
			return
		}
		if !found {
			writeJSON(w, http.StatusNotFound, errorResponse{Error: "request not found"})
			return
		}
		creatorID, err := repo.GroupCreator(r.Context(), db, groupID)
		if err != nil || creatorID != current.ID {
			writeJSON(w, http.StatusForbidden, errorResponse{Error: "forbidden"})
			return
		}
		if err := repo.AddGroupMember(r.Context(), db, groupID, userID, "member"); err != nil {
			writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "accept failed"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "accepted"})
	}
}

func RefuseJoinRequest(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		current, ok := middleware.CurrentUser(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
			return
		}
		requestID, ok := parseIDParam(r, "id")
		if !ok {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid id"})
			return
		}
		groupID, _, found, err := repo.UpdateJoinRequestStatus(r.Context(), db, requestID, "refused")
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "refuse failed"})
			return
		}
		if !found {
			writeJSON(w, http.StatusNotFound, errorResponse{Error: "request not found"})
			return
		}
		creatorID, err := repo.GroupCreator(r.Context(), db, groupID)
		if err != nil || creatorID != current.ID {
			writeJSON(w, http.StatusForbidden, errorResponse{Error: "forbidden"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "refused"})
	}
}

func ListGroupMembers(db *sql.DB) http.HandlerFunc {
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
		members, err := repo.ListGroupMembers(r.Context(), db, groupID)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "members failed"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"users": members})
	}
}

func CreateGroupPost(db *sql.DB) http.HandlerFunc {
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
		post, err := repo.CreatePost(r.Context(), db, current.ID, text, "group", req.MediaPath, nil, &groupID)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "create failed"})
			return
		}
		writeJSON(w, http.StatusCreated, toPostResponse(post))
	}
}

func ListGroupPosts(db *sql.DB) http.HandlerFunc {
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
		limit, offset, err := parsePagination(r)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
			return
		}
		posts, err := repo.GroupPosts(r.Context(), db, current.ID, groupID, limit, offset)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "posts failed"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"posts": toPostResponses(posts), "limit": limit, "offset": offset})
	}
}

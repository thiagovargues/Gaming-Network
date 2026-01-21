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

type followActionResponse struct {
	Status string `json:"status"`
}

type listResponse struct {
	Users []userProfile `json:"users"`
}

func FollowRequest(db *sql.DB) http.Handler {
	type request struct {
		TargetID int64 `json:"target_id"`
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		currentUser, ok := middleware.CurrentUser(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
			return
		}

		var req request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid json"})
			return
		}

		if req.TargetID <= 0 {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "target_id required"})
			return
		}

		if req.TargetID == currentUser.ID {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "cannot follow yourself"})
			return
		}

		ctx := r.Context()
		isPrivate, found, err := store.GetUserPrivacy(ctx, db, req.TargetID)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "follow failed"})
			return
		}
		if !found {
			writeJSON(w, http.StatusNotFound, errorResponse{Error: "user not found"})
			return
		}

		following, err := store.IsFollowing(ctx, db, currentUser.ID, req.TargetID)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "follow failed"})
			return
		}
		if following {
			writeJSON(w, http.StatusOK, followActionResponse{Status: "following"})
			return
		}

		if !isPrivate {
			if _, err := store.CreateFollow(ctx, db, currentUser.ID, req.TargetID); err != nil {
				writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "follow failed"})
				return
			}
			_, _ = store.DeleteFollowRequest(ctx, db, currentUser.ID, req.TargetID)
			writeJSON(w, http.StatusOK, followActionResponse{Status: "following"})
			return
		}

		if _, err := store.CreateFollowRequest(ctx, db, currentUser.ID, req.TargetID); err != nil {
			writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "follow failed"})
			return
		}

		writeJSON(w, http.StatusOK, followActionResponse{Status: "requested"})
	})
}

func FollowAccept(db *sql.DB) http.Handler {
	type request struct {
		RequesterID int64 `json:"requester_id"`
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		currentUser, ok := middleware.CurrentUser(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
			return
		}

		var req request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid json"})
			return
		}

		if req.RequesterID <= 0 {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "requester_id required"})
			return
		}

		if req.RequesterID == currentUser.ID {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid requester_id"})
			return
		}

		accepted, err := store.AcceptFollowRequest(r.Context(), db, req.RequesterID, currentUser.ID)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "accept failed"})
			return
		}
		if !accepted {
			writeJSON(w, http.StatusNotFound, errorResponse{Error: "request not found"})
			return
		}

		writeJSON(w, http.StatusOK, followActionResponse{Status: "accepted"})
	})
}

func FollowRefuse(db *sql.DB) http.Handler {
	type request struct {
		RequesterID int64 `json:"requester_id"`
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		currentUser, ok := middleware.CurrentUser(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
			return
		}

		var req request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid json"})
			return
		}

		if req.RequesterID <= 0 {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "requester_id required"})
			return
		}

		if req.RequesterID == currentUser.ID {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid requester_id"})
			return
		}

		removed, err := store.DeleteFollowRequest(r.Context(), db, req.RequesterID, currentUser.ID)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "refuse failed"})
			return
		}
		if !removed {
			writeJSON(w, http.StatusNotFound, errorResponse{Error: "request not found"})
			return
		}

		writeJSON(w, http.StatusOK, followActionResponse{Status: "refused"})
	})
}

func FollowUnfollow(db *sql.DB) http.Handler {
	type request struct {
		TargetID int64 `json:"target_id"`
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		currentUser, ok := middleware.CurrentUser(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
			return
		}

		var req request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid json"})
			return
		}

		if req.TargetID <= 0 {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "target_id required"})
			return
		}

		if req.TargetID == currentUser.ID {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "cannot unfollow yourself"})
			return
		}

		removed, err := store.DeleteFollow(r.Context(), db, currentUser.ID, req.TargetID)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "unfollow failed"})
			return
		}
		if !removed {
			writeJSON(w, http.StatusNotFound, errorResponse{Error: "not following"})
			return
		}

		writeJSON(w, http.StatusOK, followActionResponse{Status: "unfollowed"})
	})
}

func profileRelations(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		currentUser, ok := middleware.CurrentUser(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
			return
		}

		profileID, action, ok := parseProfilePath(r.URL.Path)
		if !ok {
			http.NotFound(w, r)
			return
		}

		ctx := r.Context()
		isPrivate, found, err := store.GetUserPrivacy(ctx, db, profileID)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "list failed"})
			return
		}
		if !found {
			writeJSON(w, http.StatusNotFound, errorResponse{Error: "user not found"})
			return
		}

		if isPrivate && currentUser.ID != profileID {
			following, err := store.IsFollowing(ctx, db, currentUser.ID, profileID)
			if err != nil {
				writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "list failed"})
				return
			}
			if !following {
				writeJSON(w, http.StatusForbidden, errorResponse{Error: "forbidden"})
				return
			}
		}

		var profiles []store.UserProfile
		err = nil

		switch action {
		case "followers":
			profiles, err = store.ListFollowers(ctx, db, profileID)
		case "following":
			profiles, err = store.ListFollowing(ctx, db, profileID)
		default:
			http.NotFound(w, r)
			return
		}

		if err != nil {
			writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "list failed"})
			return
		}

		writeJSON(w, http.StatusOK, listResponse{Users: toAPIProfiles(profiles)})
	})
}

func parseProfilePath(path string) (int64, string, bool) {
	const prefix = "/api/profile/"
	if !strings.HasPrefix(path, prefix) {
		return 0, "", false
	}
	rest := strings.TrimPrefix(path, prefix)
	parts := strings.Split(rest, "/")
	if len(parts) != 2 {
		return 0, "", false
	}
	id, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil || id <= 0 {
		return 0, "", false
	}
	if parts[1] != "followers" && parts[1] != "following" {
		return 0, "", false
	}
	return id, parts[1], true
}

func toAPIProfiles(profiles []store.UserProfile) []userProfile {
	result := make([]userProfile, 0, len(profiles))
	for _, profile := range profiles {
		result = append(result, userProfile{
			ID:         profile.ID,
			Email:      profile.Email,
			FirstName:  profile.FirstName,
			LastName:   profile.LastName,
			BirthDate:  profile.BirthDate,
			Nickname:   profile.Nickname,
			AboutMe:    profile.AboutMe,
			AvatarPath: profile.AvatarPath,
			IsPrivate:  profile.IsPrivate,
			CreatedAt:  profile.CreatedAt,
		})
	}
	return result
}

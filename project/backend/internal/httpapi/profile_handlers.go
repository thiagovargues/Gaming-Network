package httpapi

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	store "backend/internal/db"
)

func handleProfileRoutes(db *sql.DB) http.Handler {
	relationsHandler := handleProfileRelations(db)
	profileHandler := handleProfileView(db)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, _, ok := parseProfilePath(r.URL.Path); ok {
			relationsHandler.ServeHTTP(w, r)
			return
		}

		if _, ok := parseProfileID(r.URL.Path); ok {
			profileHandler.ServeHTTP(w, r)
			return
		}

		http.NotFound(w, r)
	})
}

func handleProfileView(db *sql.DB) http.Handler {
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

		profileID, ok := parseProfileID(r.URL.Path)
		if !ok {
			http.NotFound(w, r)
			return
		}

		ctx := r.Context()
		profile, found, err := store.GetUserByID(ctx, db, profileID)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "profile failed"})
			return
		}
		if !found {
			writeJSON(w, http.StatusNotFound, errorResponse{Error: "user not found"})
			return
		}

		if profile.IsPrivate && currentUser.ID != profileID {
			following, err := store.IsFollowing(ctx, db, currentUser.ID, profileID)
			if err != nil {
				writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "profile failed"})
				return
			}
			if !following {
				writeJSON(w, http.StatusForbidden, errorResponse{Error: "forbidden"})
				return
			}
		}

		writeJSON(w, http.StatusOK, toAPIProfile(profile))
	})
}

func handleProfilePrivacy(db *sql.DB) http.Handler {
	type request struct {
		IsPrivate bool `json:"is_private"`
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

		profile, found, err := store.UpdateUserPrivacy(r.Context(), db, currentUser.ID, req.IsPrivate)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "profile failed"})
			return
		}
		if !found {
			writeJSON(w, http.StatusNotFound, errorResponse{Error: "user not found"})
			return
		}

		writeJSON(w, http.StatusOK, toAPIProfile(profile))
	})
}

func parseProfileID(path string) (int64, bool) {
	const prefix = "/api/profile/"
	if !strings.HasPrefix(path, prefix) {
		return 0, false
	}
	rest := strings.TrimPrefix(path, prefix)
	if rest == "" || strings.Contains(rest, "/") {
		return 0, false
	}
	id, err := strconv.ParseInt(rest, 10, 64)
	if err != nil || id <= 0 {
		return 0, false
	}
	return id, true
}

func toAPIProfile(profile store.UserProfile) userProfile {
	return userProfile{
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
	}
}

package httpapi

import (
	"database/sql"
	"net/http"
)

type healthResponse struct {
	OK bool `json:"ok"`
}

func NewRouter(db *sql.DB) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, healthResponse{OK: true})
	})

	mux.Handle("/api/auth/register", handleRegister(db))
	mux.Handle("/api/auth/login", handleLogin(db))
	mux.Handle("/api/auth/logout", RequireAuth(db, handleLogout(db)))
	mux.Handle("/api/me", RequireAuth(db, handleMe()))

	mux.Handle("/api/follow/request", RequireAuth(db, handleFollowRequest(db)))
	mux.Handle("/api/follow/accept", RequireAuth(db, handleFollowAccept(db)))
	mux.Handle("/api/follow/refuse", RequireAuth(db, handleFollowRefuse(db)))
	mux.Handle("/api/follow/unfollow", RequireAuth(db, handleFollowUnfollow(db)))

	mux.Handle("/api/posts", RequireAuth(db, handleCreatePost(db)))
	mux.Handle("/api/posts/", RequireAuth(db, handlePostComments(db)))
	mux.Handle("/api/feed", RequireAuth(db, handleFeed(db)))

	mux.Handle("/api/profile/privacy", RequireAuth(db, handleProfilePrivacy(db)))
	mux.Handle("/api/profile/", RequireAuth(db, handleProfileRoutes(db)))

	return mux
}

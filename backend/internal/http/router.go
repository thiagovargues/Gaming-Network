package httpapi

import (
	"database/sql"
	"net/http"

	"backend/internal/http/handlers"
	"backend/internal/http/middleware"
)

type healthResponse struct {
	OK bool `json:"ok"`
}

func NewRouter(db *sql.DB) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, healthResponse{OK: true})
	})

	mux.Handle("/api/auth/register", handlers.Register(db))
	mux.Handle("/api/auth/login", handlers.Login(db))
	mux.Handle("/api/auth/logout", middleware.RequireAuth(db, handlers.Logout(db)))
	mux.Handle("/api/me", middleware.RequireAuth(db, handlers.Me()))

	mux.Handle("/api/follow/request", middleware.RequireAuth(db, handlers.FollowRequest(db)))
	mux.Handle("/api/follow/accept", middleware.RequireAuth(db, handlers.FollowAccept(db)))
	mux.Handle("/api/follow/refuse", middleware.RequireAuth(db, handlers.FollowRefuse(db)))
	mux.Handle("/api/follow/unfollow", middleware.RequireAuth(db, handlers.FollowUnfollow(db)))

	mux.Handle("/api/posts", middleware.RequireAuth(db, handlers.CreatePost(db)))
	mux.Handle("/api/posts/", middleware.RequireAuth(db, handlers.PostComments(db)))
	mux.Handle("/api/feed", middleware.RequireAuth(db, handlers.Feed(db)))

	mux.Handle("/api/profile/privacy", middleware.RequireAuth(db, handlers.ProfilePrivacy(db)))
	mux.Handle("/api/profile/", middleware.RequireAuth(db, handlers.ProfileRoutes(db)))

	return mux
}

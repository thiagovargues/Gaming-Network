package http

import (
	"database/sql"
	"net/http"
	"time"

	"backend/internal/config"
	"backend/internal/http/handlers"
	appmw "backend/internal/http/middleware"
	"backend/internal/ws"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
)

func NewRouter(cfg config.Config, db *sql.DB) http.Handler {
	r := chi.NewRouter()
	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(chimw.Recoverer)
	r.Use(chimw.Timeout(30 * time.Second))
	r.Use(chimw.NoCache)
	r.Use(chimw.CleanPath)

	r.Use(appmw.CORS(cfg.CORSOrigin))

	r.Get("/health", handlers.Health())

	r.Route("/api", func(api chi.Router) {
		api.Post("/auth/register", handlers.Register(cfg, db))
		api.Post("/auth/login", handlers.Login(cfg, db))
		api.Post("/auth/logout", handlers.Logout(cfg, db))
		api.Get("/auth/google/start", handlers.GoogleStart(cfg))
		api.Get("/auth/google/callback", handlers.GoogleCallback(cfg, db))
		api.With(appmw.RequireAuth(cfg, db)).Get("/me", handlers.Me())

		api.With(appmw.RequireAuth(cfg, db)).Get("/users/{id}", handlers.GetUser(db))
		api.With(appmw.RequireAuth(cfg, db)).Patch("/users/me", handlers.UpdateMe(db))
		api.With(appmw.RequireAuth(cfg, db)).Get("/users/{id}/followers", handlers.ListFollowers(db))
		api.With(appmw.RequireAuth(cfg, db)).Get("/users/{id}/following", handlers.ListFollowing(db))

		api.With(appmw.RequireAuth(cfg, db)).Post("/follows/request", handlers.FollowRequest(db))
		api.With(appmw.RequireAuth(cfg, db)).Post("/follows/request/{id}/accept", handlers.AcceptFollow(db))
		api.With(appmw.RequireAuth(cfg, db)).Post("/follows/request/{id}/refuse", handlers.RefuseFollow(db))
		api.With(appmw.RequireAuth(cfg, db)).Delete("/follows/{id}", handlers.Unfollow(db))
		api.With(appmw.RequireAuth(cfg, db)).Get("/follows/requests/incoming", handlers.ListIncomingFollowRequests(db))
		api.With(appmw.RequireAuth(cfg, db)).Get("/follows/requests/outgoing", handlers.ListOutgoingFollowRequests(db))

		api.With(appmw.RequireAuth(cfg, db)).Post("/posts", handlers.CreatePost(db))
		api.With(appmw.RequireAuth(cfg, db)).Delete("/posts/{id}", handlers.DeletePost(db))
		api.With(appmw.RequireAuth(cfg, db)).Get("/feed", handlers.Feed(db))
		api.With(appmw.RequireAuth(cfg, db)).Get("/users/{id}/posts", handlers.UserPosts(db))
		api.With(appmw.RequireAuth(cfg, db)).Post("/posts/{id}/comments", handlers.CreateComment(db))
		api.With(appmw.RequireAuth(cfg, db)).Get("/posts/{id}/comments", handlers.ListComments(db))

		api.With(appmw.RequireAuth(cfg, db)).Post("/media/upload", handlers.UploadMedia(cfg, db))

		api.With(appmw.RequireAuth(cfg, db)).Post("/groups", handlers.CreateGroup(db))
		api.With(appmw.RequireAuth(cfg, db)).Get("/groups", handlers.ListGroups(db))
		api.With(appmw.RequireAuth(cfg, db)).Get("/groups/{id}", handlers.GetGroup(db))
		api.With(appmw.RequireAuth(cfg, db)).Post("/groups/{id}/invite", handlers.InviteToGroup(db))
		api.With(appmw.RequireAuth(cfg, db)).Post("/groups/invites/{id}/accept", handlers.AcceptGroupInvite(db))
		api.With(appmw.RequireAuth(cfg, db)).Post("/groups/invites/{id}/refuse", handlers.RefuseGroupInvite(db))
		api.With(appmw.RequireAuth(cfg, db)).Post("/groups/{id}/join-request", handlers.RequestJoinGroup(db))
		api.With(appmw.RequireAuth(cfg, db)).Post("/groups/join-requests/{id}/accept", handlers.AcceptJoinRequest(db))
		api.With(appmw.RequireAuth(cfg, db)).Post("/groups/join-requests/{id}/refuse", handlers.RefuseJoinRequest(db))
		api.With(appmw.RequireAuth(cfg, db)).Get("/groups/{id}/members", handlers.ListGroupMembers(db))
		api.With(appmw.RequireAuth(cfg, db)).Post("/groups/{id}/posts", handlers.CreateGroupPost(db))
		api.With(appmw.RequireAuth(cfg, db)).Get("/groups/{id}/posts", handlers.ListGroupPosts(db))

		api.With(appmw.RequireAuth(cfg, db)).Post("/groups/{id}/events", handlers.CreateEvent(db))
		api.With(appmw.RequireAuth(cfg, db)).Get("/groups/{id}/events", handlers.ListEvents(db))
		api.With(appmw.RequireAuth(cfg, db)).Post("/events/{id}/respond", handlers.RespondEvent(db))

		api.With(appmw.RequireAuth(cfg, db)).Get("/notifications", handlers.ListNotifications(db))
		api.With(appmw.RequireAuth(cfg, db)).Post("/notifications/{id}/read", handlers.MarkNotificationRead(db))
		api.With(appmw.RequireAuth(cfg, db)).Post("/notifications/read-all", handlers.MarkAllNotificationsRead(db))

		api.With(appmw.RequireAuth(cfg, db)).Get("/ws", ws.NewHandler(cfg, db))
	})

	r.Handle("/media/*", handlers.Media(cfg))

	return r
}

package middleware

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"backend/internal/config"
	"backend/internal/repo"
)

type contextKey string

const userContextKey contextKey = "user"

type errorResponse struct {
	Error string `json:"error"`
}

func RequireAuth(cfg config.Config, db *sql.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			session, err := r.Cookie(cfg.CookieName)
			if err != nil || session.Value == "" {
				writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
				return
			}

			profile, ok, err := repo.GetSessionUser(r.Context(), db, session.Value)
			if err != nil || !ok {
				writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
				return
			}

			ctx := context.WithValue(r.Context(), userContextKey, profile)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func CurrentUser(r *http.Request) (repo.UserProfile, bool) {
	value := r.Context().Value(userContextKey)
	profile, ok := value.(repo.UserProfile)
	return profile, ok
}

func SetSessionCookie(cfg config.Config, w http.ResponseWriter, token string, expiresAt time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name:     cfg.CookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   cfg.CookieSecure,
		SameSite: http.SameSiteLaxMode,
		Expires:  expiresAt,
	})
}

func ClearSessionCookie(cfg config.Config, w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     cfg.CookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   cfg.CookieSecure,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
	})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

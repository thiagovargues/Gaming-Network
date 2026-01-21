package middleware

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
)

type UserProfile struct {
	ID         int64   `json:"id"`
	Email      string  `json:"email"`
	FirstName  string  `json:"first_name"`
	LastName   string  `json:"last_name"`
	BirthDate  string  `json:"birth_date"`
	Nickname   *string `json:"nickname"`
	AboutMe    *string `json:"about_me"`
	AvatarPath *string `json:"avatar_path"`
	IsPrivate  bool    `json:"is_private"`
	CreatedAt  string  `json:"created_at"`
}

type contextKey string

const userContextKey contextKey = "user"

type errorResponse struct {
	Error string `json:"error"`
}

func RequireAuth(db *sql.DB, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("sid")
		if err != nil || cookie.Value == "" {
			writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
			return
		}

		var profile UserProfile
		var nickname sql.NullString
		var aboutMe sql.NullString
		var avatarPath sql.NullString
		var isPrivate int

		row := db.QueryRow(
			`SELECT users.id, users.email, users.first_name, users.last_name, users.birth_date,
				users.nickname, users.about_me, users.avatar_path, users.is_private, users.created_at
			FROM sessions
			JOIN users ON users.id = sessions.user_id
			WHERE sessions.id = ? AND sessions.expires_at > CURRENT_TIMESTAMP`,
			cookie.Value,
		)
		if err := row.Scan(
			&profile.ID,
			&profile.Email,
			&profile.FirstName,
			&profile.LastName,
			&profile.BirthDate,
			&nickname,
			&aboutMe,
			&avatarPath,
			&isPrivate,
			&profile.CreatedAt,
		); err != nil {
			writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
			return
		}

		profile.IsPrivate = isPrivate != 0
		profile.Nickname = nullableStringPtr(nickname)
		profile.AboutMe = nullableStringPtr(aboutMe)
		profile.AvatarPath = nullableStringPtr(avatarPath)

		ctx := context.WithValue(r.Context(), userContextKey, profile)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func CurrentUser(r *http.Request) (UserProfile, bool) {
	value := r.Context().Value(userContextKey)
	profile, ok := value.(UserProfile)
	return profile, ok
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func nullableStringPtr(value sql.NullString) *string {
	if !value.Valid {
		return nil
	}
	val := strings.TrimSpace(value.String)
	if val == "" {
		return nil
	}
	return &val
}

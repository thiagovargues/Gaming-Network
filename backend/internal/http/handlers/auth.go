package handlers

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"backend/internal/http/middleware"
	"golang.org/x/crypto/bcrypt"
)

type authResponse struct {
	ID    int64  `json:"id"`
	Email string `json:"email"`
}

func Register(db *sql.DB) http.Handler {
	type request struct {
		Email     string  `json:"email"`
		Password  string  `json:"password"`
		FirstName string  `json:"first_name"`
		LastName  string  `json:"last_name"`
		BirthDate string  `json:"birth_date"`
		Nickname  *string `json:"nickname"`
		AboutMe   *string `json:"about_me"`
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		var req request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid json"})
			return
		}

		if strings.TrimSpace(req.Email) == "" ||
			strings.TrimSpace(req.Password) == "" ||
			strings.TrimSpace(req.FirstName) == "" ||
			strings.TrimSpace(req.LastName) == "" ||
			strings.TrimSpace(req.BirthDate) == "" {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "missing required fields"})
			return
		}

		birthDate, err := time.Parse("2006-01-02", req.BirthDate)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "birth_date must be YYYY-MM-DD"})
			return
		}

		hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "hash failed"})
			return
		}

		nickname := nullableString(req.Nickname)
		aboutMe := nullableString(req.AboutMe)

		result, err := db.Exec(
			`INSERT INTO users (
				email, password_hash, first_name, last_name, birth_date, nickname, about_me
			) VALUES (?, ?, ?, ?, ?, ?, ?)`,
			req.Email,
			string(hash),
			req.FirstName,
			req.LastName,
			birthDate.Format("2006-01-02"),
			nickname,
			aboutMe,
		)
		if err != nil {
			if isSQLiteConstraintError(err) {
				writeJSON(w, http.StatusConflict, errorResponse{Error: "email already exists"})
				return
			}
			writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "register failed"})
			return
		}

		id, _ := result.LastInsertId()
		writeJSON(w, http.StatusCreated, authResponse{ID: id, Email: req.Email})
	})
}

func Login(db *sql.DB) http.Handler {
	type request struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		var req request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid json"})
			return
		}

		if strings.TrimSpace(req.Email) == "" || strings.TrimSpace(req.Password) == "" {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "email and password required"})
			return
		}

		var id int64
		var passwordHash string
		row := db.QueryRow("SELECT id, password_hash FROM users WHERE email = ?", req.Email)
		if err := row.Scan(&id, &passwordHash); err != nil {
			writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "invalid credentials"})
			return
		}

		if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.Password)); err != nil {
			writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "invalid credentials"})
			return
		}

		sessionID, err := newSession(db, id)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "login failed"})
			return
		}

		setSessionCookie(w, sessionID)
		writeJSON(w, http.StatusOK, authResponse{ID: id, Email: req.Email})
	})
}

func Logout(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		cookie, err := r.Cookie("sid")
		if err == nil && cookie.Value != "" {
			_, _ = db.Exec("DELETE FROM sessions WHERE id = ?", cookie.Value)
		}

		clearSessionCookie(w)
		w.WriteHeader(http.StatusNoContent)
	})
}

func Me() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		currentUser, ok := middleware.CurrentUser(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
			return
		}

		writeJSON(w, http.StatusOK, currentUser)
	})
}

func newSession(db *sql.DB, userID int64) (string, error) {
	sessionID, err := newSessionID()
	if err != nil {
		return "", err
	}

	expiresAt := time.Now().Add(7 * 24 * time.Hour).UTC()
	_, err = db.Exec(
		"INSERT INTO sessions (id, user_id, expires_at) VALUES (?, ?, ?)",
		sessionID,
		userID,
		expiresAt,
	)
	if err != nil {
		return "", err
	}

	return sessionID, nil
}

func newSessionID() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

func setSessionCookie(w http.ResponseWriter, sessionID string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "sid",
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

func clearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "sid",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
	})
}

func isSQLiteConstraintError(err error) bool {
	return err != nil && strings.Contains(err.Error(), "UNIQUE constraint failed")
}

func nullableString(value *string) sql.NullString {
	if value == nil {
		return sql.NullString{}
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: trimmed, Valid: true}
}

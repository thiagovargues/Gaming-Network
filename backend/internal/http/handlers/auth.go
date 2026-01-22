package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"backend/internal/config"
	"backend/internal/http/middleware"
	"backend/internal/repo"

	"golang.org/x/crypto/bcrypt"
)

type authResponse struct {
	ID    int64  `json:"id"`
	Email string `json:"email"`
}

func Register(cfg config.Config, db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Email     string  `json:"email"`
			Password  string  `json:"password"`
			FirstName string  `json:"first_name"`
			LastName  string  `json:"last_name"`
			DOB       string  `json:"dob"`
			Avatar    *string `json:"avatar"`
			Nickname  *string `json:"nickname"`
			About     *string `json:"about"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid json"})
			return
		}
		if strings.TrimSpace(req.Email) == "" || strings.TrimSpace(req.Password) == "" ||
			strings.TrimSpace(req.FirstName) == "" || strings.TrimSpace(req.LastName) == "" || strings.TrimSpace(req.DOB) == "" {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "missing required fields"})
			return
		}

		hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "hash failed"})
			return
		}

		id, err := repo.CreateUser(r.Context(), db, req.Email, string(hash), req.FirstName, req.LastName, req.DOB, req.Avatar, req.Nickname, req.About)
		if err != nil {
			writeJSON(w, http.StatusConflict, errorResponse{Error: "email already exists"})
			return
		}
		writeJSON(w, http.StatusCreated, authResponse{ID: id, Email: req.Email})
	}
}

func Login(cfg config.Config, db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid json"})
			return
		}
		if strings.TrimSpace(req.Email) == "" || strings.TrimSpace(req.Password) == "" {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "email and password required"})
			return
		}

		id, hash, err := repo.GetUserByEmail(r.Context(), db, req.Email)
		if err != nil {
			writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "invalid credentials"})
			return
		}
		if hash == nil || bcrypt.CompareHashAndPassword([]byte(*hash), []byte(req.Password)) != nil {
			writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "invalid credentials"})
			return
		}

		token, expiresAt, err := repo.CreateSession(r.Context(), db, id, 7*24*time.Hour)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "login failed"})
			return
		}
		middleware.SetSessionCookie(cfg, w, token, expiresAt)
		writeJSON(w, http.StatusOK, authResponse{ID: id, Email: req.Email})
	}
}

func Logout(cfg config.Config, db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, _ := r.Cookie(cfg.CookieName)
		if cookie != nil && cookie.Value != "" {
			_ = repo.DeleteSession(r.Context(), db, cookie.Value)
		}
		middleware.ClearSessionCookie(cfg, w)
		w.WriteHeader(http.StatusNoContent)
	}
}

func Me() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := middleware.CurrentUser(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
			return
		}
		writeJSON(w, http.StatusOK, user)
	}
}

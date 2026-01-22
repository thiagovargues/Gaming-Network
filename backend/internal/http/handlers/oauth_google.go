package handlers

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"backend/internal/config"
	"backend/internal/http/middleware"
	"backend/internal/repo"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const (
	googleProvider   = "google"
	oauthStateCookie = "oauth_state"
)

type googleUserInfo struct {
	Sub           string `json:"sub"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
}

type oauthErrorResponse struct {
	Error  string `json:"error"`
	Detail string `json:"detail,omitempty"`
}

func GoogleStart(cfg config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conf, err := googleOAuthConfig(cfg)
		if err != nil {
			log.Printf("oauth google config error: %v", err)
			writeJSON(w, http.StatusInternalServerError, errorResponse{Error: err.Error()})
			return
		}

		state, err := randomState(32)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "state failed"})
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     oauthStateCookie,
			Value:    state,
			Path:     "/",
			HttpOnly: true,
			Secure:   cfg.CookieSecure,
			SameSite: http.SameSiteLaxMode,
			Expires:  time.Now().Add(10 * time.Minute),
		})

		url := conf.AuthCodeURL(state, oauth2.AccessTypeOnline)
		http.Redirect(w, r, url, http.StatusFound)
	}
}

func GoogleCallback(cfg config.Config, db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conf, err := googleOAuthConfig(cfg)
		if err != nil {
			log.Printf("oauth google config error: %v", err)
			writeJSON(w, http.StatusInternalServerError, oauthErrorResponse{Error: "oauth_config", Detail: err.Error()})
			return
		}

		state := r.URL.Query().Get("state")
		if state == "" {
			log.Printf("oauth google missing state")
			writeJSON(w, http.StatusBadRequest, oauthErrorResponse{Error: "missing_state"})
			return
		}
		cookie, err := r.Cookie(oauthStateCookie)
		if err != nil || cookie.Value == "" || cookie.Value != state {
			log.Printf("oauth google invalid state")
			writeJSON(w, http.StatusBadRequest, oauthErrorResponse{Error: "invalid_state"})
			return
		}

		code := r.URL.Query().Get("code")
		if code == "" {
			log.Printf("oauth google missing code")
			writeJSON(w, http.StatusBadRequest, oauthErrorResponse{Error: "missing_code"})
			return
		}

		token, err := conf.Exchange(r.Context(), code)
		if err != nil {
			log.Printf("oauth google exchange failed: %v", err)
			writeJSON(w, http.StatusBadRequest, oauthErrorResponse{Error: "exchange_failed", Detail: err.Error()})
			return
		}

		userInfo, err := fetchGoogleUserInfo(r, conf, token)
		if err != nil {
			log.Printf("oauth google userinfo failed: %v", err)
			writeJSON(w, http.StatusBadRequest, oauthErrorResponse{Error: "userinfo_failed", Detail: err.Error()})
			return
		}

		userID, err := ensureOAuthUser(r, db, userInfo)
		if err != nil {
			log.Printf("oauth google ensure user failed: %v", err)
			writeJSON(w, http.StatusInternalServerError, oauthErrorResponse{Error: "oauth_failed"})
			return
		}

		sessionToken, expiresAt, err := repo.CreateSession(r.Context(), db, userID, 7*24*time.Hour)
		if err != nil {
			log.Printf("oauth google session failed: %v", err)
			writeJSON(w, http.StatusInternalServerError, oauthErrorResponse{Error: "login_failed"})
			return
		}
		middleware.SetSessionCookie(cfg, w, sessionToken, expiresAt)

		http.SetCookie(w, &http.Cookie{
			Name:     oauthStateCookie,
			Value:    "",
			Path:     "/",
			HttpOnly: true,
			Secure:   cfg.CookieSecure,
			SameSite: http.SameSiteLaxMode,
			Expires:  time.Unix(0, 0),
			MaxAge:   -1,
		})

		redirectTo := cfg.FrontendURL
		if strings.TrimSpace(redirectTo) == "" {
			redirectTo = cfg.CORSOrigin
		}
		http.Redirect(w, r, redirectTo, http.StatusFound)
	}
}

func googleOAuthConfig(cfg config.Config) (*oauth2.Config, error) {
	if strings.TrimSpace(cfg.GoogleClientID) == "" || strings.TrimSpace(cfg.GoogleClientSecret) == "" || strings.TrimSpace(cfg.GoogleRedirectURL) == "" {
		return nil, errors.New("google oauth not configured")
	}
	return &oauth2.Config{
		ClientID:     cfg.GoogleClientID,
		ClientSecret: cfg.GoogleClientSecret,
		RedirectURL:  cfg.GoogleRedirectURL,
		Endpoint:     google.Endpoint,
		Scopes:       []string{"openid", "email", "profile"},
	}, nil
}

func fetchGoogleUserInfo(r *http.Request, conf *oauth2.Config, token *oauth2.Token) (googleUserInfo, error) {
	client := conf.Client(r.Context(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v3/userinfo")
	if err != nil {
		return googleUserInfo{}, errors.New("userinfo failed")
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return googleUserInfo{}, errors.New("userinfo failed")
	}
	var info googleUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return googleUserInfo{}, errors.New("userinfo invalid")
	}
	if strings.TrimSpace(info.Sub) == "" || strings.TrimSpace(info.Email) == "" {
		return googleUserInfo{}, errors.New("userinfo incomplete")
	}
	if info.EmailVerified == false {
		return googleUserInfo{}, errors.New("email not verified")
	}
	return info, nil
}

func ensureOAuthUser(r *http.Request, db *sql.DB, info googleUserInfo) (int64, error) {
	userID, found, err := repo.GetUserIDByOAuth(r.Context(), db, googleProvider, info.Sub)
	if err != nil {
		return 0, err
	}
	if found {
		return userID, nil
	}

	userID, exists, err := repo.GetUserIDByEmail(r.Context(), db, info.Email)
	if err != nil {
		return 0, err
	}
	if !exists {
		firstName, lastName := deriveNames(info)
		userID, err = repo.CreateOAuthUser(r.Context(), db, info.Email, firstName, lastName, stringPtr(info.Picture))
		if err != nil {
			return 0, err
		}
	}

	if err := repo.CreateOAuthAccount(r.Context(), db, userID, googleProvider, info.Sub, info.Email); err != nil {
		return 0, err
	}

	return userID, nil
}

func deriveNames(info googleUserInfo) (string, string) {
	first := strings.TrimSpace(info.GivenName)
	last := strings.TrimSpace(info.FamilyName)
	if first != "" || last != "" {
		return firstOrFallback(first, info), lastOrFallback(last, info)
	}

	full := strings.Fields(strings.TrimSpace(info.Name))
	if len(full) > 1 {
		return full[0], strings.Join(full[1:], " ")
	}
	if len(full) == 1 {
		return full[0], ""
	}

	local := strings.Split(strings.TrimSpace(info.Email), "@")
	if len(local) > 0 && local[0] != "" {
		return local[0], ""
	}
	return "User", ""
}

func firstOrFallback(value string, info googleUserInfo) string {
	if strings.TrimSpace(value) != "" {
		return strings.TrimSpace(value)
	}
	local := strings.Split(strings.TrimSpace(info.Email), "@")
	if len(local) > 0 && local[0] != "" {
		return local[0]
	}
	return "User"
}

func lastOrFallback(value string, info googleUserInfo) string {
	if strings.TrimSpace(value) != "" {
		return strings.TrimSpace(value)
	}
	return ""
}

func randomState(length int) (string, error) {
	buf := make([]byte, length)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func stringPtr(value string) *string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

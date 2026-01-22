package api_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"backend/internal/config"
	apphttp "backend/internal/http"
	"backend/internal/repo/sqlite"
)

func newTestServer(t *testing.T) (*httptest.Server, config.Config, *sql.DB) {
	t.Helper()
	tmp := t.TempDir()
	cfg := config.Config{
		Port:        "0",
		DBPath:      filepath.Join(tmp, "test.db"),
		MediaDir:    filepath.Join(tmp, "media"),
		CookieName:  "sid",
		CookieSecure: false,
		CORSOrigin:  "http://localhost:3000",
		Env:         "test",
	}
	db, err := sqlite.Open(cfg.DBPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("get cwd: %v", err)
	}
	if err := os.Chdir(filepath.Join(cwd, "..", "..")); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(cwd)
	})
	if err := applyBaseMigration(db); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	srv := httptest.NewServer(apphttp.NewRouter(cfg, db))
	return srv, cfg, db
}

func applyBaseMigration(db *sql.DB) error {
	path := filepath.Join("migrations", "sqlite", "000001_init.up.sql")
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	_, err = db.Exec(string(data))
	return err
}

func postJSON(t *testing.T, url string, body any, cookies []*http.Cookie) (*http.Response, []byte) {
	t.Helper()
	data, _ := json.Marshal(body)
	req, _ := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	for _, c := range cookies {
		req.AddCookie(c)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()
	buf, _ := io.ReadAll(resp.Body)
	return resp, buf
}

func patchJSON(t *testing.T, url string, body any, cookies []*http.Cookie) (*http.Response, []byte) {
	t.Helper()
	data, _ := json.Marshal(body)
	req, _ := http.NewRequest(http.MethodPatch, url, bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	for _, c := range cookies {
		req.AddCookie(c)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()
	buf, _ := io.ReadAll(resp.Body)
	return resp, buf
}

func registerUser(t *testing.T, baseURL, email string) {
	resp, _ := postJSON(t, baseURL+"/api/auth/register", map[string]any{
		"email": email,
		"password": "password123",
		"first_name": "Test",
		"last_name": "User",
		"dob": "1990-01-01",
	}, nil)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("register failed: %d", resp.StatusCode)
	}
}

func loginUser(t *testing.T, baseURL, email string) []*http.Cookie {
	resp, _ := postJSON(t, baseURL+"/api/auth/login", map[string]any{
		"email": email,
		"password": "password123",
	}, nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("login failed: %d", resp.StatusCode)
	}
	return resp.Cookies()
}

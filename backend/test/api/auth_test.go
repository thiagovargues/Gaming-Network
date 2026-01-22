package api_test

import (
	"net/http"
	"testing"
)

func TestAuthFlow(t *testing.T) {
	srv, _, db := newTestServer(t)
	defer srv.Close()
	defer db.Close()

	registerUser(t, srv.URL, "alice@example.com")
	cookies := loginUser(t, srv.URL, "alice@example.com")

	req, _ := http.NewRequest(http.MethodGet, srv.URL+"/api/me", nil)
	for _, c := range cookies {
		req.AddCookie(c)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("me failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("me status: %d", resp.StatusCode)
	}

	req, _ = http.NewRequest(http.MethodPost, srv.URL+"/api/auth/logout", nil)
	for _, c := range cookies {
		req.AddCookie(c)
	}
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("logout failed: %v", err)
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("logout status: %d", resp.StatusCode)
	}
}

package api_test

import (
	"net/http"
	"testing"
)

func TestFollowRequestFlow(t *testing.T) {
	srv, _, db := newTestServer(t)
	defer srv.Close()
	defer db.Close()

	registerUser(t, srv.URL, "alice@example.com")
	registerUser(t, srv.URL, "bob@example.com")

	aliceCookies := loginUser(t, srv.URL, "alice@example.com")
	bobCookies := loginUser(t, srv.URL, "bob@example.com")

	// Bob makes profile private
	resp, _ := patchJSON(t, srv.URL+"/api/users/me", map[string]any{"is_public": false}, bobCookies)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("privacy update failed: %d", resp.StatusCode)
	}

	// Alice requests follow
	resp, _ = postJSON(t, srv.URL+"/api/follows/request", map[string]any{"to_user_id": 2}, aliceCookies)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("follow request failed: %d", resp.StatusCode)
	}

	// Bob lists incoming requests
	req, _ := http.NewRequest(http.MethodGet, srv.URL+"/api/follows/requests/incoming", nil)
	for _, c := range bobCookies {
		req.AddCookie(c)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("incoming status: %d", resp.StatusCode)
	}
}

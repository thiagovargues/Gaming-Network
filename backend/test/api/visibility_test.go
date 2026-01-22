package api_test

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestPostVisibility(t *testing.T) {
	srv, _, db := newTestServer(t)
	defer srv.Close()
	defer db.Close()

	registerUser(t, srv.URL, "alice@example.com")
	registerUser(t, srv.URL, "bob@example.com")

	aliceCookies := loginUser(t, srv.URL, "alice@example.com")
	bobCookies := loginUser(t, srv.URL, "bob@example.com")

	// Bob follows Alice (public by default)
	resp, _ := postJSON(t, srv.URL+"/api/follows/request", map[string]any{"to_user_id": 1}, bobCookies)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("follow failed: %d", resp.StatusCode)
	}

	// Alice creates private post for Bob
	resp, _ = postJSON(t, srv.URL+"/api/posts", map[string]any{
		"text": "secret",
		"visibility": "private",
		"allowed_follower_ids": []int64{2},
	}, aliceCookies)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("post failed: %d", resp.StatusCode)
	}

	req, _ := http.NewRequest(http.MethodGet, srv.URL+"/api/feed", nil)
	for _, c := range bobCookies {
		req.AddCookie(c)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("feed failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("feed status: %d", resp.StatusCode)
	}
	var payload struct {
		Posts []struct {
			Text string `json:"text"`
		} `json:"posts"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(payload.Posts) == 0 {
		t.Fatalf("expected post")
	}
}

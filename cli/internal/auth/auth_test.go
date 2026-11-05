package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

// newVerifyServer mocks the API key check endpoint on the bridge domain.
// Only requests carrying validKey succeed with 204; everything else gets 401.
func newVerifyServer(t *testing.T, validKey string) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/api-keys/check", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-API-Key") != validKey {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
	return httptest.NewServer(mux)
}

func TestVerifyAPIKeyValid(t *testing.T) {
	srv := newVerifyServer(t, "devbox_good-key")
	defer srv.Close()

	if err := VerifyAPIKey(context.Background(), srv.URL, "devbox_good-key"); err != nil {
		t.Fatalf("expected valid key to pass, got: %v", err)
	}
}

func TestVerifyAPIKeyInvalid(t *testing.T) {
	srv := newVerifyServer(t, "devbox_good-key")
	defer srv.Close()

	err := VerifyAPIKey(context.Background(), srv.URL, "devbox_bad-key")
	if err == nil {
		t.Fatal("expected an error for an invalid key")
	}
	if !IsUnauthorized(err) {
		t.Fatalf("expected ErrUnauthorized, got: %v", err)
	}
}

func TestVerifyAPIKeyDevboxKey(t *testing.T) {
	srv := newVerifyServer(t, "devbox_good-key")
	defer srv.Close()

	if err := VerifyAPIKey(context.Background(), srv.URL, "devbox_good-key"); err != nil {
		t.Fatalf("expected valid devbox key to pass, got: %v", err)
	}
	if err := VerifyAPIKey(context.Background(), srv.URL, "devbox_bad-key"); !IsUnauthorized(err) {
		t.Fatalf("expected ErrUnauthorized for invalid devbox key, got: %v", err)
	}
}

func TestVerifyAPIKeyEmptyFailsFast(t *testing.T) {
	if err := VerifyAPIKey(context.Background(), "http://127.0.0.1:1", ""); !IsUnauthorized(err) {
		t.Fatalf("expected ErrUnauthorized for empty key, got: %v", err)
	}
}

func TestStartLoginWithCallback(t *testing.T) {
	session, err := StartLogin("https://devstation.connect.huaweicloud.com")
	if err != nil {
		t.Fatalf("start login: %v", err)
	}
	defer session.Close()

	if !strings.Contains(session.URL, "origin=devbox") {
		t.Errorf("login URL should carry the origin param: %s", session.URL)
	}

	u, err := url.Parse(session.URL)
	if err != nil {
		t.Fatalf("parse login URL: %v", err)
	}
	port := u.Query().Get("callback")
	if port == "" {
		t.Fatalf("login URL missing callback port: %s", session.URL)
	}

	callback := "http://127.0.0.1:" + port + "/"
	body := strings.NewReader(`{"apiKey":"callback-key-42","userName":"tester","userId":"u-1"}`)
	req, err := http.NewRequest(http.MethodPost, callback, body)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", "https://devstation.connect.huaweicloud.com")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("post callback: %v", err)
	}
	_ = resp.Body.Close()

	cred, ui, err := session.Wait(context.Background())
	if err != nil {
		t.Fatalf("wait: %v", err)
	}
	if cred.APIKey != "callback-key-42" {
		t.Fatalf("expected callback-key-42, got %q", cred.APIKey)
	}
	if ui == nil || ui.UserName != "tester" || ui.UserID != "u-1" {
		t.Fatalf("expected user info from the callback, got %+v", ui)
	}
}

func TestStartLoginRejectsMissingKey(t *testing.T) {
	session, err := StartLogin("https://devstation.connect.huaweicloud.com")
	if err != nil {
		t.Fatalf("start login: %v", err)
	}
	defer session.Close()

	u, _ := url.Parse(session.URL)
	port := u.Query().Get("callback")

	req, err := http.NewRequest(http.MethodPost, "http://127.0.0.1:"+port+"/",
		strings.NewReader(`{"error_code":"E1"}`))
	if err != nil {
		t.Fatalf("build request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", "https://devstation.connect.huaweicloud.com")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("post callback: %v", err)
	}
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing key, got %d", resp.StatusCode)
	}
}

// postCallback posts body with the given Origin header ("" omits it) and
// returns the response status.
func postCallback(t *testing.T, port, origin, body string) int {
	t.Helper()
	req, err := http.NewRequest(http.MethodPost, "http://127.0.0.1:"+port+"/",
		strings.NewReader(body))
	if err != nil {
		t.Fatalf("build request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if origin != "" {
		req.Header.Set("Origin", origin)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("post callback: %v", err)
	}
	_ = resp.Body.Close()
	return resp.StatusCode
}

// TestStartLoginRejectsForeignOrigin verifies the callback fails closed: only
// the login page origin is accepted; foreign or missing Origin headers get 403
// and never satisfy the wait (a planted-key CSRF is rejected).
func TestStartLoginRejectsForeignOrigin(t *testing.T) {
	cases := []struct {
		name   string
		origin string
	}{
		{"mismatched origin", "https://evil.example.com"},
		{"missing origin", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			session, err := StartLogin("https://devstation.connect.huaweicloud.com")
			if err != nil {
				t.Fatalf("start login: %v", err)
			}
			defer session.Close()

			u, _ := url.Parse(session.URL)
			port := u.Query().Get("callback")
			status := postCallback(t, port, tc.origin,
				`{"apiKey":"planted-key","userName":"evil","userId":"e-1"}`)
			if status != http.StatusForbidden {
				t.Fatalf("expected 403, got %d", status)
			}

			// The wait must not be satisfied by the rejected callback.
			ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
			defer cancel()
			if _, _, err := session.Wait(ctx); err == nil {
				t.Fatal("Wait must not succeed for a rejected callback")
			}
		})
	}
}

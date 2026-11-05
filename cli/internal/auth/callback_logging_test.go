package auth

import (
	"context"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"
)

func TestCallbackDebugLogging(t *testing.T) {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug})))

	session, err := StartLogin("https://devstation.connect.huaweicloud.com")
	if err != nil {
		t.Fatalf("start login: %v", err)
	}
	defer session.Close()

	u, _ := url.Parse(session.URL)
	port := u.Query().Get("callback")
	req, err := http.NewRequest(http.MethodPost, "http://127.0.0.1:"+port+"/",
		strings.NewReader(`{"apiKey":"callback-key-42","userName":"tester","userId":"u-1"}`))
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
	if err != nil || cred.APIKey != "callback-key-42" {
		t.Fatalf("wait: key=%q err=%v", cred.APIKey, err)
	}
	if ui == nil || ui.UserName != "tester" {
		t.Fatalf("expected user info from the callback, got %+v", ui)
	}
}

package usecases

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLocateLocal(t *testing.T) {
	if got := locate(context.Background(), "127.0.0.1"); got != "Local" {
		t.Fatalf("got %q", got)
	}
}

func TestLocateFormatsCity(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"city":"Tehran","country":"Iran"}`))
	}))
	t.Cleanup(server.Close)

	previous := geoLookupURL
	geoLookupURL = server.URL + "/"
	t.Cleanup(func() { geoLookupURL = previous })

	if got := locate(context.Background(), "8.8.8.8"); got != "Tehran, Iran" {
		t.Fatalf("got %q", got)
	}
}

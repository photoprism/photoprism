package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandleLoginRedirects(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/login?id=abc123", nil)
	w := httptest.NewRecorder()

	HandleLogin(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusFound {
		t.Fatalf("expected status %d, got %d", http.StatusFound, resp.StatusCode)
	}
	location := resp.Header.Get("Location")
	if location != "/authorize/callback?id=abc123:usertoken" {
		t.Fatalf("unexpected redirect location: %s", location)
	}
}

func TestHandleLoginBadRequest(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader("%zz"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	HandleLogin(w, req)

	if w.Result().StatusCode != http.StatusBadRequest {
		t.Fatalf("expected bad request, got %d", w.Result().StatusCode)
	}
}

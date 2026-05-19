package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/zitadel/oidc/v3/pkg/oidc"

	"caos-test-op/mock"
)

func newTestStorage(t *testing.T) (*mock.AuthStorage, string) {
	t.Helper()
	storage := mock.NewAuthStorage()
	req, err := storage.CreateAuthRequest(context.Background(), &oidc.AuthRequest{
		ClientID:     "csg6yqvykh0780f9",
		RedirectURI:  "https://app.localssl.dev/api/v1/oidc/redirect",
		ResponseType: oidc.ResponseTypeCode,
		Scopes:       []string{"openid", "email", "profile"},
	}, "")
	if err != nil {
		t.Fatalf("create auth request: %v", err)
	}
	return storage, req.GetID()
}

func TestHandleLoginRedirects(t *testing.T) {
	storage, id := newTestStorage(t)

	callback := func(_ context.Context, requestID string) string {
		return "/callback?id=" + requestID
	}

	req := httptest.NewRequest(http.MethodGet, "/login?id="+id, nil)
	w := httptest.NewRecorder()

	HandleLogin(w, req, storage, callback)

	resp := w.Result()
	if resp.StatusCode != http.StatusFound {
		t.Fatalf("expected status %d, got %d", http.StatusFound, resp.StatusCode)
	}
	if location := resp.Header.Get("Location"); location != "/callback?id="+id {
		t.Fatalf("unexpected redirect location: %s", location)
	}
}

func TestHandleLoginUnknownID(t *testing.T) {
	storage := mock.NewAuthStorage()

	callback := func(_ context.Context, requestID string) string {
		return "/callback?id=" + requestID
	}

	req := httptest.NewRequest(http.MethodGet, "/login?id=not-a-real-id", nil)
	w := httptest.NewRecorder()

	HandleLogin(w, req, storage, callback)

	if w.Result().StatusCode != http.StatusBadRequest {
		t.Fatalf("expected bad request for unknown id, got %d", w.Result().StatusCode)
	}
}

func TestHandleLoginMissingID(t *testing.T) {
	storage := mock.NewAuthStorage()

	req := httptest.NewRequest(http.MethodGet, "/login", nil)
	w := httptest.NewRecorder()

	HandleLogin(w, req, storage, func(context.Context, string) string { return "/callback" })

	if w.Result().StatusCode != http.StatusBadRequest {
		t.Fatalf("expected bad request for missing id, got %d", w.Result().StatusCode)
	}
}

func TestHandleLoginParseError(t *testing.T) {
	storage := mock.NewAuthStorage()

	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader("%zz"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	HandleLogin(w, req, storage, func(context.Context, string) string { return "/callback" })

	if w.Result().StatusCode != http.StatusBadRequest {
		t.Fatalf("expected bad request, got %d", w.Result().StatusCode)
	}
}

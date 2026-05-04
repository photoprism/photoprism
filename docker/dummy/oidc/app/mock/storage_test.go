package mock

import (
	"context"
	"testing"

	"github.com/zitadel/oidc/v3/pkg/oidc"
)

func TestAuthRequestResponseModeDefault(t *testing.T) {
	req := &AuthRequest{}
	if got := req.GetResponseMode(); got != oidc.ResponseModeQuery {
		t.Fatalf("expected default response mode %q, got %q", oidc.ResponseModeQuery, got)
	}
}

func TestRevokeTokenNoError(t *testing.T) {
	s := NewAuthStorage()
	if err := s.RevokeToken(context.Background(), "unknown-token", "user", "client"); err != nil {
		t.Fatalf("expected nil error from RevokeToken, got %v", err)
	}
}

func TestCreateAndFetchAuthRequest(t *testing.T) {
	s := NewAuthStorage()
	req, err := s.CreateAuthRequest(context.Background(), &oidc.AuthRequest{
		ClientID:     "web",
		RedirectURI:  "https://app.localssl.dev/api/v1/oidc/redirect",
		ResponseType: oidc.ResponseTypeCode,
		Scopes:       []string{"openid", "email"},
	}, "")
	if err != nil {
		t.Fatalf("CreateAuthRequest: %v", err)
	}
	if req.GetID() == "" {
		t.Fatalf("expected non-empty auth request id")
	}

	if err := s.MarkRequestDone(req.GetID()); err != nil {
		t.Fatalf("MarkRequestDone: %v", err)
	}

	got, err := s.AuthRequestByID(context.Background(), req.GetID())
	if err != nil {
		t.Fatalf("AuthRequestByID: %v", err)
	}
	if !got.Done() {
		t.Fatalf("expected auth request to be marked done")
	}
}

func TestNewClientDefaults(t *testing.T) {
	c := NewClient("anything")
	if c.GetID() != "anything" {
		t.Fatalf("unexpected id: %q", c.GetID())
	}
	if !c.DevMode() {
		t.Fatalf("expected dev mode for permissive dummy clients")
	}
	if uris := c.RedirectURIs(); len(uris) == 0 {
		t.Fatalf("expected default redirect URIs")
	}
}

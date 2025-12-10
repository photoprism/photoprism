package mock

import (
	"context"
	"testing"

	"github.com/zitadel/oidc/pkg/oidc"
)

func TestAuthRequestResponseModeDefault(t *testing.T) {
	req := &AuthRequest{}
	if got := req.GetResponseMode(); got != oidc.ResponseModeQuery {
		t.Fatalf("expected default response mode %q, got %q", oidc.ResponseModeQuery, got)
	}
}

func TestRevokeTokenNoError(t *testing.T) {
	s := &AuthStorage{}
	if err := s.RevokeToken(
		context.TODO(),
		"token",
		"user",
		"client",
	); err != nil {
		t.Fatalf("expected nil error from RevokeToken, got %v", err)
	}
}

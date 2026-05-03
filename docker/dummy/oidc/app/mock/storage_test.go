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

func TestAuthRequestResponseModeCustom(t *testing.T) {
	req := &AuthRequest{ResponseMode: oidc.ResponseModeFragment}
	if got := req.GetResponseMode(); got != oidc.ResponseModeFragment {
		t.Fatalf("expected response mode %q, got %q", oidc.ResponseModeFragment, got)
	}
}

func TestAuthRequestGetters(t *testing.T) {
	req := &AuthRequest{
		ID:          "test-id",
		ClientID:    "test-client",
		Nonce:       "test-nonce",
		RedirectURI: "https://example.com/callback",
	}

	if got := req.GetID(); got != "test-id" {
		t.Fatalf("expected ID %q, got %q", "test-id", got)
	}
	if got := req.GetClientID(); got != "test-client" {
		t.Fatalf("expected ClientID %q, got %q", "test-client", got)
	}
	if got := req.GetNonce(); got != "test-nonce" {
		t.Fatalf("expected Nonce %q, got %q", "test-nonce", got)
	}
	if got := req.GetRedirectURI(); got != "https://example.com/callback" {
		t.Fatalf("expected RedirectURI %q, got %q", "https://example.com/callback", got)
	}
	if got := req.GetSubject(); got != "sub00000001" {
		t.Fatalf("expected Subject %q, got %q", "sub00000001", got)
	}
	if got := req.GetACR(); got != "" {
		t.Fatalf("expected empty ACR, got %q", got)
	}
	if len(req.GetAMR()) != 0 {
		t.Fatalf("expected empty AMR, got %v", req.GetAMR())
	}
	if !req.Done() {
		t.Fatal("expected Done() to return true")
	}
}

func TestAuthRequestAudience(t *testing.T) {
	req := &AuthRequest{ClientID: "my-client"}
	aud := req.GetAudience()
	if len(aud) != 1 || aud[0] != "my-client" {
		t.Fatalf("expected audience [my-client], got %v", aud)
	}
}

func TestAuthRequestScopes(t *testing.T) {
	req := &AuthRequest{}
	scopes := req.GetScopes()
	expected := []string{"openid", "profile", "email"}
	if len(scopes) != len(expected) {
		t.Fatalf("expected %d scopes, got %d", len(expected), len(scopes))
	}
	for i, s := range expected {
		if scopes[i] != s {
			t.Fatalf("expected scope %q at index %d, got %q", s, i, scopes[i])
		}
	}
}

func TestNewAuthStorage(t *testing.T) {
	storage := NewAuthStorage()
	if storage == nil {
		t.Fatal("expected non-nil storage")
	}
	as, ok := storage.(*AuthStorage)
	if !ok {
		t.Fatal("expected *AuthStorage type")
	}
	if as.key == nil {
		t.Fatal("expected non-nil RSA key")
	}
	if as.kid == "" {
		t.Fatal("expected non-empty kid")
	}
}

func TestAuthStorageHealth(t *testing.T) {
	storage := NewAuthStorage().(*AuthStorage)
	if err := storage.Health(context.Background()); err != nil {
		t.Fatalf("expected nil from Health(), got %v", err)
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

func TestAuthStorageGetKeySet(t *testing.T) {
	storage := NewAuthStorage().(*AuthStorage)
	keySet, err := storage.GetKeySet(context.Background())
	if err != nil {
		t.Fatalf("unexpected error from GetKeySet: %v", err)
	}
	if keySet == nil {
		t.Fatal("expected non-nil key set")
	}
	if len(keySet.Keys) != 2 {
		t.Fatalf("expected 2 keys, got %d", len(keySet.Keys))
	}
}

func TestAuthStorageGetKey(t *testing.T) {
	storage := NewAuthStorage().(*AuthStorage)
	key, err := storage.GetKey(context.Background())
	if err != nil {
		t.Fatalf("unexpected error from GetKey: %v", err)
	}
	if key == nil {
		t.Fatal("expected non-nil key")
	}
}

func TestAuthStorageCreateAccessToken(t *testing.T) {
	storage := NewAuthStorage().(*AuthStorage)
	tokenID, expiration, err := storage.CreateAccessToken(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tokenID != "loginId" {
		t.Fatalf("expected tokenID 'loginId', got %q", tokenID)
	}
	if expiration.IsZero() {
		t.Fatal("expected non-zero expiration")
	}
}

func TestAuthStorageCreateAccessAndRefreshTokens(t *testing.T) {
	storage := NewAuthStorage().(*AuthStorage)
	accessID, refresh, expiration, err := storage.CreateAccessAndRefreshTokens(context.Background(), nil, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if accessID != "loginId" {
		t.Fatalf("expected accessID 'loginId', got %q", accessID)
	}
	if refresh != "refreshToken" {
		t.Fatalf("expected refresh 'refreshToken', got %q", refresh)
	}
	if expiration.IsZero() {
		t.Fatal("expected non-zero expiration")
	}
}

func TestAuthStorageTerminateSession(t *testing.T) {
	storage := NewAuthStorage().(*AuthStorage)
	if err := storage.TerminateSession(context.Background(), "user", "client"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAuthStorageAuthorizeClientIDSecret(t *testing.T) {
	storage := NewAuthStorage().(*AuthStorage)
	if err := storage.AuthorizeClientIDSecret(context.Background(), "client", "secret"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAuthStorageValidateJWTProfileScopes(t *testing.T) {
	storage := NewAuthStorage().(*AuthStorage)
	scopes := []string{"openid", "profile"}
	result, err := storage.ValidateJWTProfileScopes(context.Background(), "user", scopes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != len(scopes) {
		t.Fatalf("expected %d scopes, got %d", len(scopes), len(result))
	}
}

func TestAuthStorageGetPrivateClaimsFromScopes(t *testing.T) {
	storage := NewAuthStorage().(*AuthStorage)
	claims, err := storage.GetPrivateClaimsFromScopes(context.Background(), "", "", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if claims["private_claim"] != "test" {
		t.Fatalf("expected private_claim 'test', got %v", claims["private_claim"])
	}
}

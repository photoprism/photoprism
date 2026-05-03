package mock

import (
	"testing"
	"time"

	"github.com/zitadel/oidc/pkg/oidc"
	"github.com/zitadel/oidc/pkg/op"
)

func TestConfClientGetID(t *testing.T) {
	c := &ConfClient{ID: "test-client"}
	if got := c.GetID(); got != "test-client" {
		t.Fatalf("expected ID 'test-client', got %q", got)
	}
}

func TestConfClientRedirectURIs(t *testing.T) {
	c := &ConfClient{}
	uris := c.RedirectURIs()
	if len(uris) == 0 {
		t.Fatal("expected non-empty redirect URIs")
	}
	// Check that localhost:2342 PhotoPrism callback is included
	found := false
	for _, uri := range uris {
		if uri == "http://localhost:2342/api/v1/oidc/redirect" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected PhotoPrism OIDC redirect URI in list")
	}
}

func TestConfClientPostLogoutRedirectURIs(t *testing.T) {
	c := &ConfClient{}
	uris := c.PostLogoutRedirectURIs()
	if len(uris) != 0 {
		t.Fatalf("expected empty post-logout redirect URIs, got %v", uris)
	}
}

func TestConfClientLoginURL(t *testing.T) {
	c := &ConfClient{}
	url := c.LoginURL("abc123")
	expected := "login?id=abc123"
	if url != expected {
		t.Fatalf("expected login URL %q, got %q", expected, url)
	}
}

func TestConfClientApplicationType(t *testing.T) {
	tests := []struct {
		appType  op.ApplicationType
		expected op.ApplicationType
	}{
		{op.ApplicationTypeWeb, op.ApplicationTypeWeb},
		{op.ApplicationTypeNative, op.ApplicationTypeNative},
		{op.ApplicationTypeUserAgent, op.ApplicationTypeUserAgent},
	}
	for _, tt := range tests {
		c := &ConfClient{applicationType: tt.appType}
		if got := c.ApplicationType(); got != tt.expected {
			t.Fatalf("expected ApplicationType %v, got %v", tt.expected, got)
		}
	}
}

func TestConfClientAuthMethod(t *testing.T) {
	c := &ConfClient{authMethod: oidc.AuthMethodBasic}
	if got := c.AuthMethod(); got != oidc.AuthMethodBasic {
		t.Fatalf("expected AuthMethod %v, got %v", oidc.AuthMethodBasic, got)
	}
}

func TestConfClientIDTokenLifetime(t *testing.T) {
	c := &ConfClient{}
	expected := 60 * time.Minute
	if got := c.IDTokenLifetime(); got != expected {
		t.Fatalf("expected IDTokenLifetime %v, got %v", expected, got)
	}
}

func TestConfClientAccessTokenType(t *testing.T) {
	c := &ConfClient{accessTokenType: op.AccessTokenTypeJWT}
	if got := c.AccessTokenType(); got != op.AccessTokenTypeJWT {
		t.Fatalf("expected AccessTokenType %v, got %v", op.AccessTokenTypeJWT, got)
	}
}

func TestConfClientResponseTypes(t *testing.T) {
	expected := []oidc.ResponseType{oidc.ResponseTypeCode}
	c := &ConfClient{responseTypes: expected}
	got := c.ResponseTypes()
	if len(got) != len(expected) {
		t.Fatalf("expected %d response types, got %d", len(expected), len(got))
	}
}

func TestConfClientGrantTypes(t *testing.T) {
	expected := []oidc.GrantType{oidc.GrantTypeCode}
	c := &ConfClient{grantTypes: expected}
	got := c.GrantTypes()
	if len(got) != len(expected) {
		t.Fatalf("expected %d grant types, got %d", len(expected), len(got))
	}
}

func TestConfClientDevMode(t *testing.T) {
	c := &ConfClient{devMode: true}
	if !c.DevMode() {
		t.Fatal("expected DevMode to be true")
	}
	c.devMode = false
	if c.DevMode() {
		t.Fatal("expected DevMode to be false")
	}
}

func TestConfClientAllowedScopes(t *testing.T) {
	c := &ConfClient{}
	if got := c.AllowedScopes(); got != nil {
		t.Fatalf("expected nil AllowedScopes, got %v", got)
	}
}

func TestConfClientRestrictAdditionalIdTokenScopes(t *testing.T) {
	c := &ConfClient{}
	fn := c.RestrictAdditionalIdTokenScopes()
	scopes := []string{"openid", "profile"}
	got := fn(scopes)
	if len(got) != len(scopes) {
		t.Fatalf("expected same scopes returned, got %v", got)
	}
}

func TestConfClientRestrictAdditionalAccessTokenScopes(t *testing.T) {
	c := &ConfClient{}
	fn := c.RestrictAdditionalAccessTokenScopes()
	scopes := []string{"openid", "profile"}
	got := fn(scopes)
	if len(got) != len(scopes) {
		t.Fatalf("expected same scopes returned, got %v", got)
	}
}

func TestConfClientIsScopeAllowed(t *testing.T) {
	c := &ConfClient{}
	if c.IsScopeAllowed("openid") {
		t.Fatal("expected IsScopeAllowed to return false")
	}
}

func TestConfClientIDTokenUserinfoClaimsAssertion(t *testing.T) {
	c := &ConfClient{}
	if c.IDTokenUserinfoClaimsAssertion() {
		t.Fatal("expected IDTokenUserinfoClaimsAssertion to return false")
	}
}

func TestConfClientClockSkew(t *testing.T) {
	c := &ConfClient{}
	if got := c.ClockSkew(); got != 0 {
		t.Fatalf("expected ClockSkew 0, got %v", got)
	}
}

package jwt

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"
	"time"

	gojwt "github.com/golang-jwt/jwt/v5"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/auth/acl"
)

// userTokenSetup spins up an in-memory key manager and issuer with a frozen
// clock so token timestamps are deterministic across assertions.
func userTokenSetup(t *testing.T) (*Issuer, *Manager) {
	cfg := newTestConfig(t)
	mgr, err := NewManager(cfg)
	require.NoError(t, err)
	mgr.now = func() time.Time { return time.Date(2026, 5, 24, 12, 0, 0, 0, time.UTC) }
	_, err = mgr.EnsureActiveKey()
	require.NoError(t, err)

	iss := NewIssuer(mgr)
	iss.now = func() time.Time { return time.Date(2026, 5, 24, 12, 0, 0, 0, time.UTC) }
	return iss, mgr
}

// decodeUserClaims parses a Portal-issued JWT against the manager's public
// key and returns the typed UserClaims for assertions.
func decodeUserClaims(t *testing.T, mgr *Manager, token string) *UserClaims {
	t.Helper()
	parsed := &UserClaims{}
	parser := gojwt.NewParser(
		gojwt.WithValidMethods([]string{gojwt.SigningMethodEdDSA.Alg()}),
		gojwt.WithoutClaimsValidation(),
	)
	_, err := parser.ParseWithClaims(token, parsed, func(*gojwt.Token) (any, error) {
		key, _ := mgr.ActiveKey()
		return key.PublicKey, nil
	})
	require.NoError(t, err)
	return parsed
}

// decodeUserPayload decodes the JWT payload as a generic map so tests can
// assert on the raw JSON shape (presence/absence of optional claims under
// omitempty rules).
func decodeUserPayload(t *testing.T, token string) map[string]any {
	t.Helper()
	parts := strings.Split(token, ".")
	require.Len(t, parts, 3, "JWT must have three segments")
	raw, err := base64.RawURLEncoding.DecodeString(parts[1])
	require.NoError(t, err)
	out := map[string]any{}
	require.NoError(t, json.Unmarshal(raw, &out))
	return out
}

func validSpec() UserClaimsSpec {
	return UserClaimsSpec{
		Issuer:   "https://portal.example.com/",
		Subject:  "uxxxxxxxxxxxxxxx",
		Audience: "cs5gfen1bgxz7s9i",
		NodeUUID: "0192abcd-1234-7000-8000-000000000001",
		Role:     "admin",
	}
}

func TestIssueUser_Success(t *testing.T) {
	iss, mgr := userTokenSetup(t)

	spec := validSpec()
	spec.Nonce = "nonce-abc"
	spec.AuthTime = time.Date(2026, 5, 24, 11, 0, 0, 0, time.UTC)

	token, err := iss.IssueUser(spec)
	require.NoError(t, err)

	got := decodeUserClaims(t, mgr, token)
	assert.Equal(t, spec.Issuer, got.Issuer)
	assert.Equal(t, spec.Subject, got.Subject)
	assert.Contains(t, got.Audience, spec.Audience)
	assert.Equal(t, "nonce-abc", got.Nonce)
	assert.Equal(t, spec.AuthTime.Unix(), got.AuthTime)
	assert.Equal(t, "admin", got.PortalRole)
	assert.Equal(t, spec.NodeUUID, got.PortalNodeUUID)
	assert.Equal(t, acl.RolePortal.String(), got.PortalIssuerKind)
	require.NotNil(t, got.IssuedAt)
	require.NotNil(t, got.ExpiresAt)
	assert.Equal(t, TokenTTL, got.ExpiresAt.Sub(got.IssuedAt.Time))
}

func TestIssueUser_Validation(t *testing.T) {
	iss, _ := userTokenSetup(t)

	cases := []struct {
		name string
		mut  func(*UserClaimsSpec)
		err  string
	}{
		{"MissingIssuer", func(s *UserClaimsSpec) { s.Issuer = "" }, "jwt: issuer required"},
		{"MissingSubject", func(s *UserClaimsSpec) { s.Subject = "" }, "jwt: subject required"},
		{"MissingAudience", func(s *UserClaimsSpec) { s.Audience = "" }, "jwt: audience required"},
		{"MissingNodeUUID", func(s *UserClaimsSpec) { s.NodeUUID = "" }, "jwt: node uuid required"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := validSpec()
			tc.mut(&s)
			_, err := iss.IssueUser(s)
			assert.EqualError(t, err, tc.err)
		})
	}
}

func TestIssueUser_TTLClamp(t *testing.T) {
	iss, mgr := userTokenSetup(t)

	spec := validSpec()
	spec.TTL = MaxTokenTTL * 4

	token, err := iss.IssueUser(spec)
	require.NoError(t, err)

	got := decodeUserClaims(t, mgr, token)
	assert.Equal(t, MaxTokenTTL, got.ExpiresAt.Sub(got.IssuedAt.Time))
}

func TestIssueUser_DefaultTTL(t *testing.T) {
	iss, mgr := userTokenSetup(t)

	// Zero TTL falls back to the package-level TokenTTL default.
	token, err := iss.IssueUser(validSpec())
	require.NoError(t, err)

	got := decodeUserClaims(t, mgr, token)
	assert.Equal(t, TokenTTL, got.ExpiresAt.Sub(got.IssuedAt.Time))
}

func TestIssueUser_OptionalProfileClaimsOmitted(t *testing.T) {
	iss, _ := userTokenSetup(t)

	// Caller has not requested any profile-scope claims.
	token, err := iss.IssueUser(validSpec())
	require.NoError(t, err)

	payload := decodeUserPayload(t, token)
	for _, k := range []string{
		"email", "email_verified", "name", "given_name", "family_name",
		"preferred_username", "locale", "zoneinfo", "picture", "groups", "auth_time", "nonce",
	} {
		_, present := payload[k]
		assert.Falsef(t, present, "claim %q must be omitted when caller did not set it", k)
	}
}

func TestIssueUser_EmailVerifiedExplicit(t *testing.T) {
	iss, mgr := userTokenSetup(t)

	t.Run("True", func(t *testing.T) {
		spec := validSpec()
		spec.Email = "alice@example.com"
		yes := true
		spec.EmailVerified = &yes

		token, err := iss.IssueUser(spec)
		require.NoError(t, err)
		got := decodeUserClaims(t, mgr, token)
		require.NotNil(t, got.EmailVerified)
		assert.True(t, *got.EmailVerified)
	})

	t.Run("FalseEmittedExplicitly", func(t *testing.T) {
		// Pointer semantics: an explicit &false MUST land in the token as
		// `email_verified: false`, not be silently dropped (which is the
		// trap with bool + omitempty).
		spec := validSpec()
		spec.Email = "alice@example.com"
		no := false
		spec.EmailVerified = &no

		token, err := iss.IssueUser(spec)
		require.NoError(t, err)

		payload := decodeUserPayload(t, token)
		verified, present := payload["email_verified"]
		require.True(t, present, "email_verified must be present when caller sets &false")
		assert.Equal(t, false, verified)
	})

	t.Run("NilOmitted", func(t *testing.T) {
		spec := validSpec()
		spec.Email = "alice@example.com" // email set but verified intentionally nil

		token, err := iss.IssueUser(spec)
		require.NoError(t, err)

		payload := decodeUserPayload(t, token)
		_, present := payload["email_verified"]
		assert.False(t, present, "email_verified must be omitted when caller leaves it nil")
	})
}

func TestIssueUser_GroupsEmission(t *testing.T) {
	iss, mgr := userTokenSetup(t)

	spec := validSpec()
	spec.Groups = []string{"media-acme-admin", "media-acme-viewer"}

	token, err := iss.IssueUser(spec)
	require.NoError(t, err)

	got := decodeUserClaims(t, mgr, token)
	assert.Equal(t, []string{"media-acme-admin", "media-acme-viewer"}, got.Groups)
}

func TestIssueUser_SignatureVerifiesAgainstJWKS(t *testing.T) {
	iss, mgr := userTokenSetup(t)

	spec := validSpec()
	token, err := iss.IssueUser(spec)
	require.NoError(t, err)

	keys := mgr.JWKS().Keys
	// WithoutClaimsValidation skips the temporal nbf/exp checks because the
	// issuer's frozen clock isn't aligned with wall-clock time; the point of
	// this test is signature + JWKS resolution, not freshness. Issuer and
	// audience are asserted explicitly on the parsed struct below.
	parser := gojwt.NewParser(
		gojwt.WithValidMethods([]string{gojwt.SigningMethodEdDSA.Alg()}),
		gojwt.WithoutClaimsValidation(),
	)

	parsed := &UserClaims{}
	_, err = parser.ParseWithClaims(token, parsed, func(tok *gojwt.Token) (any, error) {
		kid, _ := tok.Header["kid"].(string)
		for _, k := range keys {
			if k.Kid == kid {
				raw, decodeErr := base64.RawURLEncoding.DecodeString(k.X)
				require.NoError(t, decodeErr)
				return ed25519.PublicKey(raw), nil
			}
		}
		t.Fatalf("no matching kid %s in JWKS", kid)
		return nil, nil
	})
	require.NoError(t, err)
	assert.Equal(t, spec.Subject, parsed.Subject)
	assert.Equal(t, spec.Issuer, parsed.Issuer)
	assert.Contains(t, parsed.Audience, spec.Audience)
}

func TestIssueUser_NilIssuer(t *testing.T) {
	var iss *Issuer
	_, err := iss.IssueUser(validSpec())
	assert.EqualError(t, err, "jwt: issuer not initialized")
}

func TestIssueUser_IssuerWithoutManager(t *testing.T) {
	iss := &Issuer{}
	_, err := iss.IssueUser(validSpec())
	assert.EqualError(t, err, "jwt: issuer not initialized")
}

func TestUserClaimsSpec_Validate(t *testing.T) {
	// Direct coverage of the unexported validate() helper because the rule
	// requires focused tests for new helpers, not just integration coverage
	// through IssueUser.
	t.Run("Valid", func(t *testing.T) {
		assert.NoError(t, validSpec().validate())
	})
	t.Run("TrimsWhitespaceOnlyValues", func(t *testing.T) {
		s := validSpec()
		s.Issuer = "   "
		assert.EqualError(t, s.validate(), "jwt: issuer required")
	})
}

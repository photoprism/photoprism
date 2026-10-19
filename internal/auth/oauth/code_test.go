package oauth

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// testCodeSpec returns a valid code spec.
func testCodeSpec() CodeSpec {
	return CodeSpec{
		ClientID:            "cs5gfen1bgxz7s9i",
		ClientName:          "Test App",
		UserUID:             "uqxetse3cy5eo9z2",
		RedirectURI:         "https://claude.ai/api/mcp/auth_callback",
		Scope:               "mcp",
		Resource:            "https://photos.example.com/api/v1/mcp",
		CodeChallenge:       testChallenge(testVerifier),
		CodeChallengeMethod: PKCEMethodS256,
	}
}

func TestIssueCode(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		raw, err := IssueCode(testCodeSpec())
		require.NoError(t, err)
		assert.Len(t, raw, 48)

		m, err := entity.FindAuthCode(HashSecret(raw), entity.AuthCodeCode)
		require.NoError(t, err)
		require.NotNil(t, m)
		assert.Equal(t, "uqxetse3cy5eo9z2", m.UserUID)
		assert.NotEqual(t, raw, m.CodeHash)
		assert.WithinDuration(t, time.Now().UTC().Add(CodeTTL), m.ExpiresAt, 5*time.Second)
		assert.False(t, m.Expired())
	})
	t.Run("LongClientName", func(t *testing.T) {
		spec := testCodeSpec()
		spec.ClientName = strings.Repeat("x", 300)
		raw, err := IssueCode(spec)
		require.NoError(t, err)
		m, err := entity.FindAuthCode(HashSecret(raw), entity.AuthCodeCode)
		require.NoError(t, err)
		require.NotNil(t, m)
		assert.Len(t, m.ClientName, 200)
	})
	cases := map[string]func(s *CodeSpec){
		"NoClient":      func(s *CodeSpec) { s.ClientID = "" },
		"NoUser":        func(s *CodeSpec) { s.UserUID = " " },
		"NoRedirectURI": func(s *CodeSpec) { s.RedirectURI = "" },
		"NoResource":    func(s *CodeSpec) { s.Resource = "" },
		"PlainMethod":   func(s *CodeSpec) { s.CodeChallengeMethod = "plain" },
		"NoChallenge":   func(s *CodeSpec) { s.CodeChallenge = "" },
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			spec := testCodeSpec()
			mutate(&spec)
			raw, err := IssueCode(spec)
			assert.Error(t, err)
			assert.Empty(t, raw)
		})
	}
}

func TestCodeSpec_Validate(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		assert.NoError(t, testCodeSpec().validate())
	})
	t.Run("InvalidChallenge", func(t *testing.T) {
		spec := testCodeSpec()
		spec.CodeChallenge = "abc"
		assert.Error(t, spec.validate())
	})
}

func TestRedeemCode(t *testing.T) {
	spec := testCodeSpec()

	t.Run("Success", func(t *testing.T) {
		raw, err := IssueCode(spec)
		require.NoError(t, err)

		m, err := RedeemCode(raw, spec.ClientID, spec.RedirectURI, testVerifier)
		require.NoError(t, err)
		assert.Equal(t, spec.UserUID, m.UserUID)
		assert.Equal(t, spec.Scope, m.Scope)
		assert.Equal(t, spec.Resource, m.Resource)

		// A code can be redeemed only once.
		_, err = RedeemCode(raw, spec.ClientID, spec.RedirectURI, testVerifier)
		assert.ErrorIs(t, err, ErrInvalidGrant)
	})
	t.Run("OtherClient", func(t *testing.T) {
		raw, err := IssueCode(spec)
		require.NoError(t, err)
		_, err = RedeemCode(raw, "cs5cpu17n6gj2qo5", spec.RedirectURI, testVerifier)
		assert.ErrorIs(t, err, ErrInvalidGrant)

		// The failed attempt consumed the code.
		_, err = RedeemCode(raw, spec.ClientID, spec.RedirectURI, testVerifier)
		assert.ErrorIs(t, err, ErrInvalidGrant)
	})
	t.Run("OtherRedirectURI", func(t *testing.T) {
		raw, err := IssueCode(spec)
		require.NoError(t, err)
		_, err = RedeemCode(raw, spec.ClientID, "https://claude.ai/other", testVerifier)
		assert.ErrorIs(t, err, ErrInvalidGrant)
		_, err = RedeemCode(raw, spec.ClientID, spec.RedirectURI, testVerifier)
		assert.ErrorIs(t, err, ErrInvalidGrant)
	})
	t.Run("WrongVerifier", func(t *testing.T) {
		raw, err := IssueCode(spec)
		require.NoError(t, err)
		_, err = RedeemCode(raw, spec.ClientID, spec.RedirectURI, testVerifier+"x")
		assert.ErrorIs(t, err, ErrInvalidGrant)
		_, err = RedeemCode(raw, spec.ClientID, spec.RedirectURI, testVerifier)
		assert.ErrorIs(t, err, ErrInvalidGrant)
	})
	t.Run("NoUser", func(t *testing.T) {
		raw := rnd.AuthToken()
		m := &entity.AuthCode{CodeHash: HashSecret(raw), CodeType: entity.AuthCodeCode, ClientID: spec.ClientID, RedirectURI: spec.RedirectURI,
			Resource: spec.Resource, CodeChallenge: spec.CodeChallenge, CodeChallengeMethod: spec.CodeChallengeMethod, ExpiresAt: time.Now().UTC().Add(time.Minute)}
		require.NoError(t, m.Create())
		_, err := RedeemCode(raw, spec.ClientID, spec.RedirectURI, testVerifier)
		assert.ErrorIs(t, err, ErrInvalidGrant)
	})
	t.Run("Expired", func(t *testing.T) {
		raw, err := IssueCode(spec)
		require.NoError(t, err)
		require.NoError(t, entity.UnscopedDb().Model(&entity.AuthCode{}).Where("code_hash = ?", HashSecret(raw)).
			UpdateColumn("expires_at", time.Now().UTC().Add(-time.Second)).Error)
		_, err = RedeemCode(raw, spec.ClientID, spec.RedirectURI, testVerifier)
		assert.ErrorIs(t, err, ErrInvalidGrant)
	})
	t.Run("Unknown", func(t *testing.T) {
		_, err := RedeemCode("0123456789abcdef0123456789abcdef0123456789abcdef", spec.ClientID, spec.RedirectURI, testVerifier)
		assert.ErrorIs(t, err, ErrInvalidGrant)
	})
	t.Run("Empty", func(t *testing.T) {
		_, err := RedeemCode("", spec.ClientID, spec.RedirectURI, testVerifier)
		assert.ErrorIs(t, err, ErrInvalidGrant)
	})
	t.Run("RequestIDIsNotACode", func(t *testing.T) {
		raw := "0123456789abcdef0123456789abcdef0123456789abcdee"
		m := &entity.AuthCode{CodeHash: HashSecret(raw), CodeType: entity.AuthCodeRequest, ClientID: spec.ClientID, RedirectURI: spec.RedirectURI,
			CodeChallenge: spec.CodeChallenge, CodeChallengeMethod: spec.CodeChallengeMethod, ExpiresAt: time.Now().UTC().Add(time.Minute)}
		require.NoError(t, m.Create())
		_, err := RedeemCode(raw, spec.ClientID, spec.RedirectURI, testVerifier)
		assert.ErrorIs(t, err, ErrInvalidGrant)
	})
}

func TestPurgeExpired(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		raw, err := IssueCode(testCodeSpec())
		require.NoError(t, err)
		require.NoError(t, entity.UnscopedDb().Model(&entity.AuthCode{}).Where("code_hash = ?", HashSecret(raw)).
			UpdateColumn("expires_at", time.Now().UTC().Add(-time.Minute)).Error)

		PurgeExpired()

		m, err := entity.FindAuthCode(HashSecret(raw), entity.AuthCodeCode)
		assert.NoError(t, err)
		assert.Nil(t, m)
	})
}

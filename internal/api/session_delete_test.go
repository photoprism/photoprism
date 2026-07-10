package api

import (
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zitadel/oidc/v3/pkg/client/rp"
	"golang.org/x/oauth2"

	"github.com/photoprism/photoprism/internal/auth/oidc"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/authn"
)

// fakeRelyingParty implements only the rp.RelyingParty methods EndSessionURL reads.
type fakeRelyingParty struct {
	rp.RelyingParty
	endSession string
	clientID   string
}

func (f fakeRelyingParty) GetEndSessionEndpoint() string { return f.endSession }
func (f fakeRelyingParty) OAuthConfig() *oauth2.Config   { return &oauth2.Config{ClientID: f.clientID} }

func oidcProvider(endSession string) *oidc.Client {
	return &oidc.Client{RelyingParty: fakeRelyingParty{endSession: endSession, clientID: "client-abc"}}
}

func oidcSession(provider authn.ProviderType, idToken string) *entity.Session {
	s := &entity.Session{}
	s.SetProvider(provider)
	s.IdToken = idToken
	return s
}

func TestOidcLogoutURL(t *testing.T) {
	conf := config.NewConfig(config.CliTestContext())
	provider := oidcProvider("https://provider.example/logout")

	t.Run("Success", func(t *testing.T) {
		conf.Options().OIDCLogout = true
		defer func() { conf.Options().OIDCLogout = false }()

		result := oidcLogoutURL(conf, provider, oidcSession(authn.ProviderOIDC, "id-token-123"))
		require.NotEmpty(t, result)

		u, err := url.Parse(result)
		require.NoError(t, err)
		assert.Equal(t, "id-token-123", u.Query().Get("id_token_hint"))
		assert.Equal(t, "client-abc", u.Query().Get("client_id"))
		// post_logout_redirect_uri must be absolute (providers reject relative paths).
		plr := u.Query().Get("post_logout_redirect_uri")
		assert.Equal(t, AbsoluteLoginURL(conf), plr)
		assert.Contains(t, plr, "://")
	})
	t.Run("FlagDisabled", func(t *testing.T) {
		conf.Options().OIDCLogout = false
		assert.Equal(t, "", oidcLogoutURL(conf, provider, oidcSession(authn.ProviderOIDC, "id-token-123")))
	})
	t.Run("NotOIDCSession", func(t *testing.T) {
		conf.Options().OIDCLogout = true
		defer func() { conf.Options().OIDCLogout = false }()
		assert.Equal(t, "", oidcLogoutURL(conf, provider, oidcSession(authn.ProviderLocal, "id-token-123")))
	})
	t.Run("NoIdToken", func(t *testing.T) {
		conf.Options().OIDCLogout = true
		defer func() { conf.Options().OIDCLogout = false }()
		assert.Equal(t, "", oidcLogoutURL(conf, provider, oidcSession(authn.ProviderOIDC, "")))
	})
	t.Run("ProviderWithoutEndSession", func(t *testing.T) {
		conf.Options().OIDCLogout = true
		defer func() { conf.Options().OIDCLogout = false }()
		assert.Equal(t, "", oidcLogoutURL(conf, oidcProvider(""), oidcSession(authn.ProviderOIDC, "id-token-123")))
	})
	t.Run("NilProvider", func(t *testing.T) {
		conf.Options().OIDCLogout = true
		defer func() { conf.Options().OIDCLogout = false }()
		assert.Equal(t, "", oidcLogoutURL(conf, nil, oidcSession(authn.ProviderOIDC, "id-token-123")))
	})
}

func TestAbsoluteLoginURL(t *testing.T) {
	t.Run("RelativeLoginPathResolvedToOrigin", func(t *testing.T) {
		conf := config.NewConfig(config.CliTestContext())
		result := AbsoluteLoginURL(conf)
		assert.Contains(t, result, "://", "must be absolute")
		assert.True(t, strings.HasSuffix(result, conf.LoginUri()), "must end with the login path")
	})
	t.Run("AbsoluteLoginUriPassedThrough", func(t *testing.T) {
		conf := config.NewConfig(config.CliTestContext())
		conf.Options().LoginUri = "https://sso.example.com/login"
		assert.Equal(t, "https://sso.example.com/login", AbsoluteLoginURL(conf))
	})
}

package oidc

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zitadel/oidc/v3/pkg/client/rp"
	"golang.org/x/oauth2"
)

// fakeRelyingParty implements only the rp.RelyingParty methods EndSessionURL reads; any
// other call panics on the nil embedded interface, which keeps the fake honest.
type fakeRelyingParty struct {
	rp.RelyingParty
	endSession string
	clientID   string
}

func (f fakeRelyingParty) GetEndSessionEndpoint() string { return f.endSession }
func (f fakeRelyingParty) OAuthConfig() *oauth2.Config   { return &oauth2.Config{ClientID: f.clientID} }

func TestClient_EndSessionURL(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		c := &Client{RelyingParty: fakeRelyingParty{
			endSession: "https://provider.example/logout",
			clientID:   "client-abc",
		}}

		result, err := c.EndSessionURL("id-token-123", "https://app.localssl.dev/library/login", "state-xyz")
		require.NoError(t, err)

		u, err := url.Parse(result)
		require.NoError(t, err)
		assert.Equal(t, "https://provider.example/logout", u.Scheme+"://"+u.Host+u.Path)

		q := u.Query()
		assert.Equal(t, "id-token-123", q.Get("id_token_hint"))
		assert.Equal(t, "client-abc", q.Get("client_id"))
		assert.Equal(t, "https://app.localssl.dev/library/login", q.Get("post_logout_redirect_uri"))
		assert.Equal(t, "state-xyz", q.Get("state"))
	})
	t.Run("NoEndSessionEndpoint", func(t *testing.T) {
		c := &Client{RelyingParty: fakeRelyingParty{endSession: "", clientID: "client-abc"}}

		result, err := c.EndSessionURL("id-token-123", "https://app.localssl.dev/library/login", "")
		require.NoError(t, err)
		assert.Equal(t, "", result)
	})
	t.Run("PreservesExistingQuery", func(t *testing.T) {
		c := &Client{RelyingParty: fakeRelyingParty{
			endSession: "https://provider.example/logout?realm=master",
			clientID:   "client-abc",
		}}

		result, err := c.EndSessionURL("id-token-123", "", "")
		require.NoError(t, err)

		u, err := url.Parse(result)
		require.NoError(t, err)
		q := u.Query()
		assert.Equal(t, "master", q.Get("realm"))
		assert.Equal(t, "id-token-123", q.Get("id_token_hint"))
		assert.Equal(t, "", q.Get("post_logout_redirect_uri"))
		assert.Equal(t, "", q.Get("state"))
	})
	t.Run("InvalidEndpoint", func(t *testing.T) {
		c := &Client{RelyingParty: fakeRelyingParty{endSession: "://not a url", clientID: "client-abc"}}

		_, err := c.EndSessionURL("id-token-123", "", "")
		assert.Error(t, err)
	})
}

package entity

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/pkg/authn"
)

func TestClient_IsApp(t *testing.T) {
	t.Run("App", func(t *testing.T) {
		assert.True(t, (&Client{AuthMethod: authn.MethodPKCE.String()}).IsApp())
	})
	t.Run("Client", func(t *testing.T) {
		assert.False(t, (&Client{AuthMethod: authn.MethodOAuth2.String()}).IsApp())
	})
	t.Run("Nil", func(t *testing.T) {
		var m *Client
		assert.False(t, m.IsApp())
		assert.Nil(t, m.RedirectURIs())
	})
}

func TestClient_RedirectURIs(t *testing.T) {
	t.Run("App", func(t *testing.T) {
		m := NewClient().SetFormValues(form.Client{ClientName: "Claude", AuthMethod: authn.MethodPKCE.String(), RedirectURIs: []string{"https://claude.ai/cb"}})
		assert.Equal(t, []string{"https://claude.ai/cb"}, m.RedirectURIs())
	})
	t.Run("None", func(t *testing.T) {
		assert.Empty(t, NewClient().RedirectURIs())
	})
}

func TestClient_SetFormValues_App(t *testing.T) {
	t.Run("App", func(t *testing.T) {
		frm := form.Client{ClientName: "Claude", AuthMethod: authn.MethodPKCE.String(), AuthScope: "mcp",
			RedirectURIs: []string{"https://claude.ai/api/mcp/auth_callback"}}
		m := NewClient().SetFormValues(frm)
		assert.True(t, m.IsApp())
		assert.Equal(t, authn.ClientPublic, m.ClientType)
		assert.Equal(t, "client", m.ClientRole)
		assert.Equal(t, []string{"https://claude.ai/api/mcp/auth_callback"}, m.RedirectURIs())
		assert.NoError(t, m.Validate())
	})
	t.Run("KeepsCallbacks", func(t *testing.T) {
		m := NewClient().SetFormValues(form.Client{ClientName: "Claude", AuthMethod: authn.MethodPKCE.String(), RedirectURIs: []string{"https://claude.ai/cb"}})
		m.SetFormValues(form.Client{ClientName: "Renamed"})
		assert.Equal(t, []string{"https://claude.ai/cb"}, m.RedirectURIs())
	})
}

func TestClient_Validate_App(t *testing.T) {
	t.Run("NoCallback", func(t *testing.T) {
		m := NewClient().SetFormValues(form.Client{ClientName: "Claude", AuthMethod: authn.MethodPKCE.String()})
		assert.Error(t, m.Validate())
	})
	t.Run("DataTooLarge", func(t *testing.T) {
		m := NewClient().SetFormValues(form.Client{ClientName: "Claude", AuthMethod: authn.MethodPKCE.String(),
			RedirectURIs: []string{"https://claude.ai/" + strings.Repeat("a", ClientDataMaxBytes)}})
		assert.Error(t, m.Validate())
	})
	t.Run("AddClient", func(t *testing.T) {
		m, err := AddClient(form.Client{ClientName: "App Test", AuthMethod: authn.MethodPKCE.String(), AuthScope: "mcp",
			RedirectURIs: []string{"https://claude.ai/api/mcp/auth_callback"}})
		require.NoError(t, err)

		found := FindClientByUID(m.ClientUID)
		require.NotNil(t, found)
		assert.True(t, found.IsApp())
		assert.Equal(t, authn.ClientPublic, found.ClientType)
		assert.Equal(t, []string{"https://claude.ai/api/mcp/auth_callback"}, found.RedirectURIs())
		assert.NoError(t, found.Delete())
	})
}

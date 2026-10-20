package api

import (
	"testing"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/pkg/authn"
	"github.com/photoprism/photoprism/pkg/rnd"
)

func TestMcpGrantSessions(t *testing.T) {
	server := sdkmcp.NewServer(&sdkmcp.Implementation{Name: "test", Version: "1.0"}, nil)

	t.Run("AddPrunesUnknownSessions", func(t *testing.T) {
		m := newMcpGrantSessions(server)
		m.Add("session-1", "gs5gfen1bgxz7s9i", nil)
		assert.Equal(t, 1, m.Len())

		// Sessions the server no longer has are forgotten when another one is added.
		m.Add("session-2", "gs5gfen1bgxz7s9i", nil)
		assert.Equal(t, 1, m.Len())
	})
	t.Run("Close", func(t *testing.T) {
		m := newMcpGrantSessions(server)
		m.Add("session-1", "gs5gfen1bgxz7s9i", nil)
		m.Close("gs5gfen1bgxz7s9j")
		assert.Equal(t, 1, m.Len())
		m.Close("gs5gfen1bgxz7s9i")
		assert.Equal(t, 0, m.Len())
	})
	t.Run("GrantGone", func(t *testing.T) {
		m := newMcpGrantSessions(server)
		checked := 0
		m.Add("session-1", "gs5gfen1bgxz7s9i", func(string) bool { checked++; return false })
		assert.Equal(t, 1, checked)
		assert.Equal(t, 0, m.Len())

		// A known session is not checked again.
		m.Add("session-2", "gs5gfen1bgxz7s9i", func(string) bool { return true })
		m.Add("session-2", "gs5gfen1bgxz7s9i", func(string) bool { checked++; return false })
		assert.Equal(t, 1, checked)
		assert.Equal(t, 1, m.Len())
	})
	t.Run("RevokedWhileRecording", func(t *testing.T) {
		m := newMcpGrantSessions(server)

		// The grant is revoked after the session was recorded and before the check.
		m.Add("session-1", "gs5gfen1bgxz7s9i", func(uid string) bool { m.Close(uid); return true })
		assert.Equal(t, 0, m.Len())
	})
	t.Run("Empty", func(t *testing.T) {
		m := newMcpGrantSessions(server)
		m.Add("", "gs5gfen1bgxz7s9i", nil)
		m.Add("session-1", "", nil)
		assert.Equal(t, 0, m.Len())
		var empty *mcpGrantSessions
		empty.Add("session-1", "gs5gfen1bgxz7s9i", nil)
		empty.Close("gs5gfen1bgxz7s9i")
		assert.Equal(t, 0, empty.Len())
	})
}

func TestMcpGrantActive(t *testing.T) {
	alice := entity.UserFixtures.Pointer("alice")

	newGrant := func(t *testing.T, clientID string) *entity.Grant {
		g := &entity.Grant{UserUID: alice.UserUID, ClientID: clientID, Scope: "mcp", RefreshHash: rnd.Sha256([]byte(rnd.AuthToken()))}
		require.NoError(t, g.Create())
		t.Cleanup(func() { _ = g.Delete() })
		return g
	}

	t.Run("Missing", func(t *testing.T) {
		assert.False(t, mcpGrantActive("gs5gfen1bgxz7s9i"))
		assert.False(t, mcpGrantActive(""))
	})
	t.Run("RegisteredClient", func(t *testing.T) {
		client, err := entity.AddClient(form.Client{ClientName: "Grant Active App", AuthMethod: authn.MethodPKCE.String(), AuthScope: "mcp",
			RedirectURIs: []string{"https://claude.ai/cb"}})
		require.NoError(t, err)
		t.Cleanup(func() { _ = client.Delete() })

		g := newGrant(t, client.ClientUID)
		assert.True(t, mcpGrantActive(g.GrantUID))

		require.NoError(t, client.Updates(entity.Values{"auth_enabled": false}))
		assert.False(t, mcpGrantActive(g.GrantUID))
	})
	t.Run("MetadataClient", func(t *testing.T) {
		g := newGrant(t, "https://app.example.com/oauth/client.json")
		assert.True(t, mcpGrantActive(g.GrantUID))
	})
}

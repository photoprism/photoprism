package api

import (
	"testing"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
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

func TestMcpGrantExists(t *testing.T) {
	assert.False(t, mcpGrantExists("gs5gfen1bgxz7s9i"))
	assert.False(t, mcpGrantExists(""))
}

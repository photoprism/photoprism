package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/entity"
)

func TestUsersCommands_MCP(t *testing.T) {
	const name = "cli-mcp-user"

	t.Cleanup(func() {
		if m := entity.FindUserByName(name); m != nil {
			_ = entity.UnscopedDb().Delete(&entity.User{}, "user_uid = ?", m.UserUID).Error
		}
	})

	t.Run("Flags", func(t *testing.T) {
		for _, cmd := range []string{"add", "mod"} {
			output, err := RunWithTestContext(UsersCommands, []string{"users", cmd, "--help"})
			require.NoError(t, err)
			assert.Contains(t, output, "--mcp", cmd)
		}
	})
	t.Run("AddDefault", func(t *testing.T) {
		_, err := RunWithTestContext(UsersAddCommand, []string{"add", "--password=test1234", "--role=admin", name})
		require.NoError(t, err)
		m := entity.FindUserByName(name)
		require.NotNil(t, m)
		assert.Equal(t, 0, m.MCP)
		assert.True(t, m.CanUseMCP())
	})
	t.Run("ModFalse", func(t *testing.T) {
		_, err := RunWithTestContext(UsersModCommand, []string{"mod", "--mcp=false", name})
		require.NoError(t, err)
		m := entity.FindUserByName(name)
		require.NotNil(t, m)
		assert.Equal(t, -1, m.MCP)
		assert.False(t, m.CanUseMCP())
	})
	t.Run("ModUnset", func(t *testing.T) {
		_, err := RunWithTestContext(UsersModCommand, []string{"mod", "--name=CLI MCP", name})
		require.NoError(t, err)
		assert.Equal(t, -1, entity.FindUserByName(name).MCP)
	})
	t.Run("List", func(t *testing.T) {
		output, err := RunWithTestContext(UsersListCommand, []string{"ls", "--md", name})
		require.NoError(t, err)
		assert.Contains(t, output, "| Web Login | WebDAV   | MCP      |")
		assert.Contains(t, output, "| Enabled   | Disabled | Disabled |")
	})
	t.Run("ModTrue", func(t *testing.T) {
		_, err := RunWithTestContext(UsersModCommand, []string{"mod", "--mcp", name})
		require.NoError(t, err)
		assert.Equal(t, 1, entity.FindUserByName(name).MCP)
	})
}

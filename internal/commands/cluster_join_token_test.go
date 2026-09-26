package commands

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/internal/service/cluster"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/rnd"
)

func firstLine(s string) string {
	trimmed := strings.TrimSpace(s)
	if trimmed == "" {
		return ""
	}
	if idx := strings.IndexRune(trimmed, '\n'); idx >= 0 {
		return trimmed[:idx]
	}
	return trimmed
}

func TestClusterJoinToken_PrintOnly(t *testing.T) {
	out, err := RunWithTestContext(ClusterJoinTokenCommand, []string{"join-token"})
	assert.NoError(t, err)

	token := firstLine(out)
	assert.True(t, rnd.IsJoinToken(token, false))
}

func TestClusterJoinToken_Save(t *testing.T) {
	conf := get.Config()
	prevEdition := conf.Options().Edition
	prevRole := conf.Options().NodeRole
	conf.Options().Edition = config.Portal
	conf.Options().NodeRole = cluster.RolePortal
	t.Cleanup(func() {
		conf.Options().Edition = prevEdition
		conf.Options().NodeRole = prevRole
	})
	targetFile := conf.PortalJoinTokenFile()
	_ = os.Remove(targetFile)

	out, err := RunWithTestContext(ClusterJoinTokenCommand, []string{"join-token", "--save", "--yes"})
	assert.NoError(t, err)

	token := firstLine(out)
	assert.True(t, rnd.IsJoinToken(token, false))

	data, readErr := os.ReadFile(conf.PortalJoinTokenFile())
	assert.NoError(t, readErr)
	assert.Equal(t, token, strings.TrimSpace(string(data)))
}

// TestClusterJoinToken_Replace verifies that replacing a saved token asks for confirmation, and that the
// command exits with a usage error when it cannot ask.
func TestClusterJoinToken_Replace(t *testing.T) {
	t.Setenv("PHOTOPRISM_CLI", "")

	conf := get.Config()
	prevEdition := conf.Options().Edition
	prevRole := conf.Options().NodeRole
	conf.Options().Edition = config.Portal
	conf.Options().NodeRole = cluster.RolePortal
	t.Cleanup(func() {
		conf.Options().Edition = prevEdition
		conf.Options().NodeRole = prevRole
	})

	targetFile := conf.PortalJoinTokenFile()
	existing := rnd.JoinToken()
	require.NoError(t, os.MkdirAll(filepath.Dir(targetFile), fs.ModeDir))
	require.NoError(t, os.WriteFile(targetFile, []byte(existing), fs.ModeSecretFile))
	t.Cleanup(func() { _ = os.Remove(targetFile) })

	saved := func() string {
		data, err := os.ReadFile(targetFile)
		require.NoError(t, err)
		return strings.TrimSpace(string(data))
	}

	t.Run("NoFile", func(t *testing.T) {
		// Without a saved token, there is nothing to confirm.
		require.NoError(t, os.Remove(targetFile))

		out, err := RunWithTestContext(ClusterJoinTokenCommand, []string{"join-token", "--save"})

		assert.NoError(t, err)
		assert.Equal(t, firstLine(out), saved())
		require.NoError(t, os.WriteFile(targetFile, []byte(existing), fs.ModeSecretFile))
	})
	t.Run("NoTerminal", func(t *testing.T) {
		_, err := RunWithTestContext(ClusterJoinTokenCommand, []string{"join-token", "--save"})

		var exit cli.ExitCoder
		require.ErrorAs(t, err, &exit)
		assert.Equal(t, 2, exit.ExitCode())
		assert.Contains(t, err.Error(), "--yes")
		assert.Equal(t, existing, saved())
	})
	t.Run("AnsweredNo", func(t *testing.T) {
		pipeResetAnswers(t, "n\n")

		_, err := RunWithTestContext(ClusterJoinTokenCommand, []string{"join-token", "--save"})

		assert.NoError(t, err)
		assert.Equal(t, existing, saved())
	})
	t.Run("AnsweredYes", func(t *testing.T) {
		pipeResetAnswers(t, "y\n")

		out, err := RunWithTestContext(ClusterJoinTokenCommand, []string{"join-token", "--save"})

		assert.NoError(t, err)
		assert.Equal(t, firstLine(out), saved())
		assert.NotEqual(t, existing, saved())
	})
}

package commands

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/internal/service/cluster"
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
	prevRole := conf.Options().NodeRole
	conf.Options().NodeRole = cluster.RolePortal
	t.Cleanup(func() {
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

package commands

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/internal/service/cluster"
	reg "github.com/photoprism/photoprism/internal/service/cluster/registry"
	"github.com/photoprism/photoprism/pkg/rnd"
)

func TestAuthJWTCommands(t *testing.T) {
	conf := get.Config()

	origEdition := conf.Options().Edition
	origRole := conf.Options().NodeRole
	origUUID := conf.Options().ClusterUUID
	origPortal := conf.Options().PortalUrl
	origJWKS := conf.JWKSUrl()

	conf.Options().Edition = config.Portal
	conf.Options().NodeRole = string(cluster.RolePortal)
	conf.Options().ClusterUUID = "11111111-1111-4111-8111-111111111111"
	conf.Options().PortalUrl = "https://portal.test"
	conf.SetJWKSUrl("https://portal.test/.well-known/jwks.json")

	get.SetConfig(conf)
	conf.RegisterDb()

	require.True(t, conf.Portal())

	manager := get.JWTManager()
	require.NotNil(t, manager)
	_, err := manager.EnsureActiveKey()
	require.NoError(t, err)

	registry, err := reg.NewClientRegistryWithConfig(conf)
	require.NoError(t, err)

	nodeUUID := rnd.UUID()
	node := &reg.Node{}
	node.UUID = nodeUUID
	node.Name = "pp-node-01"
	node.Role = string(cluster.RoleInstance)
	require.NoError(t, registry.Put(node))
	t.Cleanup(func() {
		conf.Options().Edition = origEdition
		conf.Options().NodeRole = origRole
		conf.Options().ClusterUUID = origUUID
		conf.Options().PortalUrl = origPortal
		conf.SetJWKSUrl(origJWKS)
		get.SetConfig(conf)
		conf.RegisterDb()
	})

	output, err := RunWithTestContext(AuthJWTIssueCommand, []string{"issue", "--node", nodeUUID})
	require.NoError(t, err)
	assert.Contains(t, output, "Issued JWT")

	jsonOut, err := RunWithTestContext(AuthJWTIssueCommand, []string{"issue", "--node", nodeUUID, "--json"})
	require.NoError(t, err)

	var payload struct {
		Token string `json:"token"`
	}
	require.NoError(t, json.Unmarshal([]byte(jsonOut), &payload))
	require.NotEmpty(t, payload.Token)

	inspectOut, err := RunWithTestContext(AuthJWTInspectCommand, []string{"inspect", "--json", payload.Token})
	require.NoError(t, err)
	assert.Contains(t, inspectOut, "\"verified\": true")

	inspectStrict, err := RunWithTestContext(AuthJWTInspectCommand, []string{"inspect", "--json", "--expect-audience", "node:" + nodeUUID, "--require-scope", "cluster", payload.Token})
	require.NoError(t, err)
	assert.Contains(t, inspectStrict, "\"verified\": true")

	keysOut, err := RunWithTestContext(AuthJWTKeysListCommand, []string{"ls", "--json"})
	require.NoError(t, err)
	assert.Contains(t, keysOut, "\"keys\"")

	statusOut, err := RunWithTestContext(AuthJWTStatusCommand, []string{"status"})
	require.NoError(t, err)
	assert.Contains(t, statusOut, "JWKS URL")
	assert.Contains(t, statusOut, "Cached Keys")

	// invalid scope should fail
	_, err = RunWithTestContext(AuthJWTIssueCommand, []string{"issue", "--node", nodeUUID, "--scope", "unknown"})
	require.Error(t, err)
}

package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/internal/service/cluster"
	reg "github.com/photoprism/photoprism/internal/service/cluster/registry"
)

func TestClusterNodesMod_LegacyAliasAppToInstance(t *testing.T) {
	c := get.Config()
	prevEdition := c.Options().Edition
	prevRole := c.Options().NodeRole
	c.Options().Edition = config.Portal
	c.Options().NodeRole = cluster.RolePortal
	t.Cleanup(func() {
		c.Options().Edition = prevEdition
		c.Options().NodeRole = prevRole
	})

	r, err := reg.NewClientRegistryWithConfig(c)
	assert.NoError(t, err)

	n := &reg.Node{Node: cluster.Node{Name: "pp-mod-alias", Role: cluster.RoleService}}
	assert.NoError(t, r.Put(n))

	_, err = RunWithTestContext(ClusterNodesModCommand, []string{"mod", "--role=app", "-y", "pp-mod-alias"})
	assert.NoError(t, err)

	updated, err := r.FindByName("pp-mod-alias")
	assert.NoError(t, err)
	if assert.NotNil(t, updated) {
		assert.Equal(t, cluster.RoleInstance, updated.Role)
	}
}

func TestClusterNodesMod_InvalidRole(t *testing.T) {
	c := get.Config()
	prevEdition := c.Options().Edition
	prevRole := c.Options().NodeRole
	c.Options().Edition = config.Portal
	c.Options().NodeRole = cluster.RolePortal
	t.Cleanup(func() {
		c.Options().Edition = prevEdition
		c.Options().NodeRole = prevRole
	})

	r, err := reg.NewClientRegistryWithConfig(c)
	assert.NoError(t, err)

	n := &reg.Node{Node: cluster.Node{Name: "pp-mod-invalid", Role: cluster.RoleService}}
	assert.NoError(t, r.Put(n))

	_, err = RunWithTestContext(ClusterNodesModCommand, []string{"mod", "--role=invalid", "-y", "pp-mod-invalid"})
	assert.Error(t, err)

	ec, ok := err.(cli.ExitCoder)
	if !ok {
		t.Fatalf("expected ExitCoder, got %T", err)
	}
	assert.Equal(t, 2, ec.ExitCode())
}

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

func TestClusterNodesMod_DisplayNameOverride(t *testing.T) {
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

	// Instance reports a display name on registration.
	assert.NoError(t, r.Put(&reg.Node{Node: cluster.Node{Name: "pp-mod-dn", Role: cluster.RoleInstance, DisplayName: "Reported"}}))

	// Admin override pins a new value.
	_, err = RunWithTestContext(ClusterNodesModCommand, []string{"mod", "--display-name=Pinned Name", "-y", "pp-mod-dn"})
	assert.NoError(t, err)
	updated, err := r.FindByName("pp-mod-dn")
	assert.NoError(t, err)
	if !assert.NotNil(t, updated) {
		return
	}
	assert.Equal(t, "Pinned Name", updated.DisplayName)

	// A subsequent instance registration must not overwrite the pinned override.
	assert.NoError(t, r.Put(&reg.Node{Node: cluster.Node{ClientID: updated.ClientID, Name: updated.Name, DisplayName: "Reported Again"}}))
	updated, _ = r.FindByName("pp-mod-dn")
	assert.Equal(t, "Pinned Name", updated.DisplayName)

	// Clearing via an empty override un-pins and falls back to instance-reported.
	_, err = RunWithTestContext(ClusterNodesModCommand, []string{"mod", "--display-name=", "-y", "pp-mod-dn"})
	assert.NoError(t, err)
	updated, _ = r.FindByName("pp-mod-dn")
	assert.Equal(t, "", updated.DisplayName)
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

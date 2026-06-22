package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"

	"github.com/photoprism/photoprism/internal/auth/acl"
)

// TestGroupRoleFlagsUsage pins the hidden --oidc-group-role and
// --cluster-allow-group-roles flags to the shared acl.ClusterInstanceRoles list,
// so the role help cannot drift. These flags are Hidden and therefore never
// render via "--help" or "show commands", so this is the only coverage of their
// usage text.
func TestGroupRoleFlagsUsage(t *testing.T) {
	roles := acl.ClusterInstanceRolesCliUsageString()
	assert.Equal(t, "admin, manager, user, contributor, viewer, or guest", roles)

	found := map[string]bool{"oidc-group-role": false, "cluster-allow-group-roles": false}
	for _, f := range Flags {
		ssf, ok := f.Flag.(*cli.StringSliceFlag)
		if !ok {
			continue
		}
		if _, tracked := found[ssf.Name]; !tracked {
			continue
		}
		found[ssf.Name] = true
		assert.True(t, ssf.Hidden, "%s must stay hidden", ssf.Name)
		assert.Contains(t, ssf.Usage, roles, "%s usage must list the cluster instance roles", ssf.Name)
	}
	for name, ok := range found {
		assert.Truef(t, ok, "flag --%s not found in config.Flags", name)
	}
}

package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"
)

func TestClientRoleFlagUsage_ListsAssignableRoles(t *testing.T) {
	assertRoleFlagUsage := func(t *testing.T, cmd *cli.Command, owner string) {
		var roleFlag *cli.StringFlag
		for _, f := range cmd.Flags {
			if rf, ok := f.(*cli.StringFlag); ok && rf.Name == "role" {
				roleFlag = rf
				break
			}
		}
		if roleFlag == nil {
			t.Fatalf("role flag not found on %s", owner)
		}
		// Lists the assignable client roles; the "none" alias is excluded from CLI help.
		for _, role := range []string{"admin", "client", "instance", "portal", "service"} {
			assert.Contains(t, roleFlag.Usage, role)
		}
		assert.NotContains(t, roleFlag.Usage, "none")
	}
	t.Run("AddCommandRoleFlag", func(t *testing.T) {
		assertRoleFlagUsage(t, ClientsAddCommand, "ClientsAddCommand")
	})
	t.Run("ModCommandRoleFlag", func(t *testing.T) {
		assertRoleFlagUsage(t, ClientsModCommand, "ClientsModCommand")
	})
}

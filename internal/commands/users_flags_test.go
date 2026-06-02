package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"
)

func TestUserRoleFlagUsage_ListsAssignableRoles(t *testing.T) {
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
		// Lists the assignable user roles; the "none" alias and the "visitor"
		// share-link role are excluded from CLI help.
		assert.Contains(t, roleFlag.Usage, "admin")
		assert.Contains(t, roleFlag.Usage, "guest")
		assert.NotContains(t, roleFlag.Usage, "none")
		assert.NotContains(t, roleFlag.Usage, "visitor")
	}
	t.Run("AddCommandUserRoleFlag", func(t *testing.T) {
		assertRoleFlagUsage(t, UsersAddCommand, "UsersAddCommand")
	})
	t.Run("ModCommandUserRoleFlag", func(t *testing.T) {
		assertRoleFlagUsage(t, UsersModCommand, "UsersModCommand")
	})
}

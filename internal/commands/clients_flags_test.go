package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"
)

func TestClientRoleFlagUsage_ExcludesNoneAlias(t *testing.T) {
	t.Run("AddCommandRoleFlagExcludesNone", func(t *testing.T) {
		var roleFlag *cli.StringFlag
		for _, f := range ClientsAddCommand.Flags {
			if rf, ok := f.(*cli.StringFlag); ok && rf.Name == "role" {
				roleFlag = rf
				break
			}
		}
		if roleFlag == nil {
			t.Fatal("role flag not found on ClientsAddCommand")
		}
		// Real client roles are listed; the "none" alias is excluded.
		assert.Contains(t, roleFlag.Usage, "client")
		assert.Contains(t, roleFlag.Usage, "service")
		assert.NotContains(t, roleFlag.Usage, "none")
	})
	t.Run("ModCommandRoleFlagExcludesNone", func(t *testing.T) {
		var roleFlag *cli.StringFlag
		for _, f := range ClientsModCommand.Flags {
			if rf, ok := f.(*cli.StringFlag); ok && rf.Name == "role" {
				roleFlag = rf
				break
			}
		}
		if roleFlag == nil {
			t.Fatal("role flag not found on ClientsModCommand")
		}
		assert.Contains(t, roleFlag.Usage, "client")
		assert.Contains(t, roleFlag.Usage, "service")
		assert.NotContains(t, roleFlag.Usage, "none")
	})
}

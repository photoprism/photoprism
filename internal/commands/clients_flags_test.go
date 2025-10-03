package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"
)

func TestClientRoleFlagUsage_IncludesNoneAlias(t *testing.T) {
	t.Run("AddCommandRoleFlagIncludesNone", func(t *testing.T) {
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
		assert.Contains(t, roleFlag.Usage, "none")
	})
	t.Run("ModCommandRoleFlagIncludesNone", func(t *testing.T) {
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
		assert.Contains(t, roleFlag.Usage, "none")
	})
}

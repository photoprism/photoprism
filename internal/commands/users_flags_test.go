package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"
)

func TestUserRoleFlagUsage_IncludesNoneAlias(t *testing.T) {
	t.Run("AddCommand user role flag includes none", func(t *testing.T) {
		var roleFlag *cli.StringFlag
		for _, f := range UsersAddCommand.Flags {
			if rf, ok := f.(*cli.StringFlag); ok && rf.Name == "role" {
				roleFlag = rf
				break
			}
		}
		if roleFlag == nil {
			t.Fatal("role flag not found on UsersAddCommand")
		}
		assert.Contains(t, roleFlag.Usage, "none")
	})

	t.Run("ModCommand user role flag includes none", func(t *testing.T) {
		var roleFlag *cli.StringFlag
		for _, f := range UsersModCommand.Flags {
			if rf, ok := f.(*cli.StringFlag); ok && rf.Name == "role" {
				roleFlag = rf
				break
			}
		}
		if roleFlag == nil {
			t.Fatal("role flag not found on UsersModCommand")
		}
		assert.Contains(t, roleFlag.Usage, "none")
	})
}

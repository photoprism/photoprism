package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"
)

func TestUserRoleFlagUsage_ExcludesNoneAndVisitor(t *testing.T) {
	t.Run("AddCommandUserRoleFlagExcludesNoneAndVisitor", func(t *testing.T) {
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
		// Offered roles are listed; the "none" alias and "visitor" are excluded.
		assert.Contains(t, roleFlag.Usage, "admin")
		assert.Contains(t, roleFlag.Usage, "guest")
		assert.NotContains(t, roleFlag.Usage, "none")
		assert.NotContains(t, roleFlag.Usage, "visitor")
	})
	t.Run("ModCommandUserRoleFlagExcludesNoneAndVisitor", func(t *testing.T) {
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
		assert.Contains(t, roleFlag.Usage, "admin")
		assert.Contains(t, roleFlag.Usage, "guest")
		assert.NotContains(t, roleFlag.Usage, "none")
		assert.NotContains(t, roleFlag.Usage, "visitor")
	})
}

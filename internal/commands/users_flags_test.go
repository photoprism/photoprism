package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"

	"github.com/photoprism/photoprism/internal/auth/acl"
)

func TestUserRoleUsageFor(t *testing.T) {
	// Lists exactly the assignable roles in the given map — including cluster_admin
	// when the map registers it (the Portal case) — and drops the visitor and
	// uploader-alias entries via RoleStrings.Strings.
	m := acl.RoleStrings{
		string(acl.RoleAdmin):        acl.RoleAdmin,
		string(acl.RoleClusterAdmin): acl.RoleClusterAdmin,
		"uploader":                   acl.RoleContributor,
		string(acl.RoleVisitor):      acl.RoleVisitor,
		string(acl.RoleGuest):        acl.RoleGuest,
	}
	u := UserRoleUsageFor(m)
	assert.Contains(t, u, "user account `ROLE`")
	assert.Contains(t, u, "cluster_admin")
	assert.Contains(t, u, "guest")
	assert.NotContains(t, u, "uploader")
	assert.NotContains(t, u, "visitor")
}

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

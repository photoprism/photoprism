package acl

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPermissions_String(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		perms := Permissions{}
		assert.Equal(t, "", perms.String())
	})
	t.Run("FullAccess", func(t *testing.T) {
		perms := Permissions{FullAccess}
		assert.Equal(t, "full access", perms.String())
	})
	t.Run("ManageUploadAll", func(t *testing.T) {
		perms := Permissions{ActionManage, ActionUpload, AccessAll}
		assert.Equal(t, "manage, upload, access all", perms.String())
	})
}

func TestPermissions_First(t *testing.T) {
	t.Run("Default", func(t *testing.T) {
		perms := Permissions{}
		assert.Equal(t, ActionUse.String(), perms.First())
	})
	t.Run("Explicit", func(t *testing.T) {
		perms := Permissions{ActionManage, ActionUpload}
		assert.Equal(t, ActionManage.String(), perms.First())
	})
}

func TestPermissions_Contains(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		perms := Permissions{}
		assert.False(t, perms.Contains(ActionUse))
	})
	t.Run("ExactMatch", func(t *testing.T) {
		perms := Permissions{ActionManage, ActionUpload}
		assert.True(t, perms.Contains(ActionUpload))
		assert.False(t, perms.Contains(ActionDelete))
	})
	t.Run("WildcardInSet", func(t *testing.T) {
		perms := Permissions{Any}
		assert.True(t, perms.Contains(ActionDelete))
	})
	t.Run("WildcardRequested", func(t *testing.T) {
		perms := Permissions{ActionManage}
		assert.True(t, perms.Contains(Any))
	})
}

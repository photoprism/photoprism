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

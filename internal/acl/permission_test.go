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

func TestPermission_LogId(t *testing.T) {
	t.Run("FullAccess", func(t *testing.T) {
		assert.Equal(t, "full access", FullAccess.LogId())
	})
	t.Run("ActionUpload", func(t *testing.T) {
		assert.Equal(t, "upload", ActionUpload.LogId())
	})
}

func TestPermission_Equal(t *testing.T) {
	t.Run("True", func(t *testing.T) {
		assert.True(t, FullAccess.Equal("full access"))
	})
	t.Run("False", func(t *testing.T) {
		assert.False(t, ActionUpload.Equal("full access"))
	})
}

func TestPermission_NotEqual(t *testing.T) {
	t.Run("False", func(t *testing.T) {
		assert.False(t, FullAccess.NotEqual("full access"))
	})
	t.Run("True", func(t *testing.T) {
		assert.True(t, ActionUpload.NotEqual("full access"))
	})
}

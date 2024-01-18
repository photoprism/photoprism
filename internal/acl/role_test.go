package acl

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRole_String(t *testing.T) {
	t.Run("Admin", func(t *testing.T) {
		assert.Equal(t, "admin", RoleAdmin.String())
	})
	t.Run("Client", func(t *testing.T) {
		assert.Equal(t, "client", RoleClient.String())
	})
	t.Run("Visitor", func(t *testing.T) {
		assert.Equal(t, "visitor", RoleVisitor.String())
	})
}

func TestRole_Pretty(t *testing.T) {
	t.Run("Admin", func(t *testing.T) {
		assert.Equal(t, "Admin", RoleAdmin.Pretty())
	})
	t.Run("None", func(t *testing.T) {
		assert.Equal(t, "None", RoleNone.Pretty())
	})
	t.Run("Visitor", func(t *testing.T) {
		assert.Equal(t, "Visitor", RoleVisitor.Pretty())
	})
}

func TestRole_LogId(t *testing.T) {
	t.Run("Admin", func(t *testing.T) {
		assert.Equal(t, "role admin", RoleAdmin.LogId())
	})
	t.Run("Client", func(t *testing.T) {
		assert.Equal(t, "role client", RoleClient.LogId())
	})
	t.Run("Visitor", func(t *testing.T) {
		assert.Equal(t, "role visitor", RoleVisitor.LogId())
	})
}

func TestRole_Equal(t *testing.T) {
	t.Run("True", func(t *testing.T) {
		assert.True(t, RoleAdmin.Equal("admin"))
	})
	t.Run("False", func(t *testing.T) {
		assert.False(t, RoleClient.Equal("admin"))
	})
}

func TestRole_NotEqual(t *testing.T) {
	t.Run("True", func(t *testing.T) {
		assert.True(t, RoleClient.NotEqual("admin"))
	})
	t.Run("False", func(t *testing.T) {
		assert.False(t, RoleAdmin.NotEqual("admin"))
	})
}

func TestRole_Valid(t *testing.T) {
	t.Run("True", func(t *testing.T) {
		assert.True(t, RoleAdmin.Valid("admin"))
	})
	t.Run("False", func(t *testing.T) {
		assert.False(t, RoleClient.Valid("client"))
	})
}

func TestRole_Invalid(t *testing.T) {
	t.Run("False", func(t *testing.T) {
		assert.False(t, RoleAdmin.Invalid("admin"))
	})
	t.Run("True", func(t *testing.T) {
		assert.True(t, RoleClient.Invalid("client"))
	})
}

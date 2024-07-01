package acl

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestACL_Grants(t *testing.T) {
	t.Run("RoleAdmin", func(t *testing.T) {
		result := Rules.Grants(RoleAdmin)
		assert.True(t, result[ResourcePhotos][ActionManage])
		assert.True(t, result[ResourceConfig][ActionManage])
	})
	t.Run("RoleGuest", func(t *testing.T) {
		result := Rules.Grants(RoleGuest)
		assert.False(t, result[ResourcePhotos][ActionUpdate])
		assert.False(t, result[ResourceConfig][ActionManage])
	})
	t.Run("RoleVisitor", func(t *testing.T) {
		result := Rules.Grants(RoleVisitor)
		assert.False(t, result[ResourcePhotos][ActionUpdate])
		assert.False(t, result[ResourceConfig][ActionManage])
	})
}

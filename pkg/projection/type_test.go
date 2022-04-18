package projection

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestType_Equal(t *testing.T) {
	t.Run("UnknownUnknown", func(t *testing.T) {
		assert.True(t, Unknown.Equal(""))
	})
	t.Run("CubestripCubestrip", func(t *testing.T) {
		assert.True(t, Cubestrip.Equal(Cubestrip.String()))
	})
	t.Run("CubestripCylindrical", func(t *testing.T) {
		assert.False(t, Cubestrip.Equal(Cylindrical.String()))
	})
	t.Run("CylindricalUnknown", func(t *testing.T) {
		assert.False(t, Cylindrical.Equal(Unknown.String()))
	})
}

func TestType_NotEqual(t *testing.T) {
	t.Run("UnknownUnknown", func(t *testing.T) {
		assert.False(t, Unknown.NotEqual(""))
	})
	t.Run("CubestripCubestrip", func(t *testing.T) {
		assert.False(t, Cubestrip.NotEqual(Cubestrip.String()))
	})
	t.Run("CubestripCylindrical", func(t *testing.T) {
		assert.True(t, Cubestrip.NotEqual(Cylindrical.String()))
	})
	t.Run("CylindricalUnknown", func(t *testing.T) {
		assert.True(t, Cylindrical.NotEqual(Unknown.String()))
	})
}

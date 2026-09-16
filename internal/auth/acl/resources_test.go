package acl

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResources_String(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		assert.Equal(t, "", Resources{}.String())
	})
	t.Run("One", func(t *testing.T) {
		assert.Equal(t, "photos", Resources{ResourcePhotos}.String())
	})
	t.Run("Many", func(t *testing.T) {
		assert.Equal(t, "files, photos", Resources{ResourceFiles, ResourcePhotos}.String())
	})
	t.Run("Unnamed", func(t *testing.T) {
		assert.Equal(t, "default", Resources{Resource("")}.String())
	})
}

func TestResources_First(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		assert.Equal(t, ResourceDefault, Resources{}.First())
		assert.Equal(t, ResourceDefault, Resources(nil).First())
	})
	t.Run("Many", func(t *testing.T) {
		assert.Equal(t, ResourceFiles, Resources{ResourceFiles, ResourcePhotos}.First())
	})
}

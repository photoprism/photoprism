// +build osm

package osm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOSM_Name(t *testing.T) {
	t.Run("Nice Name", func(t *testing.T) {
		l := &Location{LocCategory: "natural", LocName: "Nice Name", LocType: "hill", LocDisplayName: "display name"}
		assert.Equal(t, "Nice Name", l.Name())
	})

	t.Run("Water", func(t *testing.T) {
		l := &Location{LocCategory: "", LocName: "", LocType: "water", LocDisplayName: "display name"}
		assert.Equal(t, "Water", l.Name())
	})

	t.Run("Nice Name 2", func(t *testing.T) {
		l := &Location{LocCategory: "shop", LocName: "Nice Name 2", LocType: "", LocDisplayName: "display name"}
		assert.Equal(t, "Nice Name 2", l.Name())
	})

	t.Run("Cat", func(t *testing.T) {
		l := &Location{LocCategory: "xxx", LocName: "Cat,Dog", LocType: "", LocDisplayName: "display name"}
		assert.Equal(t, "Cat", l.Name())
	})

	t.Run("airport", func(t *testing.T) {
		l := &Location{LocCategory: "aeroway", LocName: "", LocType: "", LocDisplayName: "display name"}
		assert.Equal(t, "Airport", l.Name())
	})

	t.Run("Cow", func(t *testing.T) {
		l := &Location{LocCategory: "xxx", LocName: "Cow - Cat - Dog", LocType: "", LocDisplayName: "display name"}
		assert.Equal(t, "Cow", l.Name())
	})
}

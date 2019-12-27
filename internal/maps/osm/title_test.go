package osm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOSM_Name(t *testing.T) {
	t.Run("Nice Name", func(t *testing.T) {
		l := &Location{LocCategory: "natural", LocLat: "52.5208", LocLng: "13.40953", LocName: "Nice Name", LocType: "hill", LocDisplayName: "dipslay name"}
		assert.Equal(t, "Nice Name", l.Name())
	})

	t.Run("Water", func(t *testing.T) {
		l := &Location{LocCategory: "", LocLat: "52.5208", LocLng: "13.40953", LocName: "", LocType: "water", LocDisplayName: "dipslay name"}
		assert.Equal(t, "Water", l.Name())
	})

	t.Run("Nice Name 2", func(t *testing.T) {
		l := &Location{LocCategory: "shop", LocLat: "52.5208", LocLng: "13.40953", LocName: "Nice Name 2", LocType: "", LocDisplayName: "dipslay name"}
		assert.Equal(t, "Nice Name 2", l.Name())
	})

	t.Run("Cat", func(t *testing.T) {
		l := &Location{LocCategory: "xxx", LocLat: "52.5208", LocLng: "13.40953", LocName: "Cat,Dog", LocType: "", LocDisplayName: "dipslay name"}
		assert.Equal(t, "Cat", l.Name())
	})

	t.Run("airport", func(t *testing.T) {
		l := &Location{LocCategory: "aeroway", LocLat: "52.5208", LocLng: "13.40953", LocName: "", LocType: "", LocDisplayName: "dipslay name"}
		assert.Equal(t, "Airport", l.Name())
	})

	t.Run("Cow", func(t *testing.T) {
		l := &Location{LocCategory: "xxx", LocLat: "52.5208", LocLng: "13.40953", LocName: "Cow - Cat - Dog", LocType: "", LocDisplayName: "dipslay name"}
		assert.Equal(t, "Cow", l.Name())
	})
}

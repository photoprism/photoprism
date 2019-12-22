package osm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOSM_Title(t *testing.T) {
	t.Run("Nice Title", func(t *testing.T) {

		l := &Location{LocCategory: "natural", LocLat: "52.5208", LocLng: "13.40953", LocTitle: "Nice title", LocType: "hill", LocDisplayName: "dipslay name"}
		assert.Equal(t, "Nice Title", l.Title())
	})

	t.Run("Water", func(t *testing.T) {

		l := &Location{LocCategory: "", LocLat: "52.5208", LocLng: "13.40953", LocTitle: "", LocType: "water", LocDisplayName: "dipslay name"}
		assert.Equal(t, "Water", l.Title())
	})

	t.Run("Nice Title 2", func(t *testing.T) {

		l := &Location{LocCategory: "shop", LocLat: "52.5208", LocLng: "13.40953", LocTitle: "Nice title_2", LocType: "", LocDisplayName: "dipslay name"}
		assert.Equal(t, "Nice Title 2", l.Title())
	})

	t.Run("Cat", func(t *testing.T) {

		l := &Location{LocCategory: "xxx", LocLat: "52.5208", LocLng: "13.40953", LocTitle: "Cat,Dog", LocType: "", LocDisplayName: "dipslay name"}
		assert.Equal(t, "Cat", l.Title())
	})
	t.Run("airport", func(t *testing.T) {

		l := &Location{LocCategory: "aeroway", LocLat: "52.5208", LocLng: "13.40953", LocTitle: "", LocType: "", LocDisplayName: "dipslay name"}
		assert.Equal(t, "Airport", l.Title())
	})
	t.Run("Cow", func(t *testing.T) {

		l := &Location{LocCategory: "xxx", LocLat: "52.5208", LocLng: "13.40953", LocTitle: "Cow - Cat - Dog", LocType: "", LocDisplayName: "dipslay name"}
		assert.Equal(t, "Cow", l.Title())
	})
}

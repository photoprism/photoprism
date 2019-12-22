package osm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOSM_Category(t *testing.T) {
	t.Run("hill", func(t *testing.T) {

		l := &Location{LocCategory: "natural", LocLat: "52.5208", LocLng: "13.40953", LocTitle: "Nice title", LocType: "hill", LocDisplayName: "dipslay name"}
		assert.Equal(t, "hill", l.Category())
	})

	t.Run("water", func(t *testing.T) {

		l := &Location{LocCategory: "", LocLat: "52.5208", LocLng: "13.40953", LocTitle: "Nice title", LocType: "water", LocDisplayName: "dipslay name"}
		assert.Equal(t, "water", l.Category())
	})

	t.Run("shop", func(t *testing.T) {

		l := &Location{LocCategory: "shop", LocLat: "52.5208", LocLng: "13.40953", LocTitle: "Nice title", LocType: "", LocDisplayName: "dipslay name"}
		assert.Equal(t, "shop", l.Category())
	})

	t.Run("no label found", func(t *testing.T) {

		l := &Location{LocCategory: "xxx", LocLat: "52.5208", LocLng: "13.40953", LocTitle: "Nice title", LocType: "", LocDisplayName: "dipslay name"}
		assert.Equal(t, "", l.Category())
	})
}

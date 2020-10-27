// +build osm

package osm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOSM_Category(t *testing.T) {
	t.Run("hill", func(t *testing.T) {

		l := &Location{LocCategory: "natural", LocName: "Nice title", LocType: "hill", LocDisplayName: "display name"}
		assert.Equal(t, "hill", l.Category())
	})

	t.Run("water", func(t *testing.T) {

		l := &Location{LocCategory: "", LocName: "Nice title", LocType: "water", LocDisplayName: "display name"}
		assert.Equal(t, "water", l.Category())
	})

	t.Run("shop", func(t *testing.T) {

		l := &Location{LocCategory: "shop", LocName: "Nice title", LocType: "", LocDisplayName: "display name"}
		assert.Equal(t, "shop", l.Category())
	})

	t.Run("no label found", func(t *testing.T) {

		l := &Location{LocCategory: "xxx", LocName: "Nice title", LocType: "", LocDisplayName: "display name"}
		assert.Equal(t, "", l.Category())
	})
}

package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLinkMap_Get(t *testing.T) {
	ValidateFixtures(t)
	t.Run("GetExistingLink", func(t *testing.T) {
		r := LinkFixtures.Get("1jxf3jfn2k")
		assert.Equal(t, "ss62xpryd1ob7gtf", r.LinkUID)
		assert.EqualValues(t, 12, r.LinkViews)
		assert.IsType(t, Link{}, r)
	})
	t.Run("GetNotExistingLink", func(t *testing.T) {
		r := LinkFixtures.Get("XXX")
		assert.Equal(t, Link{}, r)
		assert.IsType(t, Link{}, r)
	})
}

func TestLinkMap_Pointer(t *testing.T) {
	ValidateFixtures(t)
	t.Run("GetExistingLinkPointer", func(t *testing.T) {
		r := LinkFixtures.Pointer("1jxf3jfn2k")
		assert.Equal(t, "ss62xpryd1ob7gtf", r.LinkUID)
		assert.EqualValues(t, 12, r.LinkViews)
		assert.IsType(t, &Link{}, r)
	})
	t.Run("GetNotExistingLinkPointer", func(t *testing.T) {
		r := LinkFixtures.Pointer("XXX")
		assert.Equal(t, &Link{}, r)
		assert.IsType(t, &Link{}, r)
	})
}

package entity

import (
	"testing"

	"github.com/photoprism/photoprism/pkg/rnd"
	"github.com/stretchr/testify/assert"
)

func TestNewLink(t *testing.T) {
	link := NewLink("st9lxuqxpogaaba1", true, false)
	assert.Equal(t, "st9lxuqxpogaaba1", link.ShareUID)
	assert.Equal(t, false, link.CanEdit)
	assert.Equal(t, true, link.CanComment)
	assert.Equal(t, 10, len(link.ShareToken))
	assert.Equal(t, 16, len(link.LinkUID))
}

func TestLink_Expired(t *testing.T) {
	const oneDay = 60 * 60 * 24

	link := NewLink("st9lxuqxpogaaba1", true, false)

	link.ModifiedAt = Timestamp().Add(-7* Day)
	link.ShareExpires = 0

	assert.False(t, link.Expired())

	link.ShareExpires = oneDay

	assert.False(t, link.Expired())

	link.ShareExpires = oneDay * 8

	assert.True(t, link.Expired())
}

func TestLink_Redeem(t *testing.T) {
	link := NewLink(rnd.PPID('a'), false, false)

	assert.Equal(t, uint(0), link.ShareViews)

	link.Redeem()

	assert.Equal(t, uint(1), link.ShareViews)

	if err := link.Save(); err != nil {
		t.Fatal(err)
	}

	link.Redeem()

	assert.Equal(t, uint(2), link.ShareViews)
}

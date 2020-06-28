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
	assert.Equal(t, 10, len(link.LinkToken))
	assert.Equal(t, 16, len(link.LinkUID))
}

func TestLink_Expired(t *testing.T) {
	const oneDay = 60 * 60 * 24

	link := NewLink("st9lxuqxpogaaba1", true, false)

	link.ModifiedAt = Timestamp().Add(-7 * Day)
	link.LinkExpires = 0

	assert.False(t, link.Expired())

	link.LinkExpires = oneDay

	assert.False(t, link.Expired())

	link.LinkExpires = oneDay * 8

	assert.True(t, link.Expired())

	link.LinkExpires = oneDay
	link.LinkViews = 9
	link.MaxViews = 10

	assert.False(t, link.Expired())

	link.Redeem()

	assert.True(t, link.Expired())
}

func TestLink_Redeem(t *testing.T) {
	link := NewLink(rnd.PPID('a'), false, false)

	assert.Equal(t, uint(0), link.LinkViews)

	link.Redeem()

	assert.Equal(t, uint(1), link.LinkViews)

	if err := link.Save(); err != nil {
		t.Fatal(err)
	}

	link.Redeem()

	assert.Equal(t, uint(2), link.LinkViews)
}

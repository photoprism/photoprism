package entity

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewLink(t *testing.T) {
	link := NewLink("st9lxuqxpogaaba1", true, false)
	assert.Equal(t, "st9lxuqxpogaaba1", link.ShareUID)
	assert.Equal(t, false, link.CanEdit)
	assert.Equal(t, true, link.CanComment)
	assert.Equal(t, 10, len(link.ShareToken))
	assert.Equal(t, 16, len(link.LinkUID))
}

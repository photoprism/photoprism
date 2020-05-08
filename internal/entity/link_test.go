package entity

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewLink(t *testing.T) {
	link := NewLink("passwd12", true, false)
	assert.Equal(t, "passwd12", link.LinkPassword)
	assert.Equal(t, false, link.CanEdit)
	assert.Equal(t, true, link.CanComment)
	assert.Equal(t, 10, len(link.LinkToken))
}

package rnd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsSessionID(t *testing.T) {
	assert.True(t, IsSessionID("69be27ac5ca305b394046a83f6fda18167ca3d3f2dbe7ac2"))
	assert.True(t, IsSessionID(SessionID()))
	assert.True(t, IsSessionID(SessionID()))
	assert.False(t, IsSessionID("55785BAC-9H4B-4747-B090-EE123FFEE437"))
	assert.False(t, IsSessionID("4B1FEF2D1CF4A5BE38B263E0637EDEAD"))
}

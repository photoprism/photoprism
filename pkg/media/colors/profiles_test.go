package colors

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProfile_Equal(t *testing.T) {
	assert.True(t, ProfileDisplayP3.Equal("Display P3"))
	assert.False(t, ProfileDisplayP3.Equal(""))
}

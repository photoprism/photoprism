package colors

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLuminance_Hex(t *testing.T) {
	lum := Luminance(1)
	assert.Equal(t, "1", lum.Hex())
}

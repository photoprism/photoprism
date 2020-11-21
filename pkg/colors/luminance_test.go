package colors

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLuminance_Hex(t *testing.T) {
	lum := Luminance(1)
	assert.Equal(t, "1", lum.Hex())
}

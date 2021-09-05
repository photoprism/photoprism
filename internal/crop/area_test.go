package crop

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestArea_String(t *testing.T) {
	t.Run("3e814d3e81f4", func(t *testing.T) {
		expected := fmt.Sprintf("%x%x00%x%x", 1000, 333, 1, 500)
		m := NewArea("face", 1.000, 0.33333, 0.001, 0.5)
		assert.Equal(t, expected, m.String())
	})
	t.Run("3360a7064042_face", func(t *testing.T) {
		m := NewArea("face", 0.822059, 0.167969, 0.1, 0.0664062)
		assert.Equal(t, "3360a7064042", m.String())
	})
	t.Run("3360a7064042_back", func(t *testing.T) {
		m := NewArea("back", 0.822059, 0.167969, 0.1, 0.0664062)
		assert.Equal(t, "3360a7064042", m.String())
	})
	t.Run("0c93e801e000", func(t *testing.T) {
		m := NewArea("face", 0.201, 1.000, 0.03, 0.00000001)
		assert.Equal(t, "0c93e801e000", m.String())
	})
	t.Run("0003e8000000", func(t *testing.T) {
		m := NewArea("face", 0.0001, 1.000, 0, 0.00000001)
		assert.Equal(t, "0003e8000000", m.String())
	})
	t.Run("00007b0003e8", func(t *testing.T) {
		m := NewArea("", -2.0001, 0.123, -0.1, 4.00000001)
		assert.Equal(t, "00007b0003e8", m.String())
	})
}

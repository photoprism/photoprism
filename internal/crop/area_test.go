package crop

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestArea_String(t *testing.T) {
	t.Run("082016010006_face", func(t *testing.T) {
		m := NewArea("face", 0.822059, 0.167969, 0.1, 0.0664062)
		assert.Equal(t, "082016010006", m.String())
	})
	t.Run("082016010006_back", func(t *testing.T) {
		m := NewArea("back", 0.822059, 0.167969, 0.1, 0.0664062)
		assert.Equal(t, "082016010006", m.String())
	})
	t.Run("020100003000", func(t *testing.T) {
		m := NewArea("face", 0.201, 1.000, 0.03, 0.00000001)
		assert.Equal(t, "020100003000", m.String())
	})
	t.Run("000100000000", func(t *testing.T) {
		m := NewArea("face", 0.0001, 1.000, 0, 0.00000001)
		assert.Equal(t, "000100000000", m.String())
	})
	t.Run("000012000100", func(t *testing.T) {
		m := NewArea("", -2.0001, 0.123, -0.1, 4.00000001)
		assert.Equal(t, "000012000100", m.String())
	})
}

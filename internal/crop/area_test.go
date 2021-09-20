package crop

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestArea_String(t *testing.T) {
	t.Run("3e814d0011f4", func(t *testing.T) {
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
	t.Run("00007b0003e8", func(t *testing.T) {
		m := NewArea("", 0, 0, 0, 0)
		assert.Equal(t, "", m.String())
	})
}

func TestArea_Thumb(t *testing.T) {
	t.Run("3e814d0011f4", func(t *testing.T) {
		expected := fmt.Sprintf("346b3897eec9ef75e35fbf0bbc4c83c55ca41e31-%x%x00%x%x", 1000, 333, 1, 500)
		m := NewArea("face", 1.000, 0.33333, 0.001, 0.5)
		assert.Equal(t, expected, m.Thumb("346b3897eec9ef75e35fbf0bbc4c83c55ca41e31"))
	})
	t.Run("3360a7064042_face", func(t *testing.T) {
		m := NewArea("face", 0.822059, 0.167969, 0.1, 0.0664062)
		assert.Equal(t, "346b3897eec9ef75e35fbf0bbc4c83c55ca41e31-3360a7064042", m.Thumb("346b3897eec9ef75e35fbf0bbc4c83c55ca41e31"))
	})
	t.Run("3360a7064042_back", func(t *testing.T) {
		m := NewArea("back", 0.822059, 0.167969, 0.1, 0.0664062)
		assert.Equal(t, "346b3897eec9ef75e35fbf0bbc4c83c55ca41e31-3360a7064042", m.Thumb("346b3897eec9ef75e35fbf0bbc4c83c55ca41e31"))
	})
	t.Run("0c93e801e000", func(t *testing.T) {
		m := NewArea("face", 0.201, 1.000, 0.03, 0.00000001)
		assert.Equal(t, "346b3897eec9ef75e35fbf0bbc4c83c55ca41e31-0c93e801e000", m.Thumb("346b3897eec9ef75e35fbf0bbc4c83c55ca41e31"))
	})
	t.Run("0003e8000000", func(t *testing.T) {
		m := NewArea("face", 0.0001, 1.000, 0, 0.00000001)
		assert.Equal(t, "346b3897eec9ef75e35fbf0bbc4c83c55ca41e31-0003e8000000", m.Thumb("346b3897eec9ef75e35fbf0bbc4c83c55ca41e31"))
	})
	t.Run("ShortHash", func(t *testing.T) {
		m := NewArea("", -2.0001, 0.123, -0.1, 4.00000001)
		assert.Equal(t, "00007b0003e8", m.Thumb("346b3897eec9ef75e35fbf0bbc4c83c55ca41"))
	})
	t.Run("Empty", func(t *testing.T) {
		m := NewArea("", 0, 0, 0, 0)
		assert.Equal(t, "", m.Thumb("346b3897eec9ef75e35fbf0bbc4c83c55ca41"))
	})
}

func TestAreaFromString(t *testing.T) {
	t.Run("3e814d0011f4", func(t *testing.T) {
		a := AreaFromString("3e814d0011f4")
		assert.Equal(t, float32(1), a.X)
		assert.Equal(t, float32(0.333), a.Y)
		assert.Equal(t, float32(0.001), a.W)
		assert.Equal(t, float32(0.5), a.H)
	})
	t.Run("3360a7064042", func(t *testing.T) {
		a := AreaFromString("3360a7064042")
		assert.Equal(t, NewArea("crop", 0.822, 0.167, 0.1, 0.066), a)
	})
}

func TestIsCroppedThumb(t *testing.T) {
	t.Run("True", func(t *testing.T) {
		assert.Equal(t, 40, IsCroppedThumb("346b3897eec9ef75e35fbf0bbc4c83c55ca41e31-00007b0003e8"))
	})
	t.Run("Invalid", func(t *testing.T) {
		assert.Equal(t, -1, IsCroppedThumb("346b3897eec9ef75e35fbf0bbc4c83c55ca41e3100007b0003e8-"))
	})
	t.Run("CropArea", func(t *testing.T) {
		assert.Equal(t, -1, IsCroppedThumb("00007b0003e8"))
	})
	t.Run("ShortHash", func(t *testing.T) {
		assert.Equal(t, -1, IsCroppedThumb("346b3897eec9ef75e35fbf0bbc4c83c55ca41"))
	})
	t.Run("HashOnly", func(t *testing.T) {
		assert.Equal(t, -1, IsCroppedThumb("346b3897eec9ef75e35fbf0bbc4c83c55ca41e31"))
	})
}

func TestParseThumb(t *testing.T) {
	t.Run("True", func(t *testing.T) {
		h, a := ParseThumb("346b3897eec9ef75e35fbf0bbc4c83c55ca41e31-00007b0003e8")
		assert.Equal(t, 40, len(h))
		assert.Equal(t, 12, len(a))
		assert.False(t, AreaFromString(a).Empty())
	})
	t.Run("Invalid", func(t *testing.T) {
		h, a := ParseThumb("346b3897eec9ef75e35fbf0bbc4c83c55ca41e3100007b0003e8-")
		assert.Equal(t, "346b3897eec9ef75e35fbf0bbc4c83c55ca41e3100007b0003e8", h)
		assert.Equal(t, 0, len(a))
		assert.True(t, AreaFromString(a).Empty())
	})
	t.Run("CropArea", func(t *testing.T) {
		h, a := ParseThumb("00007b0003e8")
		assert.Equal(t, 0, len(h))
		assert.Equal(t, 12, len(a))
		assert.False(t, AreaFromString(a).Empty())
	})
	t.Run("ShortHash", func(t *testing.T) {
		h, a := ParseThumb("346b3897eec9ef75e35fbf0bbc4c83c55ca41")
		assert.Equal(t, 37, len(h))
		assert.Equal(t, "346b3897eec9ef75e35fbf0bbc4c83c55ca41", h)
		assert.Equal(t, 0, len(a))
		assert.True(t, AreaFromString(a).Empty())
	})
	t.Run("HashOnly", func(t *testing.T) {
		h, a := ParseThumb("346b3897eec9ef75e35fbf0bbc4c83c55ca41e31")
		assert.Equal(t, 40, len(h))
		assert.Equal(t, "346b3897eec9ef75e35fbf0bbc4c83c55ca41e31", h)
		assert.Equal(t, 0, len(a))
		assert.True(t, AreaFromString(a).Empty())
	})
}

package colors

import (
	"image/color"
	"testing"

	"github.com/lucasb-eyer/go-colorful"
	"github.com/stretchr/testify/assert"
)

func TestColorful(t *testing.T) {
	t.Run("purple", func(t *testing.T) {
		c := color.RGBA{0x6a, 0x1b, 0x9a, 0xff}
		color, ok := colorful.MakeColor(c)

		if !ok {
			t.Fatal("ok should be true")
		}

		assert.Equal(t, "purple", Colorful(color).Name())
	})

	t.Run("cyan", func(t *testing.T) {
		c := color.RGBA{0xb2, 0xeb, 0xf2, 0xff}
		color, ok := colorful.MakeColor(c)

		if !ok {
			t.Fatal("ok should be true")
		}

		assert.Equal(t, "cyan", Colorful(color).Name())
	})
}

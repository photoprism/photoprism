package form

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewFile(t *testing.T) {
	t.Run("Orientation", func(t *testing.T) {
		var file = struct {
			Orientation int
		}{
			Orientation: 3,
		}

		frm, err := NewFile(file)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, 3, frm.Orientation)
	})
}

package form

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewFile(t *testing.T) {
	t.Run("Orientation", func(t *testing.T) {
		var file = struct {
			FileOrientation int
		}{
			FileOrientation: 3,
		}

		frm, err := NewFile(file)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, 3, frm.FileOrientation)
		assert.Equal(t, 3, frm.Orientation())
		frm.FileOrientation = 10
		assert.Equal(t, 10, frm.FileOrientation)
		assert.Equal(t, 0, frm.Orientation())
	})
}

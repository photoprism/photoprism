package form

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewLens(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		var l = struct {
			LensMake  string
			LensModel string
		}{
			LensMake:  "New Make",
			LensModel: "New Model",
		}

		result, err := NewLens(l)

		if err != nil {
			t.Fatal(err)
		}

		assert.IsType(t, &Lens{}, result)
		assert.Equal(t, "New Make", result.LensMake)
		assert.Equal(t, "New Model", result.LensModel)
	})
}

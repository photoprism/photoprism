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

func TestLens_Validate(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		frm := &Lens{LensMake: "Canon", LensModel: "EF 50mm f/1.8"}
		assert.NoError(t, frm.Validate())
	})
	t.Run("EmptyMake", func(t *testing.T) {
		frm := &Lens{LensMake: "", LensModel: "EF 50mm f/1.8"}
		assert.Error(t, frm.Validate())
	})
	t.Run("EmptyModel", func(t *testing.T) {
		frm := &Lens{LensMake: "Canon", LensModel: ""}
		assert.Error(t, frm.Validate())
	})
}

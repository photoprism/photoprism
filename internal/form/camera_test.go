package form

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewCamera(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		var c = struct {
			CameraMake  string
			CameraModel string
		}{
			CameraMake:  "New Make",
			CameraModel: "New Model",
		}

		result, err := NewCamera(c)

		if err != nil {
			t.Fatal(err)
		}

		assert.IsType(t, &Camera{}, result)
		assert.Equal(t, "New Make", result.CameraMake)
		assert.Equal(t, "New Model", result.CameraModel)
	})
}

func TestCamera_Validate(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		frm := &Camera{CameraMake: "Canon", CameraModel: "EOS 5D"}
		assert.NoError(t, frm.Validate())
	})
	t.Run("EmptyMake", func(t *testing.T) {
		frm := &Camera{CameraMake: "", CameraModel: "EOS 5D"}
		assert.Error(t, frm.Validate())
	})
	t.Run("EmptyModel", func(t *testing.T) {
		frm := &Camera{CameraMake: "Canon", CameraModel: ""}
		assert.Error(t, frm.Validate())
	})
}

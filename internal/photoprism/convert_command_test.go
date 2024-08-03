package photoprism

import (
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/pkg/media"
)

func TestNewConvertCommand(t *testing.T) {
	t.Run("Nil", func(t *testing.T) {
		assert.Nil(t, NewConvertCommand(nil))
	})
	t.Run("Default", func(t *testing.T) {
		result := NewConvertCommand(
			exec.Command("/usr/bin/sips", "-Z", "123", "-s", "format", "jpeg", "--out", "file.jpeg", "file.heic"),
		)
		assert.NotNil(t, result)
		assert.NotNil(t, result.Cmd)
		assert.Equal(t, "/usr/bin/sips -Z 123 -s format jpeg --out file.jpeg file.heic", result.String())
		assert.Equal(t, media.KeepOrientation, result.Orientation)
	})
	t.Run("WithOrientation", func(t *testing.T) {
		result := NewConvertCommand(
			exec.Command("/usr/bin/sips", "-Z", "123", "-s", "format", "jpeg", "--out", "file.jpeg", "file.heic"),
		)
		result.WithOrientation(media.ResetOrientation)
		assert.NotNil(t, result)
		assert.NotNil(t, result.Cmd)
		assert.Equal(t, "/usr/bin/sips -Z 123 -s format jpeg --out file.jpeg file.heic", result.String())
		assert.Equal(t, media.ResetOrientation, result.Orientation)
	})
}

func TestNewConvertCommands(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		result := NewConvertCommands()
		assert.NotNil(t, result)
	})
}

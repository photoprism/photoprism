package photoprism

import (
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/pkg/media"
)

func TestNewConvertCmd(t *testing.T) {
	t.Run("Nil", func(t *testing.T) {
		assert.Nil(t, NewConvertCmd(nil))
	})
	t.Run("Default", func(t *testing.T) {
		result := NewConvertCmd(
			exec.Command("/usr/bin/sips", "-Z", "123", "-s", "format", "jpeg", "--out", "file.jpeg", "file.heic"),
		)
		assert.NotNil(t, result)
		assert.NotNil(t, result.Cmd)
		assert.Equal(t, "/usr/bin/sips -Z 123 -s format jpeg --out file.jpeg file.heic", result.String())
		assert.Equal(t, media.KeepOrientation, result.Orientation)
	})
	t.Run("WithOrientation", func(t *testing.T) {
		result := NewConvertCmd(
			exec.Command("/usr/bin/sips", "-Z", "123", "-s", "format", "jpeg", "--out", "file.jpeg", "file.heic"),
		)
		result.WithOrientation(media.ResetOrientation)
		assert.NotNil(t, result)
		assert.NotNil(t, result.Cmd)
		assert.Equal(t, "/usr/bin/sips -Z 123 -s format jpeg --out file.jpeg file.heic", result.String())
		assert.Equal(t, media.ResetOrientation, result.Orientation)
	})
	t.Run("WithImageVerification", func(t *testing.T) {
		result := NewConvertCmd(
			exec.Command("exiftool", "-q", "-q", "-b", "-PreviewImage", "file.cr3"),
		)
		assert.False(t, result.VerifyImage)
		assert.Same(t, result, result.WithImageVerification())
		assert.True(t, result.VerifyImage)
	})
	t.Run("WithStderrRejection", func(t *testing.T) {
		result := NewConvertCmd(
			exec.Command("rawtherapee-cli", "-o", "file.jpg", "-c", "file.cr3"),
		)
		assert.Empty(t, result.RejectStderr)
		assert.Same(t, result, result.WithStderrRejection("first error", "second error"))
		assert.Equal(t, []string{"first error", "second error"}, result.RejectStderr)
	})
}

func TestConvertCmd_StderrRejected(t *testing.T) {
	cmd := NewConvertCmd(exec.Command("rawtherapee-cli", "-c", "file.cr3")).WithStderrRejection("Cannot use camera white balance")
	t.Run("Match", func(t *testing.T) {
		assert.True(t, cmd.StderrRejected("Processing...\nCannot use camera white balance.\n"))
	})
	t.Run("NoMatch", func(t *testing.T) {
		assert.False(t, cmd.StderrRejected("Warning: sidecar file requested but not found for: file.cr3"))
	})
	t.Run("EmptyStderr", func(t *testing.T) {
		assert.False(t, cmd.StderrRejected(""))
	})
	t.Run("NoPatterns", func(t *testing.T) {
		plain := NewConvertCmd(exec.Command("darktable-cli", "file.cr3", "file.jpg"))
		assert.False(t, plain.StderrRejected("Cannot use camera white balance."))
	})
}

func TestNewConvertCmds(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		result := NewConvertCmds()
		assert.NotNil(t, result)
	})
}

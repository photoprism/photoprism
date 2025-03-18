package fs

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseMode(t *testing.T) {
	t.Run("Default", func(t *testing.T) {
		mode := ParseMode("", ModeSocket)
		assert.Equal(t, ModeSocket, mode)
		assert.Equal(t, os.FileMode(0o666), mode)
	})
	t.Run("777", func(t *testing.T) {
		mode := ParseMode("777", ModeSocket)
		assert.Equal(t, os.ModePerm, mode)
		assert.Equal(t, os.FileMode(0o777), mode)
	})
	t.Run("0777", func(t *testing.T) {
		mode := ParseMode("0777", ModeSocket)
		assert.Equal(t, os.ModePerm, mode)
		assert.Equal(t, os.FileMode(0o777), mode)
	})
	t.Run("0770", func(t *testing.T) {
		mode := ParseMode("0770", ModeSocket)
		assert.Equal(t, os.FileMode(0o770), mode)
	})
	t.Run("0666", func(t *testing.T) {
		mode := ParseMode("0666", ModeSocket)
		assert.Equal(t, os.FileMode(0o666), mode)
	})
	t.Run("0660", func(t *testing.T) {
		mode := ParseMode("0660", ModeSocket)
		assert.Equal(t, os.FileMode(0o660), mode)
	})
	t.Run("660", func(t *testing.T) {
		mode := ParseMode("660", ModeSocket)
		assert.Equal(t, os.FileMode(0o660), mode)
	})
}

package meta

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDuration(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		d := Duration("")
		assert.Equal(t, "0s", d.String())
	})
	t.Run("Zero", func(t *testing.T) {
		d := Duration("0")
		assert.Equal(t, "0s", d.String())
	})
	t.Run("ZeroFive", func(t *testing.T) {
		d := Duration("0.5")
		assert.Equal(t, "500ms", d.String())
	})
	t.Run("TwoNum41S", func(t *testing.T) {
		d := Duration("2.41 s")
		assert.Equal(t, "2.41s", d.String())
	})
	t.Run("ZeroNum41S", func(t *testing.T) {
		d := Duration("0.41 s")
		assert.Equal(t, "410ms", d.String())
	})
	t.Run("Num41S", func(t *testing.T) {
		d := Duration("41 s")
		assert.Equal(t, "41s", d.String())
	})
	t.Run("ZeroZeroOne", func(t *testing.T) {
		d := Duration("0:0:1")
		assert.Equal(t, "1s", d.String())
	})
	t.Run("ZeroNum04Num25", func(t *testing.T) {
		d := Duration("0:04:25")
		assert.Equal(t, "4m25s", d.String())
	})
	t.Run("Num0001Num04Num25", func(t *testing.T) {
		d := Duration("0001:04:25")
		assert.Equal(t, "1h4m25s", d.String())
	})
	t.Run("Invalid", func(t *testing.T) {
		d := Duration("01:04:25:67")
		assert.Equal(t, "0s", d.String())
	})
}

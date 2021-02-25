package meta

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStringToDuration(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		d := StringToDuration("")
		assert.Equal(t, "0s", d.String())
	})

	t.Run("0", func(t *testing.T) {
		d := StringToDuration("0")
		assert.Equal(t, "0s", d.String())
	})

	t.Run("2.41 s", func(t *testing.T) {
		d := StringToDuration("2.41 s")
		assert.Equal(t, "2s", d.String())
	})

	t.Run("0.41 s", func(t *testing.T) {
		d := StringToDuration("0.41 s")
		assert.Equal(t, "0s", d.String())
	})

	t.Run("41 s", func(t *testing.T) {
		d := StringToDuration("41 s")
		assert.Equal(t, "41s", d.String())
	})

	t.Run("0:0:1", func(t *testing.T) {
		d := StringToDuration("0:0:1")
		assert.Equal(t, "1s", d.String())
	})

	t.Run("0:04:25", func(t *testing.T) {
		d := StringToDuration("0:04:25")
		assert.Equal(t, "4m25s", d.String())
	})

	t.Run("0001:04:25", func(t *testing.T) {
		d := StringToDuration("0001:04:25")
		assert.Equal(t, "1h4m25s", d.String())
	})

	t.Run("invalid", func(t *testing.T) {
		d := StringToDuration("01:04:25:67")
		assert.Equal(t, "0s", d.String())
	})
}

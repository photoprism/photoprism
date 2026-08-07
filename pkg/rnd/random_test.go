package rnd

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRandomFill(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		b := make([]byte, 32)
		randomFill(b)
		assert.Len(t, b, 32)
		// An all-zero buffer would mean the slice was never written.
		assert.NotEqual(t, make([]byte, 32), b)
	})
	t.Run("Empty", func(t *testing.T) {
		assert.NotPanics(t, func() { randomFill([]byte{}) })
	})
}

func TestRandomChars(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		b := make([]byte, 64)
		randomChars(b, CharsetBase62)
		assert.Len(t, b, 64)
		for _, c := range b {
			assert.Contains(t, CharsetBase62, string(c))
		}
	})
	t.Run("SingleCharCharset", func(t *testing.T) {
		b := make([]byte, 8)
		randomChars(b, "x")
		assert.Equal(t, "xxxxxxxx", string(b))
	})
	t.Run("Empty", func(t *testing.T) {
		assert.NotPanics(t, func() { randomChars([]byte{}, CharsetBase62) })
	})
	t.Run("NoCharset", func(t *testing.T) {
		assert.Panics(t, func() { randomChars(make([]byte, 1), "") })
	})
	t.Run("CharsetTooLong", func(t *testing.T) {
		assert.Panics(t, func() { randomChars(make([]byte, 1), strings.Repeat("x", 257)) })
	})
	t.Run("Uniform", func(t *testing.T) {
		// Every character must stay equally likely. A plain "% len(charset)" would
		// over-represent the first 256%62 characters of CharsetBase62 by about 25%.
		const samples = 620000

		b := make([]byte, samples)
		randomChars(b, CharsetBase62)

		counts := make(map[byte]int, len(CharsetBase62))
		for _, c := range b {
			counts[c]++
		}

		assert.Len(t, counts, len(CharsetBase62))

		expected := samples / len(CharsetBase62)
		for i := range len(CharsetBase62) {
			c := CharsetBase62[i]
			assert.InEpsilon(t, expected, counts[c], 0.1, "character %q occurred %d times", string(c), counts[c])
		}
	})
}

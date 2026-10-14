package clean

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestModelId(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		id, err := ModelId("unsloth/Qwen3.5-9B-GGUF:Q4_K_M")
		assert.NoError(t, err)
		assert.Equal(t, "unsloth/Qwen3.5-9B-GGUF:Q4_K_M", id)
	})
	t.Run("SurroundingWhitespace", func(t *testing.T) {
		id, err := ModelId(" \tgemma3:4b\n ")
		assert.NoError(t, err)
		assert.Equal(t, "gemma3:4b", id)
	})
	t.Run("Empty", func(t *testing.T) {
		id, err := ModelId("   ")
		assert.NoError(t, err)
		assert.Equal(t, "", id)
	})
	t.Run("InvalidChars", func(t *testing.T) {
		for in, want := range map[string]string{
			"gemma3\x00:4b": "gemma3:4b",
			"gemma3\t:4b":   "gemma3 :4b",
			"gemmä3:4b":     "gemm3:4b",
			"\x00gemma3":    "gemma3",
		} {
			id, err := ModelId(in)
			assert.ErrorIs(t, err, ErrInvalidChars, "%q", in)
			assert.False(t, errors.Is(err, ErrTooLong), "%q", in)
			assert.Equal(t, want, id, "%q", in)
		}
	})
	t.Run("TooLong", func(t *testing.T) {
		id, err := ModelId(strings.Repeat("m", LengthType+1))
		assert.ErrorIs(t, err, ErrTooLong)
		assert.False(t, errors.Is(err, ErrInvalidChars))
		assert.Equal(t, strings.Repeat("m", LengthType), id)
	})
	t.Run("InvalidCharsAndTooLong", func(t *testing.T) {
		id, err := ModelId("\x00" + strings.Repeat("m", LengthType+1))
		assert.ErrorIs(t, err, ErrInvalidChars)
		assert.ErrorIs(t, err, ErrTooLong)
		assert.Len(t, id, LengthType)
	})
}

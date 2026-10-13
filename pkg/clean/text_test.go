package clean

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestText(t *testing.T) {
	t.Run("Unchanged", func(t *testing.T) {
		for _, s := range []string{
			"",
			"A caption",
			"First line\nSecond line\n\n\tIndented",
			"صورة من العائلة\nשלום",
			"日本語の説明",
			"👩\u200D👩\u200D👧 at the beach",
			"می\u200Cخواهم",
			"a\xffb",
		} {
			assert.Equal(t, s, Text(s))
		}
		assert.Zero(t, testing.AllocsPerRun(10, func() { _ = Text("First line\nSecond line") }))
	})
	t.Run("LineBreaks", func(t *testing.T) {
		assert.Equal(t, "a\nb\nc\n\nd", Text("a\r\nb\rc\n\r\nd"))
		assert.Equal(t, "a\n", Text("a\r"))
		assert.Equal(t, "\n\n", Text("\r\r"))
	})
	t.Run("Whitespace", func(t *testing.T) {
		assert.Equal(t, "a b c\td", Text("a\vb\fc\td"))
	})
	t.Run("Removed", func(t *testing.T) {
		assert.Equal(t, "abcd", Text("a\x00b\x1bc\x7fd"))
		assert.Equal(t, "", Text("\x00\x01\x1f\x7f"))
		assert.Equal(t, "Ä幸", Text("Ä\x00幸"))
	})
	t.Run("ControlTable", func(t *testing.T) {
		assert.Equal(t, "Abcd\te\nfgh\xffi", Text(controlSample))
	})
}

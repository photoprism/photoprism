package clean

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJSONEscape(t *testing.T) {
	t.Run("Unchanged", func(t *testing.T) {
		for _, s := range []string{
			`{"name":"IMG_1234.jpg"}`,
			"[\"صورة 日本.jpg\",\"👩\u200D👩\u200D👧\",\"a\u00A0b\"]",
			`{"text":"line\nnext\t\u001b"}`,
			"{\n  \"name\": \"a\"\n}",
			"",
		} {
			assert.Equal(t, s, string(JSONEscape([]byte(s))))
		}
		assert.Zero(t, testing.AllocsPerRun(10, func() { _ = JSONEscape([]byte(`{"name":"IMG_1234.jpg"}`)) }))
	})
	t.Run("Escaped", func(t *testing.T) {
		in := map[string]string{"name": "a\u202Eb\u007Fc\u0085d\u200Be\U000E0041f"}
		data, err := json.Marshal(in)
		require.NoError(t, err)

		out := JSONEscape(data)
		assert.Equal(t, `{"name":"a\u202eb\u007fc\u0085d\u200be\udb40\udc41f"}`, string(out))

		var decoded map[string]string
		require.NoError(t, json.Unmarshal(out, &decoded))
		assert.Equal(t, in, decoded)
	})
	t.Run("InvalidUTF8", func(t *testing.T) {
		assert.Equal(t, "a\\ufffdb\\u202e", string(JSONEscape([]byte("a\xffb\u202E"))))
		assert.Equal(t, "a\\ufffdb", string(JSONEscape([]byte("a\xffb"))))
	})
}

func TestJsonEscapeRune(t *testing.T) {
	assert.True(t, jsonEscapeRune(0x7F, 1))
	assert.False(t, jsonEscapeRune('\n', 1))
	assert.False(t, jsonEscapeRune('\t', 1))
	assert.True(t, jsonEscapeRune(0x202E, 3))
	assert.True(t, jsonEscapeRune(0x0085, 2))
	assert.False(t, jsonEscapeRune('a', 1))
	assert.False(t, jsonEscapeRune(0x00A0, 2))
	assert.False(t, jsonEscapeRune(0x200D, 3))
	assert.True(t, jsonEscapeRune(0xFFFD, 1))
	assert.False(t, jsonEscapeRune(0xFFFD, 3))
}

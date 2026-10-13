package meta

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestJsonText(t *testing.T) {
	t.Run("Array", func(t *testing.T) {
		assert.Equal(t, "cat\x00dog, sea", jsonText(gjson.Parse(`["cat\u0000dog","sea"]`)))
		assert.Equal(t, "say \"hi\", a\\b, x/y", jsonText(gjson.Parse(`["say \"hi\"","a\\b","x\/y"]`)))
		assert.Equal(t, "a, 7", jsonText(gjson.Parse(`["a","",7]`)))
		assert.Equal(t, "", jsonText(gjson.Parse(`[]`)))
	})
	t.Run("Value", func(t *testing.T) {
		assert.Equal(t, "cat\x00dog", jsonText(gjson.Parse(`"cat\u0000dog"`)))
		assert.Equal(t, "42", jsonText(gjson.Parse(`42`)))
	})
}

func TestData_ExiftoolKeywordEscapes(t *testing.T) {
	var data Data

	body := `[{"SourceFile":"x.jpg","ExifToolVersion":13.0,"Keywords":["cat\u0000dog","sea\nside","\"quoted\""]}]`

	require.NoError(t, data.Exiftool([]byte(body), ""))
	assert.Equal(t, "catdog, quoted, sea, side", data.Keywords.String())
}

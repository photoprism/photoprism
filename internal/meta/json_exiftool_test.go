package meta

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

// TestData_ExiftoolJpegFrames verifies that JPEG capture frames are not animation frames.
func TestData_ExiftoolJpegFrames(t *testing.T) {
	t.Run("PixelHDR", func(t *testing.T) {
		data, err := JSON("testdata/pixel-hdr.jpg.json", "")
		require.NoError(t, err)
		assert.Zero(t, data.Frames)
		assert.Zero(t, data.FPS)
		assert.Zero(t, data.Duration)
		assert.False(t, data.HasVideoEmbedded)
	})
	t.Run("FileTypeWithoutMime", func(t *testing.T) {
		var data Data
		require.NoError(t, data.Exiftool([]byte(`[{"FileType":"JPEG","FrameCount":6}]`), ""))
		assert.Zero(t, data.Frames)
	})
	t.Run("MotionPhoto", func(t *testing.T) {
		var data Data
		require.NoError(t, data.Exiftool([]byte(`[{"MIMEType":"image/jpeg","FrameCount":6,"MotionPhoto":1,"Duration":2,"VideoFrameRate":24}]`), ""))
		assert.Zero(t, data.Frames)
		assert.True(t, data.HasVideoEmbedded)
		assert.Equal(t, 2*time.Second, data.Duration)
		assert.Equal(t, float64(24), data.FPS)
	})
	t.Run("AnimationAndVideo", func(t *testing.T) {
		for _, mime := range []string{"image/gif", "image/png", "image/webp", "image/avif", "video/mp4"} {
			var data Data
			require.NoError(t, data.Exiftool([]byte(`[{"MIMEType":"`+mime+`","FrameCount":6}]`), ""))
			assert.Equal(t, 6, data.Frames, mime)
		}
	})
}

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

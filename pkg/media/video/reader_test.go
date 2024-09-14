package video

import (
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gabriel-vasile/mimetype"
)

func TestReader(t *testing.T) {
	t.Run("Read", func(t *testing.T) {
		info, probeErr := ProbeFile("testdata/image-isom-avc1.jpg")

		if probeErr != nil {
			t.Fatal(probeErr)
		}

		require.NotNil(t, info)

		reader, readerErr := NewReader(info.FileName, info.VideoOffset)

		if readerErr != nil {
			t.Fatal(probeErr)
		}

		defer reader.Close()

		require.NotNil(t, reader)

		videoData, ioErr := io.ReadAll(reader)

		if ioErr != nil {
			t.Fatal(probeErr)
		}

		stat, statErr := os.Stat(info.FileName)

		if statErr != nil {
			t.Fatal(probeErr)
		}

		assert.True(t, int(stat.Size()) > len(videoData))
		assert.Equal(t, int(stat.Size()-info.VideoOffset), len(videoData))
		assert.Equal(t, info.VideoMimeType, mimetype.Detect(videoData).String())
	})
}

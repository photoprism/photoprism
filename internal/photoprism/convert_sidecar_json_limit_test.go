package photoprism

import (
	"bytes"
	"encoding/binary"
	"hash/crc32"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/meta"
	"github.com/photoprism/photoprism/pkg/fs"
)

// TestJSONOutputBuffer checks inclusive capture limits and diagnostic truncation.
func TestJSONOutputBuffer(t *testing.T) {
	t.Run("JSONOutput", func(t *testing.T) {
		b := jsonOutputBuffer{limit: 4, failOnLimit: true}
		n, err := b.Write([]byte("ab"))
		require.NoError(t, err)
		assert.Equal(t, 2, n)
		n, err = b.Write([]byte("cd"))
		require.NoError(t, err)
		assert.Equal(t, 2, n)
		assert.False(t, b.exceeded)
		n, err = b.Write([]byte("ef"))
		assert.ErrorIs(t, err, meta.ErrJSONFileTooLarge)
		assert.Zero(t, n)
		assert.True(t, b.exceeded)
		assert.Equal(t, "abcd", string(b.data))
	})
	t.Run("PartialWrite", func(t *testing.T) {
		b := jsonOutputBuffer{limit: 4, failOnLimit: true}
		n, err := b.Write([]byte("abcdef"))
		assert.ErrorIs(t, err, meta.ErrJSONFileTooLarge)
		assert.Equal(t, 4, n)
		assert.Equal(t, "abcd", string(b.data))
	})
	t.Run("Diagnostics", func(t *testing.T) {
		b := jsonOutputBuffer{limit: 4}
		n, err := b.Write([]byte("abcdef"))
		require.NoError(t, err)
		assert.Equal(t, 6, n)
		n, err = b.Write([]byte(strings.Repeat("x", 100)))
		require.NoError(t, err)
		assert.Equal(t, 100, n)
		assert.True(t, b.exceeded)
		assert.Equal(t, "abcd", string(b.data))
	})
}

// writeJSONMetadataPNG creates a PNG with a sized text field for ExifTool export tests.
func writeJSONMetadataPNG(t *testing.T, name string, textBytes int) {
	t.Helper()
	var encoded bytes.Buffer
	require.NoError(t, png.Encode(&encoded, image.NewNRGBA(image.Rect(0, 0, 1, 1))))
	payload := append([]byte("Description\x00"), bytes.Repeat([]byte{'A'}, textBytes)...)
	chunk := make([]byte, len(payload)+12)
	binary.BigEndian.PutUint32(chunk[:4], uint32(len(payload))) //nolint:gosec // Test metadata size is bounded below a few MiB.
	copy(chunk[4:8], "tEXt")
	copy(chunk[8:], payload)
	binary.BigEndian.PutUint32(chunk[len(chunk)-4:], crc32.ChecksumIEEE(chunk[4:len(chunk)-4]))
	imageData := encoded.Bytes()
	withMetadata := append([]byte{}, imageData[:33]...)
	withMetadata = append(withMetadata, chunk...)
	withMetadata = append(withMetadata, imageData[33:]...)
	require.NoError(t, os.WriteFile(name, withMetadata, fs.ModeFile))
}

// TestConvertJSONOutputLimit checks real ExifTool output and absence of partial cache files.
func TestConvertJSONOutputLimit(t *testing.T) {
	t.Setenv("PHOTOPRISM_JSON_LIMIT", "")
	c := config.NewMinimalTestConfig(t.TempDir())
	c.Options().DisableExifTool = false
	require.NotEmpty(t, c.ExifToolBin())
	require.NoError(t, c.CreateDirectories())
	previous := Config()
	SetConfig(c)
	t.Cleanup(func() { SetConfig(previous) })
	convert := NewConvert(c)

	for _, tc := range []struct {
		name  string
		size  int
		limit string
	}{
		{"Ordinary", 1024, ""},
		{"Oversized", int(meta.JSONMaxFileBytes) * 2, ""},
		{"Override", int(meta.JSONMaxFileBytes) * 2, "3145728"},
		{"OverrideOversized", int(meta.JSONMaxFileBytes) * 4, "3145728"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("PHOTOPRISM_JSON_LIMIT", tc.limit)
			name := filepath.Join(c.OriginalsPath(), tc.name+".png")
			writeJSONMetadataPNG(t, name, tc.size)
			mediaFile, err := NewMediaFile(name)
			require.NoError(t, err)
			jsonName, err := mediaFile.ExifToolJsonName()
			require.NoError(t, err)

			result, err := convert.ToJson(mediaFile, false)
			if tc.size > int(meta.JSONFileLimit()) {
				assert.ErrorIs(t, err, meta.ErrJSONFileTooLarge)
				assert.Empty(t, result)
				assert.NoFileExists(t, jsonName)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, jsonName, result)
			data, err := meta.JSON(result, "")
			require.NoError(t, err)
			assert.Equal(t, 1, data.Width)
			assert.Equal(t, 1, data.Height)
		})
	}
}

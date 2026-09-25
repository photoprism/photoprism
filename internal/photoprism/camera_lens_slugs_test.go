package photoprism

import (
	"bufio"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/meta"
)

// TestCameraLensSlugs pins the camera and lens slugs of real Exif make and model values, as indexing derives them.
// A changed slug would create a second record for the same device when existing files are indexed again.
func TestCameraLensSlugs(t *testing.T) {
	f, err := os.Open("testdata/camera-lens-slugs.tsv")
	require.NoError(t, err)
	t.Cleanup(func() { _ = f.Close() })

	s := bufio.NewScanner(f)
	require.True(t, s.Scan(), "header")

	rows := 0

	for s.Scan() {
		c := strings.Split(s.Text(), "\t")
		require.Len(t, c, 7, "row %d", rows+1)

		// Parse the raw values like ExifTool JSON sidecars, since the indexer normalizes them there as well.
		fields := map[string]string{}

		for i, k := range []string{"Make", "Model", "LensMake", "LensModel", "LensID"} {
			if c[i] != "" {
				fields[k] = c[i]
			}
		}

		jsonData, err := json.Marshal([]map[string]string{fields})
		require.NoError(t, err)

		data := meta.Data{}
		require.NoError(t, data.Exiftool(jsonData, "camera-lens-slugs.jpg"))

		assert.Equal(t, c[5], entity.NewCamera(data.CameraMake, data.CameraModel).CameraSlug, "camera %q / %q", c[0], c[1])
		assert.Equal(t, c[6], entity.NewLens(data.LensMake, data.LensModel).LensSlug, "lens %q / %q (%q)", c[2], c[3], c[4])

		rows++
	}

	require.NoError(t, s.Err())
	assert.Equal(t, 297, rows)
}

package meta

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/iotest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/pkg/fs"
)

// jsonLimitReader produces counted input without allocating its declared length.
type jsonLimitReader struct {
	remaining int64
	read      int64
}

// Read fills the requested buffer with available input and counts consumed bytes.
func (r *jsonLimitReader) Read(p []byte) (int, error) {
	if r.remaining == 0 {
		return 0, io.EOF
	}

	n := int(min(int64(len(p)), r.remaining))
	for i := range p[:n] {
		p[i] = ' '
	}

	r.remaining -= int64(n)
	r.read += int64(n)
	return n, nil
}

// TestReadJSONSidecar checks early rejection, stale size hints, and read failures.
func TestReadJSONSidecar(t *testing.T) {
	t.Setenv("PHOTOPRISM_JSON_LIMIT", "")
	t.Run("AdvertisedOversize", func(t *testing.T) {
		r := &jsonLimitReader{remaining: 1024}
		data, err := readJSONSidecar(r, 1<<40)
		assert.ErrorIs(t, err, ErrJSONFileTooLarge)
		assert.Nil(t, data)
		assert.Zero(t, r.read)
	})
	t.Run("GrownAfterStat", func(t *testing.T) {
		r := &jsonLimitReader{remaining: JSONMaxFileBytes * 3}
		data, err := readJSONSidecar(r, 0)
		assert.ErrorIs(t, err, ErrJSONFileTooLarge)
		assert.Nil(t, data)
		assert.Equal(t, JSONMaxFileBytes+1, r.read)
	})
	t.Run("ReadFailure", func(t *testing.T) {
		r := io.MultiReader(strings.NewReader("partial"), iotest.ErrReader(io.ErrUnexpectedEOF))
		data, err := readJSONSidecar(r, 0)
		assert.ErrorIs(t, err, io.ErrUnexpectedEOF)
		assert.Nil(t, data)
	})
	t.Run("SmallInput", func(t *testing.T) {
		data, err := readJSONSidecar(strings.NewReader("{}"), 2)
		require.NoError(t, err)
		assert.Equal(t, []byte("{}"), data)
	})
}

// TestJSONFileSizeLimit checks inclusive size limits before metadata dispatch.
func TestJSONFileSizeLimit(t *testing.T) {
	t.Setenv("PHOTOPRISM_JSON_LIMIT", "")
	require.Equal(t, int64(1<<20), JSONMaxFileBytes)
	const body = `[{"ExifToolVersion":12,"FileName":"control.jpg","ImageWidth":640,"ImageHeight":480}]`

	for _, tc := range []struct {
		name string
		size int64
	}{
		{"BelowLimit", JSONMaxFileBytes - 1},
		{"AtLimit", JSONMaxFileBytes},
		{"AboveLimit", JSONMaxFileBytes + 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			name := filepath.Join(t.TempDir(), "control.json")
			payload := bytes.Repeat([]byte{' '}, int(tc.size))
			copy(payload, body)
			require.NoError(t, os.WriteFile(name, payload, fs.ModeFile))

			data, err := JSON(name, "control.jpg")
			if tc.size > JSONMaxFileBytes {
				assert.ErrorIs(t, err, ErrJSONFileTooLarge)
				assert.Equal(t, Data{}, data)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, 640, data.Width)
			assert.Equal(t, 480, data.Height)
		})
	}
}

// TestDataJSONOversizedLink preserves accumulated metadata when a linked sidecar is too large.
func TestDataJSONOversizedLink(t *testing.T) {
	t.Setenv("PHOTOPRISM_JSON_LIMIT", "")
	dir := t.TempDir()
	name := filepath.Join(dir, "large.json")
	f, err := os.Create(name) //nolint:gosec // Test creates a sidecar in its temporary directory.
	require.NoError(t, err)
	require.NoError(t, f.Truncate(JSONMaxFileBytes+1))
	require.NoError(t, f.Close())
	alias := filepath.Join(dir, "alias.json")
	require.NoError(t, os.Symlink(name, alias))

	data := Data{Title: "existing-control", Width: 800}
	before := data
	assert.ErrorIs(t, data.JSON(alias, ""), ErrJSONFileTooLarge)
	assert.Equal(t, before, data)
}

// TestJSONFileLimit checks override validation and the reader's effective limit.
func TestJSONFileLimit(t *testing.T) {
	for _, tc := range []struct {
		name, value string
		want        int64
	}{
		{"Default", "", JSONMaxFileBytes},
		{"Override", "2097152", 2 << 20},
		{"Whitespace", " 2048 ", 2048},
		{"Zero", "0", JSONMaxFileBytes},
		{"Negative", "-1", JSONMaxFileBytes},
		{"Malformed", "abc", JSONMaxFileBytes},
		{"Overflow", "18446744073709551615", JSONMaxFileBytes},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("PHOTOPRISM_JSON_LIMIT", tc.value)
			require.Equal(t, tc.want, JSONFileLimit())
			r := &jsonLimitReader{remaining: tc.want}
			data, err := readJSONSidecar(r, 0)
			require.NoError(t, err)
			require.Len(t, data, int(tc.want))
			r = &jsonLimitReader{remaining: tc.want + 100}
			_, err = readJSONSidecar(r, 0)
			require.ErrorIs(t, err, ErrJSONFileTooLarge)
			require.Equal(t, tc.want+1, r.read)
		})
	}
}

package photoprism

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/pkg/fs"
)

func writeFile(t *testing.T, p string, data []byte) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(p), fs.ModeDir); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, data, fs.ModeFile); err != nil {
		t.Fatal(err)
	}
}

func readFile(t *testing.T, p string) []byte {
	t.Helper()
	// #nosec G304 -- test reads from controlled temp directory.
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func TestMediaFileCopy(t *testing.T) {
	testCases := []struct {
		name      string
		force     bool
		expectErr bool
		setup     func(t *testing.T, dir string) (src, dst string)
		before    func(t *testing.T, mf *MediaFile)
		destFn    func(src, dst string) string
		assertFn  func(t *testing.T, src, dst string)
	}{
		{
			name:      "existing destination without force",
			force:     false,
			expectErr: true,
			setup: func(t *testing.T, dir string) (string, string) {
				src := filepath.Join(dir, "src.bin")
				dst := filepath.Join(dir, "dst.bin")
				writeFile(t, src, []byte("ABC"))
				writeFile(t, dst, []byte("LONGER_DEST_CONTENT"))
				return src, dst
			},
			assertFn: func(t *testing.T, _, dst string) {
				assert.Equal(t, "LONGER_DEST_CONTENT", string(readFile(t, dst)))
			},
		},
		{
			name:  "existing empty destination without force",
			force: false,
			setup: func(t *testing.T, dir string) (string, string) {
				src := filepath.Join(dir, "src.bin")
				dst := filepath.Join(dir, "dst.bin")
				writeFile(t, src, []byte("ABC"))
				writeFile(t, dst, []byte{})
				return src, dst
			},
			assertFn: func(t *testing.T, _, dst string) {
				assert.Equal(t, "ABC", string(readFile(t, dst)))
			},
		},
		{
			name:  "force overwrites destination and propagates mtime",
			force: true,
			setup: func(t *testing.T, dir string) (string, string) {
				src := filepath.Join(dir, "src.bin")
				dst := filepath.Join(dir, "dst.bin")
				writeFile(t, src, []byte("ABC"))
				writeFile(t, dst, []byte("LONGER_DEST_CONTENT"))
				return src, dst
			},
			before: func(t *testing.T, mf *MediaFile) {
				reference := time.Date(2020, 5, 4, 3, 2, 1, 0, time.UTC)
				mf.SetModTime(reference)
			},
			assertFn: func(t *testing.T, _, dst string) {
				assert.Equal(t, "ABC", string(readFile(t, dst)))
				st, err := os.Stat(dst)
				require.NoError(t, err)
				expected := time.Date(2020, 5, 4, 3, 2, 1, 0, time.UTC)
				assert.Equal(t, expected, st.ModTime().UTC().Truncate(time.Second))
			},
		},
		{
			name:      "same path returns error",
			force:     true,
			expectErr: true,
			setup: func(t *testing.T, dir string) (string, string) {
				src := filepath.Join(dir, "file.bin")
				writeFile(t, src, []byte("DATA"))
				return src, src
			},
		},
		{
			name:      "invalid destination path",
			force:     true,
			expectErr: true,
			setup: func(t *testing.T, dir string) (string, string) {
				src := filepath.Join(dir, "file.bin")
				writeFile(t, src, []byte("DATA"))
				return src, filepath.Join(dir, "unused")
			},
			destFn: func(string, string) string { return "." },
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			src, dst := tc.setup(t, dir)
			mf, err := NewMediaFile(src)
			require.NoError(t, err)

			if tc.before != nil {
				tc.before(t, mf)
			}

			target := dst
			if tc.destFn != nil {
				target = tc.destFn(src, dst)
			}

			err = mf.Copy(target, tc.force)
			if tc.expectErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			if tc.assertFn != nil {
				tc.assertFn(t, src, target)
			}
		})
	}
}

func TestMediaFileMove(t *testing.T) {
	testCases := []struct {
		name      string
		force     bool
		expectErr bool
		setup     func(t *testing.T, dir string) (src, dst string)
		assertFn  func(t *testing.T, src, dst string, mf *MediaFile)
	}{
		{
			name:      "existing destination without force",
			force:     false,
			expectErr: true,
			setup: func(t *testing.T, dir string) (string, string) {
				src := filepath.Join(dir, "src.bin")
				dst := filepath.Join(dir, "dst.bin")
				writeFile(t, src, []byte("AAA"))
				writeFile(t, dst, []byte("BBB"))
				return src, dst
			},
			assertFn: func(t *testing.T, src, dst string, _ *MediaFile) {
				assert.FileExists(t, src)
				assert.Equal(t, "BBB", string(readFile(t, dst)))
			},
		},
		{
			name:  "existing empty destination without force",
			force: false,
			setup: func(t *testing.T, dir string) (string, string) {
				src := filepath.Join(dir, "src.bin")
				dst := filepath.Join(dir, "dst.bin")
				writeFile(t, src, []byte("AAA"))
				writeFile(t, dst, []byte{})
				return src, dst
			},
			assertFn: func(t *testing.T, src, dst string, mf *MediaFile) {
				_, srcErr := os.Stat(src)
				assert.True(t, os.IsNotExist(srcErr))
				assert.Equal(t, "AAA", string(readFile(t, dst)))
				assert.Equal(t, dst, mf.FileName())
			},
		},
		{
			name:  "force overwrites destination",
			force: true,
			setup: func(t *testing.T, dir string) (string, string) {
				src := filepath.Join(dir, "src.bin")
				dst := filepath.Join(dir, "dst.bin")
				writeFile(t, src, []byte("AAA"))
				writeFile(t, dst, []byte("BBB"))
				return src, dst
			},
			assertFn: func(t *testing.T, src, dst string, mf *MediaFile) {
				_, srcErr := os.Stat(src)
				assert.True(t, os.IsNotExist(srcErr))
				assert.Equal(t, "AAA", string(readFile(t, dst)))
				assert.Equal(t, dst, mf.FileName())
			},
		},
		{
			name:      "same path returns error",
			force:     true,
			expectErr: true,
			setup: func(t *testing.T, dir string) (string, string) {
				src := filepath.Join(dir, "file.bin")
				writeFile(t, src, []byte("DATA"))
				return src, src
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			src, dst := tc.setup(t, dir)
			mf, err := NewMediaFile(src)
			require.NoError(t, err)

			err = mf.Move(dst, tc.force)
			if tc.expectErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			if tc.assertFn != nil {
				tc.assertFn(t, src, dst, mf)
			}
		})
	}
}

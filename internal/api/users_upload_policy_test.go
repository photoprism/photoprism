package api

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/pkg/fs"
)

// TestUploadSidecarAllowed checks the built-in policy independently of operator extension settings.
func TestUploadSidecarAllowed(t *testing.T) {
	for _, name := range []string{"a.jpg", "a.dng", "a.png", "a.mp4", "a.pdf", "a.zip", "a.XMP", "a.TXT", "a.md", "a.markdown", "LRV_20240415_213145_01_035.lrv"} {
		assert.True(t, uploadSidecarAllowed(name), name)
	}

	for _, name := range fs.ReservedPathNames() {
		assert.False(t, uploadSidecarAllowed("nested/"+name+"/photo.jpg"), name)
	}

	for _, name := range []string{"a.yml", "a.YAML", "a.JSON", "a.aae", "a.xml", "a.nfo", "a.unknown", "GL010123.LRV", "a.lrv", "a.rclonelink", "nested/link.RCLONELINK/photo.jpg", ".ppignore", ".env.jpg", ".ENV.example.txt", ".git/photo.jpg"} {
		assert.False(t, uploadSidecarAllowed(name), name)
	}
}

// TestPruneUploadSidecars checks staged metadata removal and retained file contents.
func TestPruneUploadSidecars(t *testing.T) {
	t.Run("NestedBatch", func(t *testing.T) {
		dir := t.TempDir()

		for _, name := range []string{"a.jpg", "a.xmp", "a.md", "a.txt", "a.yml", "a.json", "nested/.photoprism/a.YAML", "nested/a.xml"} {
			filename := filepath.Join(dir, name)
			require.NoError(t, fs.MkdirAll(filepath.Dir(filename)))
			require.NoError(t, os.WriteFile(filename, []byte(name), fs.ModeFile))
		}

		require.NoError(t, pruneUploadSidecars(dir))

		for _, name := range []string{"a.jpg", "a.xmp", "a.md", "a.txt"} {
			data, err := os.ReadFile(filepath.Join(dir, name)) //nolint:gosec // Test reads its temporary control file.
			require.NoError(t, err)
			assert.Equal(t, name, string(data))
		}

		for _, name := range []string{"a.yml", "a.json", "nested/.photoprism/a.YAML", "nested/a.xml"} {
			assert.NoFileExists(t, filepath.Join(dir, name))
		}
	})
	t.Run("NoFollowingLinks", func(t *testing.T) {
		dir, outside := t.TempDir(), t.TempDir()
		file := filepath.Join(outside, "a.yml")
		require.NoError(t, os.WriteFile(file, []byte("outside-control"), fs.ModeFile))
		require.NoError(t, os.Symlink(outside, filepath.Join(dir, "linked")))
		require.Error(t, pruneUploadSidecars(dir))
		data, err := os.ReadFile(file) //nolint:gosec // Test reads its temporary control file.
		require.NoError(t, err)
		assert.Equal(t, "outside-control", string(data))
	})

	for _, name := range []string{"FileLink", "BrokenLink", "RootLink"} {
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			outside := filepath.Join(t.TempDir(), "outside.jpg")
			require.NoError(t, os.WriteFile(outside, []byte("outside-control"), fs.ModeFile))
			target := outside
			link := filepath.Join(dir, "allowed.jpg")
			root := dir

			if name == "BrokenLink" {
				target = filepath.Join(t.TempDir(), "missing")
			}

			if name == "RootLink" {
				target = t.TempDir()
				root = link
			}

			require.NoError(t, os.Symlink(target, link))
			assert.Error(t, pruneUploadSidecars(root))
			data, err := os.ReadFile(outside) //nolint:gosec // Test reads a controlled fixture or generated output.
			require.NoError(t, err)
			assert.Equal(t, "outside-control", string(data))
		})
	}
	t.Run("ReservedSubtree", func(t *testing.T) {
		dir := filepath.Join(t.TempDir(), ".ssh", "batch")
		for _, name := range []string{"keep.jpg", ".git/photo.jpg", ".hidden/keep.txt"} {
			file := filepath.Join(dir, name)
			require.NoError(t, fs.MkdirAll(filepath.Dir(file)))
			require.NoError(t, os.WriteFile(file, []byte("control"), fs.ModeFile))
		}
		require.NoError(t, pruneUploadSidecars(dir))
		assert.NoDirExists(t, filepath.Join(dir, ".git"))
		assert.FileExists(t, filepath.Join(dir, "keep.jpg"))
		assert.FileExists(t, filepath.Join(dir, ".hidden/keep.txt"))
	})
	t.Run("MissingRoot", func(t *testing.T) { assert.Error(t, pruneUploadSidecars(filepath.Join(t.TempDir(), "missing"))) })
}

// TestUploadCheckFileSidecars checks validation at the shared saved-file boundary.
func TestUploadCheckFileSidecars(t *testing.T) {
	for _, name := range []string{"a.yml", "a.yaml", "a.JSON", "a.xml", "a.aae", "a.nfo"} {
		t.Run(map[string]string{"a.yml": "YAMLShortExtension", "a.yaml": "YAMLLongExtension", "a.JSON": "JSONUppercase", "a.xml": "XML", "a.aae": "AppleXML", "a.nfo": "Info", "a.txt": "Text", "a.md": "Markdown", "a.xmp": "XMP"}[name], func(t *testing.T) {
			filename := filepath.Join(t.TempDir(), name)
			require.NoError(t, os.WriteFile(filename, []byte("sidecar-control"), fs.ModeFile))
			remaining, err := UploadCheckFile(filename, false, 1024)
			assert.Error(t, err)
			assert.Equal(t, int64(1024), remaining)
			assert.NoFileExists(t, filename)
		})
	}

	for _, name := range []string{"a.txt", "a.md", "a.xmp"} {
		t.Run(map[string]string{"a.yml": "YAMLShortExtension", "a.yaml": "YAMLLongExtension", "a.JSON": "JSONUppercase", "a.xml": "XML", "a.aae": "AppleXML", "a.nfo": "Info", "a.txt": "Text", "a.md": "Markdown", "a.xmp": "XMP"}[name], func(t *testing.T) {
			filename := filepath.Join(t.TempDir(), name)
			data := []byte("<metadata>control</metadata>")
			require.NoError(t, os.WriteFile(filename, data, fs.ModeFile))
			remaining, err := UploadCheckFile(filename, false, 1024)
			assert.NoError(t, err)
			assert.Equal(t, int64(1024-len(data)), remaining)
			assert.FileExists(t, filename)
		})
	}
}

// TestUploadArchiveEntryAllowed checks file and directory policy independently of ZIP layout.
func TestUploadArchiveEntryAllowed(t *testing.T) {
	assert.False(t, uploadArchiveEntryAllowed(".ssh", true))
	assert.False(t, uploadArchiveEntryAllowed(".ssh/photo.jpg", false))
	assert.False(t, uploadArchiveEntryAllowed("photo.json", false))
	assert.False(t, uploadArchiveEntryAllowed("photo.rclonelink", true))
	assert.False(t, uploadArchiveEntryAllowed("nested/photo.RCLONELINK/photo.jpg", false))
	assert.True(t, uploadArchiveEntryAllowed("nested/photo.rclonelink.jpg", false))
	assert.True(t, uploadArchiveEntryAllowed(".ordinary", true))
	assert.True(t, uploadArchiveEntryAllowed("ordinary/photo.md", false))
}

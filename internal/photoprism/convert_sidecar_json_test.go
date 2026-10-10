package photoprism

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/pkg/fs"
)

func TestConvert_ToJson(t *testing.T) {
	c := config.TestConfig()
	convert := NewConvert(c)

	t.Run("InvalidCache", func(t *testing.T) {
		if !c.ExifToolEnabled() {
			t.Skip("ExifTool must be available")
		}

		mf, err := NewMediaFile(filepath.Join(c.SamplesPath(), "beach_sand.jpg"))
		require.NoError(t, err)
		jsonName, err := mf.ExifToolJsonName()
		require.NoError(t, err)
		t.Cleanup(func() { _ = os.Remove(jsonName) })

		for _, data := range []string{"", "[{\"SourceFile\":"} {
			require.NoError(t, fs.MkdirAll(filepath.Dir(jsonName)))
			require.NoError(t, os.WriteFile(jsonName, []byte(data), fs.ModeFile))
			result, toJsonErr := convert.ToJson(mf, false)
			require.NoError(t, toJsonErr)
			assert.Equal(t, jsonName, result)
			assert.True(t, exifToolCacheValid(jsonName), "an empty or truncated cache is replaced")
		}
	})
	t.Run("UnexpectedOutput", func(t *testing.T) {
		if !c.ExifToolEnabled() {
			t.Skip("ExifTool must be enabled")
		}

		mf, err := NewMediaFile(filepath.Join(c.SamplesPath(), "beach_sand.jpg"))
		require.NoError(t, err)
		jsonName, err := mf.ExifToolJsonName()
		require.NoError(t, err)
		_ = os.Remove(jsonName)
		t.Cleanup(func() { _ = os.Remove(jsonName) })

		useExifToolStub(t, c, "echo 'Warning: no data'\n")
		_, err = convert.ToJson(mf, false)
		assert.Error(t, err)
		assert.False(t, fs.FileExists(jsonName), "unexpected output is not cached")
	})
	t.Run("ValidCache", func(t *testing.T) {
		if !c.ExifToolEnabled() {
			t.Skip("ExifTool must be enabled")
		}

		mf, err := NewMediaFile(filepath.Join(c.SamplesPath(), "beach_sand.jpg"))
		require.NoError(t, err)
		jsonName, err := mf.ExifToolJsonName()
		require.NoError(t, err)
		t.Cleanup(func() { _ = os.Remove(jsonName) })

		require.NoError(t, fs.MkdirAll(filepath.Dir(jsonName)))
		require.NoError(t, os.WriteFile(jsonName, []byte("[{\"Cached\":true}]\n"), fs.ModeFile))
		result, err := convert.ToJson(mf, false)
		require.NoError(t, err)
		assert.Equal(t, jsonName, result)
		data, err := os.ReadFile(jsonName) //nolint:gosec // test-owned cache file
		require.NoError(t, err)
		assert.Equal(t, "[{\"Cached\":true}]\n", string(data), "a valid cache is kept")
	})
	t.Run("GopherVideoMp4", func(t *testing.T) {
		fileName := filepath.Join(c.SamplesPath(), "gopher-video.mp4")

		assert.Truef(t, fs.FileExists(fileName), "input file does not exist: %s", fileName)

		mf, err := NewMediaFile(fileName)

		if err != nil {
			t.Fatal(err)
		}

		jsonName, err := convert.ToJson(mf, false)

		if err != nil {
			t.Fatal(err)
		}

		if jsonName == "" {
			t.Fatal("json file name should not be empty")
		}

		assert.FileExists(t, jsonName)

		_ = os.Remove(jsonName)
	})
	t.Run("ImgNum4120Jpg", func(t *testing.T) {
		fileName := filepath.Join(c.SamplesPath(), "IMG_4120.JPG")
		assert.Truef(t, fs.FileExists(fileName), "input file does not exist: %s", fileName)

		mf, err := NewMediaFile(fileName)

		if err != nil {
			t.Fatal(err)
		}

		jsonName, err := convert.ToJson(mf, false)

		if err != nil {
			t.Fatal(err)
		}

		if jsonName == "" {
			t.Fatal("json file name should not be empty")
		}

		assert.FileExists(t, jsonName)

		_ = os.Remove(jsonName)
	})
	t.Run("IphoneSevenHeic", func(t *testing.T) {
		fileName := c.SamplesPath() + "/iphone_7.heic"

		assert.True(t, fs.FileExists(fileName))

		mf, err := NewMediaFile(fileName)

		if err != nil {
			t.Fatal(err)
		}

		jsonName, err := convert.ToJson(mf, false)

		if err != nil {
			t.Fatal(err)
		}

		if jsonName == "" {
			t.Fatal("json file name should not be empty")
		}

		assert.FileExists(t, jsonName)

		_ = os.Remove(jsonName)
	})
	t.Run("IphoneFifteenProHeic", func(t *testing.T) {
		fileName := c.SamplesPath() + "/iphone_15_pro.heic"

		assert.True(t, fs.FileExists(fileName))

		mf, err := NewMediaFile(fileName)

		if err != nil {
			t.Fatal(err)
		}

		jsonName, err := convert.ToJson(mf, false)

		if err != nil {
			t.Fatal(err)
		}

		if jsonName == "" {
			t.Fatal("json file name should not be empty")
		}

		assert.FileExists(t, jsonName)

		_ = os.Remove(jsonName)
	})
}

func TestExifToolCacheValid(t *testing.T) {
	dir := t.TempDir()

	// valid writes the data to a new file and reports whether it is a valid cache.
	valid := func(t *testing.T, data string) bool {
		t.Helper()
		fileName := filepath.Join(dir, "cache.json")
		require.NoError(t, os.WriteFile(fileName, []byte(data), fs.ModeFile))
		return exifToolCacheValid(fileName)
	}

	t.Run("Valid", func(t *testing.T) {
		assert.True(t, valid(t, "[{\"SourceFile\":\"a.jpg\"}]\n"))
		assert.True(t, valid(t, "\n  [\n  {}\n]  \n"))
		assert.True(t, valid(t, "[]"))
		assert.True(t, valid(t, "[{\"SourceFile\":\""+strings.Repeat("x", 100)+"\"}]\n"))
	})
	t.Run("Invalid", func(t *testing.T) {
		assert.False(t, valid(t, ""))
		assert.False(t, valid(t, "["))
		assert.False(t, valid(t, "[{\"SourceFile\":\"a.jpg\""))
		assert.False(t, valid(t, "{\"SourceFile\":\"a.jpg\"}"))
		assert.False(t, valid(t, "{\"SourceFile\":\"a.jpg\"}]"))
		assert.False(t, valid(t, "   \n  "))
		assert.False(t, valid(t, "[{\"List\":[1]}, {\"SourceFile\":\""+strings.Repeat("x", 100)))
	})
	t.Run("NotFile", func(t *testing.T) {
		assert.False(t, exifToolCacheValid(filepath.Join(dir, "missing.json")))
		assert.False(t, exifToolCacheValid(dir))
	})
}

func TestWriteExifToolCache(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		dir := filepath.Join(t.TempDir(), "a", "b")
		fileName := filepath.Join(dir, "cache.json")
		require.NoError(t, writeExifToolCache(fileName, []byte("[{}]\n")))
		data, err := os.ReadFile(fileName) //nolint:gosec // test-owned cache file
		require.NoError(t, err)
		assert.Equal(t, "[{}]\n", string(data))
		entries, err := os.ReadDir(dir)
		require.NoError(t, err)
		assert.Len(t, entries, 1, "no staged file is left behind")
	})
	t.Run("Replace", func(t *testing.T) {
		dir := t.TempDir()
		fileName := filepath.Join(dir, "cache.json")
		require.NoError(t, os.WriteFile(fileName, []byte("[{"), fs.ModeFile))
		require.NoError(t, writeExifToolCache(fileName, []byte("[{}]\n")))
		data, err := os.ReadFile(fileName) //nolint:gosec // test-owned cache file
		require.NoError(t, err)
		assert.Equal(t, "[{}]\n", string(data))
		entries, err := os.ReadDir(dir)
		require.NoError(t, err)
		assert.Len(t, entries, 1)
	})
	t.Run("PublishError", func(t *testing.T) {
		// A failed publish leaves no staged file behind.
		dir := t.TempDir()
		fileName := filepath.Join(dir, "cache.json")
		require.NoError(t, os.MkdirAll(filepath.Join(fileName, "sub"), fs.ModeDir))
		assert.Error(t, writeExifToolCache(fileName, []byte("[]")))
		entries, err := os.ReadDir(dir)
		require.NoError(t, err)
		assert.Len(t, entries, 1)
	})
	t.Run("Error", func(t *testing.T) {
		parent := filepath.Join(t.TempDir(), "file")
		require.NoError(t, os.WriteFile(parent, []byte("x"), fs.ModeFile))
		assert.Error(t, writeExifToolCache(filepath.Join(parent, "cache.json"), []byte("[]")))
	})
}

func TestJsonArrayBounds(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		assert.True(t, jsonArrayBounds([]byte("[{}]"), []byte("[{}]")))
		assert.True(t, jsonArrayBounds([]byte(" \n[{"), []byte("}]\n ")))
	})
	t.Run("Invalid", func(t *testing.T) {
		assert.False(t, jsonArrayBounds(nil, nil))
		assert.False(t, jsonArrayBounds([]byte("{}"), []byte("{}")))
		assert.False(t, jsonArrayBounds([]byte("[{"), []byte("}")))
		assert.False(t, jsonArrayBounds([]byte("Warning"), []byte("]")))
	})
}

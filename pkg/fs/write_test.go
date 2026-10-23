package fs

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriteFile(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		dir := "./testdata/_WriteFile_Success"
		filePath := filepath.Join(dir, "notyetexisting.jpg")
		fileData := []byte("foobar")

		if err := MkdirAll(dir); err != nil {
			t.Fatal(err)
		}

		defer func() {
			_ = os.Remove(filePath)

			if err := os.RemoveAll(dir); err != nil {
				t.Fatal(err)
			}
		}()

		assert.True(t, PathExists(dir))

		fileErr := WriteFile(filePath, fileData, ModeFile)

		assert.NoError(t, fileErr)
		assert.FileExists(t, filePath)
	})
}

func TestWriteString(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		dir := "./testdata/_WriteString_Success"
		filePath := filepath.Join(dir, PPIgnoreFilename)
		fileData := "*"

		if err := MkdirAll(dir); err != nil {
			t.Fatal(err)
		}

		defer func() {
			_ = os.Remove(filePath)

			if err := os.RemoveAll(dir); err != nil {
				t.Fatal(err)
			}
		}()

		assert.True(t, PathExists(dir))

		fileErr := WriteString(filePath, fileData)

		assert.NoError(t, fileErr)
		assert.FileExists(t, filePath)

		readLines, readErr := ReadLines(filePath)

		assert.NoError(t, readErr)
		assert.Len(t, readLines, 1)
		assert.Equal(t, fileData, readLines[0])
	})
}

func TestWriteUnixTime(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		dir := "./testdata/_WriteUnixTime_Success"
		filePath := filepath.Join(dir, PPStorageFilename)

		if err := MkdirAll(dir); err != nil {
			t.Fatal(err)
		}

		defer func() {
			_ = os.Remove(filePath)

			if err := os.RemoveAll(dir); err != nil {
				t.Fatal(err)
			}
		}()

		assert.True(t, PathExists(dir))

		unixTime, fileErr := WriteUnixTime(filePath)

		assert.NoError(t, fileErr)
		assert.FileExists(t, filePath)

		readLines, readErr := ReadLines(filePath)

		assert.NoError(t, readErr)
		assert.Len(t, readLines, 1)
		assert.Equal(t, strconv.FormatInt(unixTime, 10), readLines[0])
	})
}

func TestWriteFileFromReader(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		dir := "./testdata/_WriteFileFromReader_Success"

		filePath1 := filepath.Join(dir, "1.txt")
		filePath2 := filepath.Join(dir, "2.txt")

		if err := MkdirAll(dir); err != nil {
			t.Fatal(err)
		}

		defer func() {
			_ = os.Remove(filePath1)
			_ = os.Remove(filePath2)

			if err := os.RemoveAll(dir); err != nil {
				t.Fatal(err)
			}
		}()

		assert.True(t, PathExists(dir))

		unixTime, writeErr := WriteUnixTime(filePath1)
		assert.NoError(t, writeErr)
		assert.True(t, unixTime >= time.Now().Unix())

		fileReader, readerErr := os.Open(filePath1) //nolint:gosec // test helper reads temp file
		assert.NoError(t, readerErr)

		fileErr := WriteFileFromReader(filePath2, fileReader)

		assert.NoError(t, fileErr)

		readLines, readErr := ReadLines(filePath2)

		assert.NoError(t, readErr)
		assert.Len(t, readLines, 1)
		assert.Equal(t, strconv.FormatInt(unixTime, 10), readLines[0])
	})
}

func TestCacheFileFromReader(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		dir := "./testdata/_CacheFileFromReader_Success"

		filePath1 := filepath.Join(dir, "1.txt")
		filePath2 := filepath.Join(dir, "2.txt")
		filePath3 := filepath.Join(dir, "3.txt")

		if err := MkdirAll(dir); err != nil {
			t.Fatal(err)
		}

		defer func() {
			_ = os.Remove(filePath1)
			_ = os.Remove(filePath2)
			_ = os.Remove(filePath3)

			if err := os.RemoveAll(dir); err != nil {
				t.Fatal(err)
			}
		}()

		assert.True(t, PathExists(dir))

		unixTime, writeErr := WriteUnixTime(filePath1)
		assert.NoError(t, writeErr)
		assert.True(t, unixTime >= time.Now().Unix())

		fileReader, readerErr := os.Open(filePath1) //nolint:gosec // test helper reads temp file
		assert.NoError(t, readerErr)

		cacheFile, cacheErr := CacheFileFromReader(filePath2, fileReader)

		assert.NoError(t, cacheErr)
		assert.Equal(t, filePath2, cacheFile)

		readLines, readErr := ReadLines(cacheFile)

		assert.NoError(t, readErr)
		assert.Len(t, readLines, 1)
		assert.Equal(t, strconv.FormatInt(unixTime, 10), readLines[0])

		if err := WriteString(filePath3, "0"); err != nil {
			t.Fatal(err)
		}

		cacheFile, cacheErr = CacheFileFromReader(filePath3, fileReader)

		assert.NoError(t, cacheErr)
		assert.Equal(t, filePath3, cacheFile)

		readLines, readErr = ReadLines(cacheFile)

		assert.NoError(t, readErr)
		assert.Len(t, readLines, 1)
		assert.Equal(t, "0", readLines[0])
	})
}

func TestWriteFile_Truncates(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "f.txt")
	assert.NoError(t, os.WriteFile(p, []byte("LONGDATA"), ModeFile))
	assert.NoError(t, WriteFile(p, []byte("short"), ModeFile))
	b, err := os.ReadFile(p) //nolint:gosec // test helper reads temp file
	assert.NoError(t, err)
	assert.Equal(t, "short", string(b))
}

func TestWriteFile_Errors(t *testing.T) {
	err := WriteFile("", []byte("x"), ModeFile)
	assert.Error(t, err)
}

func TestWriteFileFromReader_Errors(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "x.txt")

	// nil reader
	assert.Error(t, WriteFileFromReader(p, nil))
	// empty filename
	assert.Error(t, WriteFileFromReader("", bytes.NewBufferString("hi")))
}

// writeOnRead is a reader that creates a file before it reports the end of its data.
type writeOnRead struct {
	fileName string
	data     string
}

// Read writes the file and reports the end of the data.
func (r writeOnRead) Read([]byte) (int, error) {
	if err := os.WriteFile(r.fileName, []byte(r.data), ModeFile); err != nil { //nolint:gosec // test writes a temp file
		return 0, err
	}

	return 0, io.EOF
}

// failingReader returns some data and then an error.
type failingReader struct{ read bool }

// Read returns one byte and then an error.
func (r *failingReader) Read(p []byte) (int, error) {
	if r.read {
		return 0, errors.New("read failed")
	}

	r.read = true
	p[0] = 'x'

	return 1, nil
}

func TestCacheFileFromReader_Staged(t *testing.T) {
	// staged returns the staged siblings of the file that are left in its folder.
	staged := func(t *testing.T, fileName string) []string {
		matches, err := filepath.Glob(filepath.Join(filepath.Dir(fileName), "."+filepath.Base(fileName)+".*"))
		require.NoError(t, err)
		return matches
	}
	t.Run("FailingReader", func(t *testing.T) {
		fileName := filepath.Join(t.TempDir(), "video.mp4")
		_, err := CacheFileFromReader(fileName, &failingReader{})
		assert.ErrorContains(t, err, "read failed")
		assert.NoFileExists(t, fileName)
		assert.Empty(t, staged(t, fileName))
	})
	t.Run("ConcurrentPublish", func(t *testing.T) {
		fileName := filepath.Join(t.TempDir(), "video.mp4")
		result, err := CacheFileFromReader(fileName, writeOnRead{fileName: fileName, data: "other"})
		require.NoError(t, err)
		assert.Equal(t, fileName, result)
		data, err := os.ReadFile(fileName) //nolint:gosec // test reads a temp file
		require.NoError(t, err)
		assert.Equal(t, "other", string(data))
		assert.Empty(t, staged(t, fileName))
	})
	t.Run("ReplacesEmpty", func(t *testing.T) {
		fileName := filepath.Join(t.TempDir(), "video.mp4")
		require.NoError(t, os.WriteFile(fileName, nil, ModeFile)) //nolint:gosec // test writes a temp file
		_, err := CacheFileFromReader(fileName, strings.NewReader("data"))
		require.NoError(t, err)
		data, err := os.ReadFile(fileName) //nolint:gosec // test reads a temp file
		require.NoError(t, err)
		assert.Equal(t, "data", string(data))
		assert.Empty(t, staged(t, fileName))
	})
	t.Run("DirectoryAtName", func(t *testing.T) {
		fileName := filepath.Join(t.TempDir(), "video.mp4")
		require.NoError(t, os.Mkdir(fileName, ModeDir))
		_, err := CacheFileFromReader(fileName, strings.NewReader("data"))
		assert.Error(t, err)
		assert.Empty(t, staged(t, fileName))
	})
	t.Run("DanglingLink", func(t *testing.T) {
		dir := t.TempDir()
		fileName := filepath.Join(dir, "video.mp4")
		target := filepath.Join(dir, "target.mp4")
		require.NoError(t, os.Symlink(target, fileName))
		_, err := CacheFileFromReader(fileName, strings.NewReader("data"))
		assert.Error(t, err)
		assert.NoFileExists(t, target)
		assert.True(t, IsSymlink(fileName))
		assert.Empty(t, staged(t, fileName))
	})
	t.Run("NoReader", func(t *testing.T) {
		_, err := CacheFileFromReader(filepath.Join(t.TempDir(), "video.mp4"), nil)
		assert.Error(t, err)
	})
}

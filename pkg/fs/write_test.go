package fs

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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

		fileErr := WriteFile(filePath, fileData)

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

		fileReader, readerErr := os.Open(filePath1)
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

		fileReader, readerErr := os.Open(filePath1)
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

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
		pathName := "./testdata/_WriteFile_Success"
		fileName := filepath.Join(pathName, "notyetexisting.jpg")
		fileData := []byte("foobar")

		if err := MkdirAll(pathName); err != nil {
			t.Fatal(err)
		}

		defer func() {
			_ = os.Remove(fileName)

			if err := os.RemoveAll(pathName); err != nil {
				t.Fatal(err)
			}
		}()

		assert.True(t, PathExists(pathName))

		fileErr := WriteFile(fileName, fileData)

		assert.NoError(t, fileErr)
		assert.FileExists(t, fileName)
	})
}

func TestWriteString(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		pathName := "./testdata/_WriteString_Success"
		fileName := filepath.Join(pathName, PPIgnoreFilename)
		fileData := "*"

		if err := MkdirAll(pathName); err != nil {
			t.Fatal(err)
		}

		defer func() {
			_ = os.Remove(fileName)

			if err := os.RemoveAll(pathName); err != nil {
				t.Fatal(err)
			}
		}()

		assert.True(t, PathExists(pathName))

		fileErr := WriteString(fileName, fileData)

		assert.NoError(t, fileErr)
		assert.FileExists(t, fileName)

		readLines, readErr := ReadLines(fileName)

		assert.NoError(t, readErr)
		assert.Len(t, readLines, 1)
		assert.Equal(t, fileData, readLines[0])
	})
}

func TestWriteUnixTime(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		pathName := "./testdata/_WriteUnixTime_Success"
		fileName := filepath.Join(pathName, PPStorageFilename)

		if err := MkdirAll(pathName); err != nil {
			t.Fatal(err)
		}

		defer func() {
			_ = os.Remove(fileName)

			if err := os.RemoveAll(pathName); err != nil {
				t.Fatal(err)
			}
		}()

		assert.True(t, PathExists(pathName))

		unixTime, fileErr := WriteUnixTime(fileName)

		assert.NoError(t, fileErr)
		assert.FileExists(t, fileName)

		readLines, readErr := ReadLines(fileName)

		assert.NoError(t, readErr)
		assert.Len(t, readLines, 1)
		assert.Equal(t, strconv.FormatInt(unixTime, 10), readLines[0])
	})
}

func TestWriteFileFromReader(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		pathName := "./testdata/_WriteFileFromReader_Success"

		fileName1 := filepath.Join(pathName, "1.txt")
		fileName2 := filepath.Join(pathName, "2.txt")

		if err := MkdirAll(pathName); err != nil {
			t.Fatal(err)
		}

		defer func() {
			_ = os.Remove(fileName1)
			_ = os.Remove(fileName2)

			if err := os.RemoveAll(pathName); err != nil {
				t.Fatal(err)
			}
		}()

		assert.True(t, PathExists(pathName))

		unixTime, writeErr := WriteUnixTime(fileName1)
		assert.NoError(t, writeErr)
		assert.True(t, unixTime >= time.Now().Unix())

		fileReader, readerErr := os.Open(fileName1)
		assert.NoError(t, readerErr)

		fileErr := WriteFileFromReader(fileName2, fileReader)

		assert.NoError(t, fileErr)

		readLines, readErr := ReadLines(fileName2)

		assert.NoError(t, readErr)
		assert.Len(t, readLines, 1)
		assert.Equal(t, strconv.FormatInt(unixTime, 10), readLines[0])
	})
}

func TestCacheFileFromReader(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		pathName := "./testdata/_CacheFileFromReader_Success"

		fileName1 := filepath.Join(pathName, "1.txt")
		fileName2 := filepath.Join(pathName, "2.txt")
		fileName3 := filepath.Join(pathName, "3.txt")

		if err := MkdirAll(pathName); err != nil {
			t.Fatal(err)
		}

		defer func() {
			_ = os.Remove(fileName1)
			_ = os.Remove(fileName2)
			_ = os.Remove(fileName3)

			if err := os.RemoveAll(pathName); err != nil {
				t.Fatal(err)
			}
		}()

		assert.True(t, PathExists(pathName))

		unixTime, writeErr := WriteUnixTime(fileName1)
		assert.NoError(t, writeErr)
		assert.True(t, unixTime >= time.Now().Unix())

		fileReader, readerErr := os.Open(fileName1)
		assert.NoError(t, readerErr)

		cacheFile, cacheErr := CacheFileFromReader(fileName2, fileReader)

		assert.NoError(t, cacheErr)
		assert.Equal(t, fileName2, cacheFile)

		readLines, readErr := ReadLines(cacheFile)

		assert.NoError(t, readErr)
		assert.Len(t, readLines, 1)
		assert.Equal(t, strconv.FormatInt(unixTime, 10), readLines[0])

		if err := WriteString(fileName3, "0"); err != nil {
			t.Fatal(err)
		}

		cacheFile, cacheErr = CacheFileFromReader(fileName3, fileReader)

		assert.NoError(t, cacheErr)
		assert.Equal(t, fileName3, cacheFile)

		readLines, readErr = ReadLines(cacheFile)

		assert.NoError(t, readErr)
		assert.Len(t, readLines, 1)
		assert.Equal(t, "0", readLines[0])
	})
}

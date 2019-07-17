package util

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExists(t *testing.T) {
	assert.True(t, Exists("./testdata/test.jpg"))
	assert.False(t, Exists("./foo.jpg"))
}

func TestOverwrite(t *testing.T) {
	data := make([]byte, 3)
	data[1] = 3
	data[2] = 8
	tmpPath := "./testdata/_tmp"
	os.Mkdir(tmpPath, 0777)

	defer os.RemoveAll(tmpPath)
	result := Overwrite("./testdata/_tmp/notyetexisting.jpg", data)
	assert.FileExists(t, "./testdata/_tmp/notyetexisting.jpg")
	t.Log(result)

}

func TestExpandedFilename(t *testing.T) {
	t.Run("test.jpg", func(t *testing.T) {
		filename := ExpandedFilename("./testdata/test.jpg")
		assert.Contains(t, filename, "/testdata/test.jpg")
		assert.IsType(t, "", filename)
	})
	t.Run("empty filename", func(t *testing.T) {
		filename := ExpandedFilename("")
		assert.Equal(t, "", filename)
		assert.IsType(t, "", filename)
	})
	t.Run("~ in filename", func(t *testing.T) {
		filename := ExpandedFilename("~/test.jpg")
		assert.Equal(t, "/root/test.jpg", filename)
		assert.IsType(t, "", filename)
	})
}

func TestDirectoryIsEmpty(t *testing.T) {
	t.Run("not empty path", func(t *testing.T) {
		assert.Equal(t, false, DirectoryIsEmpty("./testdata"))
	})
	t.Run("not existing path", func(t *testing.T) {
		assert.Equal(t, false, DirectoryIsEmpty("./xxx"))
	})
	t.Run("empty path", func(t *testing.T) {
		os.Mkdir("./testdata/emptyDir", 0777)
		defer os.RemoveAll("./testdata/emptyDir")
		assert.Equal(t, true, DirectoryIsEmpty("./testdata/emptyDir"))
	})
}

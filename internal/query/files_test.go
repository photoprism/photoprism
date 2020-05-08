package query

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExistingFiles(t *testing.T) {
	t.Run("files found", func(t *testing.T) {
		files, err := ExistingFiles(1000, 0, "/")

		t.Logf("files: %+v", files)

		assert.Nil(t, err)
		assert.LessOrEqual(t, 5, len(files))
	})
}

func TestFilesByUUID(t *testing.T) {
	t.Run("files found", func(t *testing.T) {
		files, err := FilesByUUID([]string{"ft8es39w45bnlqdw"}, 100, 0)

		assert.Nil(t, err)
		assert.Equal(t, 1, len(files))
		assert.Equal(t, "exampleFileName.jpg", files[0].FileName)
	})
}

func TestFileByPhotoUUID(t *testing.T) {
	t.Run("files found", func(t *testing.T) {
		file, err := FileByPhotoUUID("pt9jtdre2lvl0yh8")

		assert.Nil(t, err)
		assert.Equal(t, "exampleDNGFile.dng", file.FileName)
	})

	t.Run("no files found", func(t *testing.T) {
		file, err := FileByPhotoUUID("111")

		assert.Error(t, err, "record not found")
		t.Log(file)
	})
}

func TestFileByUUID(t *testing.T) {
	t.Run("files found", func(t *testing.T) {
		file, err := FileByUUID("ft8es39w45bnlqdw")

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "exampleFileName.jpg", file.FileName)
	})

	t.Run("no files found", func(t *testing.T) {
		file, err := FileByUUID("111")

		if err == nil {
			t.Fatal("error expected")
		}

		assert.Error(t, err, "record not found")
		t.Log(file)
	})
}

func TestFileByHash(t *testing.T) {
	t.Run("files found", func(t *testing.T) {
		file, err := FileByHash("2cad9168fa6acc5c5c2965ddf6ec465ca42fd818")

		assert.Nil(t, err)
		assert.Equal(t, "exampleFileName.jpg", file.FileName)
	})

	t.Run("no files found", func(t *testing.T) {
		file, err := FileByHash("111")

		assert.Error(t, err, "record not found")
		t.Log(file)
	})
}

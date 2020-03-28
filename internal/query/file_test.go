package query

import (
	"github.com/stretchr/testify/assert"
	"testing"

	"github.com/photoprism/photoprism/internal/config"
)

func TestRepo_FindFiles(t *testing.T) {
	conf := config.TestConfig()

	search := New(conf.Db())

	t.Run("files found", func(t *testing.T) {
		files, err := search.Files(1000, 0)

		assert.Nil(t, err)
		assert.Equal(t, 5, len(files))
	})
}

func TestRepo_FindFilesByUUID(t *testing.T) {
	conf := config.TestConfig()

	search := New(conf.Db())

	t.Run("files found", func(t *testing.T) {
		files, err := search.FilesByUUID([]string{"654"}, 100, 0)

		assert.Nil(t, err)
		assert.Equal(t, 1, len(files))
		assert.Equal(t, "exampleFileName.jpg", files[0].FileName)
	})
}

func TestRepo_FindFileByPhotoUUID(t *testing.T) {
	conf := config.TestConfig()

	search := New(conf.Db())

	t.Run("files found", func(t *testing.T) {
		file, err := search.FileByPhotoUUID("655")

		assert.Nil(t, err)
		assert.Equal(t, "exampleDNGFile.dng", file.FileName)
	})

	t.Run("no files found", func(t *testing.T) {
		file, err := search.FileByPhotoUUID("111")

		assert.Error(t, err, "record not found")
		t.Log(file)
	})
}

func TestRepo_FindFileByID(t *testing.T) {
	conf := config.TestConfig()

	search := New(conf.Db())

	t.Run("files found", func(t *testing.T) {
		file, err := search.FileByID("3")

		assert.Nil(t, err)
		assert.Equal(t, "exampleXmpFile.xmp", file.FileName)
	})

	t.Run("no files found", func(t *testing.T) {
		file, err := search.FileByID("111")

		assert.Error(t, err, "record not found")
		t.Log(file)
	})
}

func TestRepo_FindFileByHash(t *testing.T) {
	conf := config.TestConfig()

	search := New(conf.Db())

	t.Run("files found", func(t *testing.T) {
		file, err := search.FileByHash("123xxx")

		assert.Nil(t, err)
		assert.Equal(t, "exampleFileName.jpg", file.FileName)
	})

	t.Run("no files found", func(t *testing.T) {
		file, err := search.FileByHash("111")

		assert.Error(t, err, "record not found")
		t.Log(file)
	})
}

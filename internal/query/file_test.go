package query

import (
	"github.com/stretchr/testify/assert"
	"testing"

	"github.com/photoprism/photoprism/internal/config"
)

func TestQuery_Files(t *testing.T) {
	conf := config.TestConfig()

	search := New(conf.Db())

	t.Run("files found", func(t *testing.T) {
		files, err := search.Files(1000, 0)

		assert.Nil(t, err)
		assert.LessOrEqual(t, 5, len(files))
	})
}

func TestQuery_FilesByUUID(t *testing.T) {
	conf := config.TestConfig()

	search := New(conf.Db())

	t.Run("files found", func(t *testing.T) {
		files, err := search.FilesByUUID([]string{"ft8es39w45bnlqdw"}, 100, 0)

		assert.Nil(t, err)
		assert.Equal(t, 1, len(files))
		assert.Equal(t, "exampleFileName.jpg", files[0].FileName)
	})
}

func TestQuery_FileByPhotoUUID(t *testing.T) {
	conf := config.TestConfig()

	search := New(conf.Db())

	t.Run("files found", func(t *testing.T) {
		file, err := search.FileByPhotoUUID("pt9jtdre2lvl0yh8")

		assert.Nil(t, err)
		assert.Equal(t, "exampleDNGFile.dng", file.FileName)
	})

	t.Run("no files found", func(t *testing.T) {
		file, err := search.FileByPhotoUUID("111")

		assert.Error(t, err, "record not found")
		t.Log(file)
	})
}

func TestQuery_FileByUUID(t *testing.T) {
	conf := config.TestConfig()

	search := New(conf.Db())

	t.Run("files found", func(t *testing.T) {
		file, err := search.FileByUUID("ft8es39w45bnlqdw")

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "exampleFileName.jpg", file.FileName)
	})

	t.Run("no files found", func(t *testing.T) {
		file, err := search.FileByUUID("111")

		if err == nil {
			t.Fatal("error expected")
		}

		assert.Error(t, err, "record not found")
		t.Log(file)
	})
}

func TestQuery_FileByHash(t *testing.T) {
	conf := config.TestConfig()

	search := New(conf.Db())

	t.Run("files found", func(t *testing.T) {
		file, err := search.FileByHash("2cad9168fa6acc5c5c2965ddf6ec465ca42fd818")

		assert.Nil(t, err)
		assert.Equal(t, "exampleFileName.jpg", file.FileName)
	})

	t.Run("no files found", func(t *testing.T) {
		file, err := search.FileByHash("111")

		assert.Error(t, err, "record not found")
		t.Log(file)
	})
}

package query

import (
	"github.com/photoprism/photoprism/internal/entity"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFilesByPath(t *testing.T) {
	t.Run("files found", func(t *testing.T) {
		files, err := FilesByPath(entity.RootDefault, "2016/11")

		t.Logf("files: %+v", files)

		if err != nil {
			t.Fatal(err)
		}

		assert.LessOrEqual(t, 1, len(files))
	})
}

func TestExistingFiles(t *testing.T) {
	t.Run("files found", func(t *testing.T) {
		files, err := ExistingFiles(1000, 0, "/")

		t.Logf("files: %+v", files)

		if err != nil {
			t.Fatal(err)
		}
		assert.LessOrEqual(t, 5, len(files))
	})
	t.Run("search for files path", func(t *testing.T) {
		files, err := ExistingFiles(1000, 0, "Photos")

		t.Logf("files: %+v", files)

		if err != nil {
			t.Fatal(err)
		}

		assert.Empty(t, files)
	})
}

func TestFilesByUID(t *testing.T) {
	t.Run("files found", func(t *testing.T) {
		files, err := FilesByUID([]string{"ft8es39w45bnlqdw"}, 100, 0)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, 1, len(files))
		assert.Equal(t, "exampleFileName.jpg", files[0].FileName)
	})
	t.Run("no files found", func(t *testing.T) {
		files, err := FilesByUID([]string{"ft8es39w45bnlxxx"}, 100, 0)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, 0, len(files))
	})
}

func TestFileByPhotoUID(t *testing.T) {
	t.Run("files found", func(t *testing.T) {
		file, err := FileByPhotoUID("pt9jtdre2lvl0y11")

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, "bridge.jpg", file.FileName)
	})

	t.Run("no files found", func(t *testing.T) {
		file, err := FileByPhotoUID("111")

		assert.Error(t, err, "record not found")
		t.Log(file)
	})
}

func TestFileByUID(t *testing.T) {
	t.Run("files found", func(t *testing.T) {
		file, err := FileByUID("ft8es39w45bnlqdw")

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "exampleFileName.jpg", file.FileName)
	})

	t.Run("no files found", func(t *testing.T) {
		file, err := FileByUID("111")

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

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, "exampleFileName.jpg", file.FileName)
	})

	t.Run("no files found", func(t *testing.T) {
		file, err := FileByHash("111")

		assert.Error(t, err, "record not found")
		t.Log(file)
	})
}

func TestSetPhotoPrimary(t *testing.T) {
	assert.Equal(t, false, entity.FileFixturesExampleXMP.FilePrimary)

	err := SetPhotoPrimary("pt9jtdre2lvl0yh7", "ft2es49whhbnlqdn")

	if err != nil {
		t.Fatal(err)
	}
	//TODO How to assert
	//assert.Equal(t, true, entity.FileFixturesExampleXMP.FilePrimary)
}

package entity

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFirstFileByHash(t *testing.T) {
	t.Run("not existing file", func(t *testing.T) {
		f, err := FirstFileByHash("xxx")
		assert.Equal(t, "record not found", err.Error())
		assert.Equal(t, uint(0), f.ID)
	})
	t.Run("existing file", func(t *testing.T) {
		f, err := FirstFileByHash("2cad9168fa6acc5c5c2965ddf6ec465ca42fd818")
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, uint(0xf4240), f.ID)
	})
}

func TestFile_DownloadFileName(t *testing.T) {
	t.Run("photo with title", func(t *testing.T) {
		photo := &Photo{TakenAtLocal: time.Date(2019, 01, 15, 0, 0, 0, 0, time.UTC), PhotoTitle: "Berlin / Morning Mood"}
		file := &File{Photo: photo, FileType: "jpg", FileUUID: "foobar345678765"}

		filename := file.ShareFileName()

		assert.Contains(t, filename, "20190115-000000-Berlin-Morning-Mood")
		assert.Contains(t, filename, ".jpg")
	})
	t.Run("photo without title", func(t *testing.T) {
		photo := &Photo{TakenAtLocal: time.Date(2019, 01, 15, 0, 0, 0, 0, time.UTC), PhotoTitle: ""}
		file := &File{Photo: photo, FileType: "jpg", PhotoUUID: "123", FileUUID: "foobar345678765"}

		filename := file.ShareFileName()

		assert.Contains(t, filename, "20190115-000000-123")
		assert.Contains(t, filename, ".jpg")
	})
	t.Run("photo without photo", func(t *testing.T) {
		file := &File{Photo: nil, FileType: "jpg", FileHash: "123Hash", FileUUID: "foobar345678765"}

		filename := file.ShareFileName()

		assert.Equal(t, "123Hash.jpg", filename)
	})
}

func TestFile_Changed(t *testing.T) {
	var deletedAt = time.Date(2019, 01, 15, 0, 0, 0, 0, time.UTC)
	t.Run("different modified times", func(t *testing.T) {
		file := &File{Photo: nil, FileType: "jpg", FileSize: 500, FileModified: time.Date(2019, 01, 15, 0, 0, 0, 0, time.UTC)}
		time := time.Date(2020, 01, 15, 0, 0, 0, 0, time.UTC)
		assert.Equal(t, true, file.Changed(500, time))
	})
	t.Run("different sizes", func(t *testing.T) {
		file := &File{Photo: nil, FileType: "jpg", FileSize: 600, FileModified: time.Date(2019, 01, 15, 0, 0, 0, 0, time.UTC)}
		time := time.Date(2019, 01, 15, 0, 0, 0, 0, time.UTC)
		assert.Equal(t, true, file.Changed(500, time))
	})
	t.Run("no change", func(t *testing.T) {
		file := &File{Photo: nil, FileType: "jpg", FileSize: 500, FileModified: time.Date(2019, 01, 15, 0, 0, 0, 0, time.UTC)}
		time := time.Date(2019, 01, 15, 0, 0, 0, 0, time.UTC)
		assert.Equal(t, false, file.Changed(500, time))
	})
	t.Run("deleted", func(t *testing.T) {
		file := &File{Photo: nil, FileType: "jpg", FileSize: 500, FileModified: time.Date(2019, 01, 15, 0, 0, 0, 0, time.UTC), DeletedAt: &deletedAt}
		time := time.Date(2019, 01, 15, 0, 0, 0, 0, time.UTC)
		assert.Equal(t, true, file.Changed(500, time))
	})
}

func TestFile_Purge(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		file := &File{Photo: nil, FileType: "jpg", FileSize: 500}
		assert.Equal(t, nil, file.Purge())
	})
}

func TestFile_AllFilesMissing(t *testing.T) {
	t.Run("true", func(t *testing.T) {
		photo := &Photo{TakenAtLocal: time.Date(2019, 01, 15, 0, 0, 0, 0, time.UTC), PhotoTitle: ""}
		file := &File{Photo: photo, FileType: "jpg", PhotoUUID: "123", FileUUID: "123", FileMissing: true}
		file2 := &File{Photo: photo, FileType: "jpg", PhotoUUID: "123", FileUUID: "456", FileMissing: true}
		assert.True(t, file.AllFilesMissing())
		assert.NotEmpty(t, file2)
	})
	//TODO test false
	/*t.Run("false", func(t *testing.T) {
		file := FileFixturesExampleJPG
		assert.False(t, file.AllFilesMissing())
		assert.NotEmpty(t, file)
	})*/
}

func TestFile_Save(t *testing.T) {
	t.Run("record not found", func(t *testing.T) {
		file := &File{Photo: nil, FileType: "jpg", PhotoUUID: "123", FileUUID: "123"}
		err := file.Save()
		assert.Equal(t, "record not found", err.Error())
	})
	//TODO test success
}

package entity

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

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

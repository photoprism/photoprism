package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFile_DownloadFileName(t *testing.T) {
	photo := &Photo{TakenAt: time.Date(2019, 01, 15, 0, 0, 0, 0, time.UTC), PhotoTitle: "Berlin / Morning Mood"}
	file := &File{Photo: photo, FileType: "jpg"}

	filename := file.DownloadFileName()

	assert.Equal(t, "20190115-000000-Berlin-Morning-Mood.jpg", filename)
}

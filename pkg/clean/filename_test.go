package clean

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFileName(t *testing.T) {
	t.Run("Path", func(t *testing.T) {
		assert.Equal(t, "", FileName("/go/src/github.com/photoprism/photoprism"))
	})
	t.Run("File", func(t *testing.T) {
		assert.Equal(t, "filename.TXT", FileName("filename.TXT"))
	})
	t.Run("TheQuickBrownFox", func(t *testing.T) {
		assert.Equal(t, "The quick brown fox.", FileName("The quick brown fox."))
	})
	t.Run("FilenameTxt", func(t *testing.T) {
		assert.Equal(t, "filename.txt", FileName("filename.txt"))
	})
	t.Run("Empty", func(t *testing.T) {
		assert.Equal(t, "", FileName(""))
	})
	t.Run("Dot", func(t *testing.T) {
		assert.Equal(t, "", FileName("."))
	})
	t.Run("DotDot", func(t *testing.T) {
		assert.Equal(t, "", FileName(".."))
	})
	t.Run("DotDotDot", func(t *testing.T) {
		assert.Equal(t, "", FileName("..."))
	})
	t.Run("Replace", func(t *testing.T) {
		assert.Equal(t, "", FileName("${https://<host>:<port>/<path>}"))
	})
	t.Run("FileNameJpg", func(t *testing.T) {
		assert.Equal(t, "filename.jpg", FileName("file?name.jpg"))
	})
	t.Run("ControlCharacter", func(t *testing.T) {
		assert.Equal(t, "filename.", FileName("filename."+string(rune(127))))
	})
}

func TestFileNameRedacted(t *testing.T) {
	t.Run("DownloadArchive", func(t *testing.T) {
		assert.Equal(t, "photoprism-download-20260727-094439-***.zip", FileNameRedacted("photoprism-download-20260727-094439-zihqtuw4.zip"))
	})
	t.Run("NoDash", func(t *testing.T) {
		assert.Equal(t, "***.zip", FileNameRedacted("zihqtuw4.zip"))
	})
	t.Run("LeadingDash", func(t *testing.T) {
		assert.Equal(t, "***.zip", FileNameRedacted("-zihqtuw4.zip"))
	})
	t.Run("NoExtension", func(t *testing.T) {
		assert.Equal(t, "photoprism-download-***", FileNameRedacted("photoprism-download-zihqtuw4"))
	})
	t.Run("Empty", func(t *testing.T) {
		assert.Equal(t, "***", FileNameRedacted(""))
	})
}

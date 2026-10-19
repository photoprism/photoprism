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
	t.Run("RealNamesKept", func(t *testing.T) {
		for _, s := range []string{
			"\U0001F469\u200D\U0001F469\u200D\U0001F467.jpg",
			"\U0001F3F4\U000E0067\U000E0062\U000E0073\U000E0063\U000E0074\U000E007F.jpg",
			"Screenshot 2026-10-14 at 10.00.00\u202FAM.png",
			"mi\u200Cxwaham.jpg",
			"Gr\xfc\xdfe.jpg",
		} {
			assert.Equal(t, s, FileName(s), "%+q", s)
		}
	})
	t.Run("CheckedAfterRemoval", func(t *testing.T) {
		assert.Equal(t, "", FileName("a.*.jpg"))
		assert.Equal(t, "", FileName("a.\u202E.jpg"))
		assert.Equal(t, "", FileName(" . "))
	})
	t.Run("ASCIISpacesTrimmed", func(t *testing.T) {
		assert.Equal(t, "a b.jpg", FileName(" a b.jpg\t"))
		assert.Equal(t, "trail\u00A0.jpg", FileName("trail\u00A0.jpg "))
	})
	t.Run("UnsafeRemoved", func(t *testing.T) {
		assert.Equal(t, "photoabc.jpg", FileName("photo\u202Eabc.jpg"))
		assert.Equal(t, "abc.jpg", FileName("a\u0085b\u2028c.jpg"))
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

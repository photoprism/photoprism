package clean

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAttachmentName(t *testing.T) {
	t.Run("Unchanged", func(t *testing.T) {
		for _, s := range []string{
			"IMG_1234.JPG",
			"20240101-120000-Sommer am See.jpg",
			"صورة من العائلة.heic",
			"日本語のファイル.png",
			"👩\u200D👩\u200D👧.jpg",
			"Screenshot 2026-10-13 at 10.00.00 AM.png",
			"photo (2).jpg",
		} {
			assert.Equal(t, s, AttachmentName(s))
		}
	})
	t.Run("Removed", func(t *testing.T) {
		assert.Equal(t, "Abcde.jpg", AttachmentName("Ab\x00c\x01d\x1be\x7f.jpg"))
		assert.Equal(t, "photoabc.jpg", AttachmentName("photo\u202Eabc.jpg"))
		assert.Equal(t, "ab.jpg", AttachmentName("a\u200Bb.jpg"))
		assert.Equal(t, "a b c.jpg", AttachmentName("a\nb\rc.jpg"))
	})
	t.Run("Separators", func(t *testing.T) {
		assert.Equal(t, "Folder_Name.jpg", AttachmentName("Folder\\Name.jpg"))
		assert.Equal(t, "a_b.jpg", AttachmentName("a/b.jpg"))
	})
	t.Run("Windows", func(t *testing.T) {
		assert.Equal(t, "x_y.jpg", AttachmentName("x:y.jpg"))
		assert.Equal(t, "p_q.jpg", AttachmentName("p|q.jpg"))
		assert.Equal(t, "a_b_c_d_e_f.jpg", AttachmentName("a*b?c\"d<e>f.jpg"))
		assert.Equal(t, "日本_写真.jpg", AttachmentName("日本:写真.jpg"))
	})
	t.Run("Empty", func(t *testing.T) {
		assert.Equal(t, "download", AttachmentName(""))
		assert.Equal(t, "download", AttachmentName("\x00\x1b"))
		assert.Equal(t, "download", AttachmentName(" .. "))
		assert.Equal(t, "download.jpg", AttachmentName("\x00.jpg"))
		assert.Equal(t, "x.jpg", AttachmentName("\u200B.x.jpg"))
		assert.Equal(t, "download", AttachmentName("..."))
		assert.Equal(t, "download.jpg", AttachmentName("..jpg"))
		assert.Equal(t, "download.jpg", AttachmentName(". .jpg"))
		assert.Equal(t, "download", AttachmentName(".."))
	})
}

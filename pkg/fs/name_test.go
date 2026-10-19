package fs

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
)

func TestClipName(t *testing.T) {
	t.Run("Short", func(t *testing.T) {
		assert.Equal(t, "photo.jpg", ClipName("photo.jpg", 9))
		assert.Equal(t, "", ClipName("", 5))
	})
	t.Run("ASCII", func(t *testing.T) {
		assert.Equal(t, "phot", ClipName("photo.jpg", 4))
	})
	t.Run("RuneBoundary", func(t *testing.T) {
		// Each character is 3 bytes long, so a cut inside one keeps the characters before it.
		name := strings.Repeat("\u3042", 4)
		for maxLen, want := range map[int]string{12: name, 11: name[:9], 10: name[:9], 9: name[:9], 8: name[:6], 2: "", 1: ""} {
			got := ClipName(name, maxLen)
			assert.Equal(t, want, got, "maxLen %d", maxLen)
			assert.True(t, utf8.ValidString(got))
		}
		assert.Equal(t, "a\U0001F600"[:1], ClipName("a\U0001F600b", 4))
		assert.Equal(t, "a\U0001F600", ClipName("a\U0001F600b", 5))
	})
	t.Run("InvalidUTF8", func(t *testing.T) {
		// Continuation bytes without a start are cut at the limit.
		assert.Equal(t, "\x80\x80\x80\x80", ClipName("\x80\x80\x80\x80\x80\x80", 4))
	})
	t.Run("NoRoom", func(t *testing.T) {
		assert.Equal(t, "", ClipName("photo.jpg", 0))
		assert.Equal(t, "", ClipName("photo.jpg", -1))
	})
}

func TestRelName(t *testing.T) {
	t.Run("Same", func(t *testing.T) {
		assert.Equal(t, "", RelName("/some/path", "/some/path"))
	})
	t.Run("Short", func(t *testing.T) {
		assert.Equal(t, "/some/", RelName("/some/", "/some/path"))
	})
	t.Run("Empty", func(t *testing.T) {
		assert.Equal(t, "", RelName("", "/some/path"))
	})
	t.Run("SomePath", func(t *testing.T) {
		assert.Equal(t, "foo/bar.baz", RelName("/some/path/foo/bar.baz", "/some/path"))
	})
	t.Run("SomePath", func(t *testing.T) {
		assert.Equal(t, "foo/bar.baz", RelName("/some/path/foo/bar.baz", "/some/path/"))
	})
	t.Run("SomePathBar", func(t *testing.T) {
		assert.Equal(t, "/some/path/foo/bar.baz", RelName("/some/path/foo/bar.baz", "/some/path/bar"))
	})
	t.Run("EmptyDir", func(t *testing.T) {
		assert.Equal(t, "/some/path/foo/bar.baz", RelName("/some/path/foo/bar.baz", ""))
	})
	t.Run("SiblingPrefix", func(t *testing.T) {
		assert.Equal(t, "/data/photos2/2024/IMG_1.heic", RelName("/data/photos2/2024/IMG_1.heic", "/data/photos"))
		assert.Equal(t, "/data/photos-import/2024", RelName("/data/photos-import/2024", "/data/photos"))
		assert.Equal(t, "2024/IMG_1.heic", RelName("/data/photos/2024/IMG_1.heic", "/data/photos"))
	})
	t.Run("Relative", func(t *testing.T) {
		assert.Equal(t, "Test.jpg", RelName("testdata/Test.jpg", "testdata"))
		assert.Equal(t, "testdata2/Test.jpg", RelName("testdata2/Test.jpg", "testdata"))
	})
	t.Run("RootDir", func(t *testing.T) {
		assert.Equal(t, "some/path", RelName("/some/path", "/"))
	})
}

func TestInDir(t *testing.T) {
	t.Run("Below", func(t *testing.T) {
		assert.True(t, InDir("/data/photos/2024/IMG_1.heic", "/data/photos"))
		assert.True(t, InDir("/data/photos/2024/IMG_1.heic", "/data/photos/"))
		assert.True(t, InDir("testdata/Test.jpg", "testdata"))
		assert.True(t, InDir("/some/path", "/"))
	})
	t.Run("Same", func(t *testing.T) {
		assert.True(t, InDir("/data/photos", "/data/photos"))
	})
	t.Run("SiblingPrefix", func(t *testing.T) {
		assert.False(t, InDir("/data/photos2/2024/IMG_1.heic", "/data/photos"))
		assert.False(t, InDir("/data/photos-import", "/data/photos"))
		assert.False(t, InDir("/data/photo", "/data/photos"))
	})
	t.Run("Empty", func(t *testing.T) {
		assert.False(t, InDir("", "/data/photos"))
		assert.False(t, InDir("/data/photos/a.jpg", ""))
	})
}

func TestFileName(t *testing.T) {
	t.Run("CreatesFolder", func(t *testing.T) {
		baseDir := t.TempDir()
		sidecar := filepath.Join(t.TempDir(), "sidecar")

		result, err := FileName(filepath.Join(baseDir, "2026", "clip.avi"), sidecar, baseDir, ExtAvc)
		assert.NoError(t, err)
		assert.Equal(t, filepath.Join(sidecar, "2026", "clip.avi.avc"), result)
		assert.DirExists(t, filepath.Join(sidecar, "2026"))
	})
	t.Run("TestCopyThreeJpg", func(t *testing.T) {
		result, err := FileName("testdata/Test (4).jpg", ".photoprism", Abs("testdata"), ".xmp")
		assert.NoError(t, err)
		assert.Equal(t, "testdata/.photoprism/Test (4).jpg.xmp", result)
	})
	t.Run("TestThreeJpg", func(t *testing.T) {
		result, err := FileName("testdata/Test (4).jpg", ".photoprism", Abs("testdata"), ".xmp")
		assert.NoError(t, err)
		assert.Equal(t, "testdata/.photoprism/Test (4).jpg.xmp", result)
	})
	t.Run("FooXmp", func(t *testing.T) {
		result, err := FileName("testdata/FOO.XMP", ".photoprism/sub", Abs("testdata"), ".jpeg")
		assert.NoError(t, err)
		assert.Equal(t, "testdata/.photoprism/sub/FOO.XMP.jpeg", result)
	})
	t.Run("TestCopyThreeJpg", func(t *testing.T) {
		tempDir := filepath.Join(os.TempDir(), PPHiddenPathname)
		result, err := FileName(Abs("testdata/Test (4).jpg"), tempDir, Abs("testdata"), ".xmp")
		assert.NoError(t, err)
		assert.Equal(t, tempDir+"/Test (4).jpg.xmp", result)
	})
	t.Run("EmptyDir", func(t *testing.T) {
		result, err := FileName("testdata/FOO.XMP", "", Abs("testdata"), ".jpeg")
		assert.NoError(t, err)
		assert.Equal(t, "testdata/FOO.XMP.jpeg", result)
	})
}

func TestFilePath(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		baseDir := t.TempDir()
		sidecar := filepath.Join(t.TempDir(), "sidecar")

		result, err := FilePath(filepath.Join(baseDir, "2026", "clip.avi"), sidecar, baseDir, ExtAvc)
		assert.NoError(t, err)
		assert.Equal(t, filepath.Join(sidecar, "2026", "clip.avi.avc"), result)

		// Unlike FileName, it does not create the folder.
		assert.NoDirExists(t, filepath.Join(sidecar, "2026"))
	})
	t.Run("SameDir", func(t *testing.T) {
		result, err := FilePath("testdata/clip.avi", "", "", ExtAvc)
		assert.NoError(t, err)
		assert.Equal(t, "testdata/clip.avi.avc", result)
	})
	t.Run("EmptyName", func(t *testing.T) {
		_, err := FilePath("", ".photoprism", "", ExtAvc)
		assert.Error(t, err)
	})
	t.Run("EmptyExt", func(t *testing.T) {
		_, err := FilePath("testdata/clip.avi", ".photoprism", "", "")
		assert.Error(t, err)
	})
}

func TestFileNameHidden(t *testing.T) {
	t.Run("AtPrefix", func(t *testing.T) {
		assert.True(t, FileNameHidden("/some/path/@eaDir"))
	})
	t.Run("Dot", func(t *testing.T) {
		assert.True(t, FileNameHidden("/some/.folder"))
	})
	t.Run("Dots", func(t *testing.T) {
		assert.False(t, FileNameHidden("/some/image.jpg."))
		assert.False(t, FileNameHidden("./.some/foo"))
	})
	t.Run("Underscore", func(t *testing.T) {
		assert.False(t, FileNameHidden("/some/_folder"))
		assert.False(t, FileNameHidden("_folder"))
		assert.True(t, FileNameHidden("/some/_.folder"))
		assert.True(t, FileNameHidden("_.folder"))
		assert.True(t, FileNameHidden("/some/__MACOSX"))
		assert.True(t, FileNameHidden("__MACOSX"))
	})
	t.Run("At", func(t *testing.T) {
		assert.False(t, FileNameHidden("/some/path/ea@Dir"))
		assert.False(t, FileNameHidden("/some/@path/ea@Dir"))
		assert.False(t, FileNameHidden("@/eaDir"))
	})
	t.Run("DollarTilde", func(t *testing.T) {
		assert.True(t, FileNameHidden("/media/$RECYCLE.BIN"))
		assert.True(t, FileNameHidden("$"))
		assert.True(t, FileNameHidden("/some/~$report.docx"))
		assert.True(t, FileNameHidden("~"))
		assert.False(t, FileNameHidden("/some/price$5.jpg"))
		assert.False(t, FileNameHidden("/some/a~b.jpg"))
		assert.False(t, FileNameHidden("/~home/a.jpg"))
	})
	t.Run("False", func(t *testing.T) {
		assert.False(t, FileNameHidden("/some/path/folder"))
	})
	t.Run("TooShort", func(t *testing.T) {
		assert.False(t, FileNameHidden("a"))
	})
	t.Run("Empty", func(t *testing.T) {
		assert.False(t, FileNameHidden(""))
	})
}

func TestFileNameSystem(t *testing.T) {
	t.Run("System", func(t *testing.T) {
		for _, name := range []string{"/some/.folder", "/some/path/@eaDir", "_.folder", "__MACOSX", "."} {
			assert.True(t, FileNameSystem(name), name)
		}
	})
	t.Run("Other", func(t *testing.T) {
		for _, name := range []string{"/media/$RECYCLE.BIN", "~$report.docx", "~", "_folder", "a", ""} {
			assert.False(t, FileNameSystem(name), name)
		}
	})
}

// TestFilePath_InSearchDir verifies that a file in the sidecar folder outside the base folder gets its sidecar
// next to it, while other files keep their names below the sidecar folder.
func TestFilePath_InSearchDir(t *testing.T) {
	t.Run("FileInSidecar", func(t *testing.T) {
		result, err := FilePath("/storage/sidecar/2024/x.mp4", "/storage/sidecar", "/originals", ".jpg")
		assert.NoError(t, err)
		assert.Equal(t, "/storage/sidecar/2024/x.mp4.jpg", result)
	})
	t.Run("FileOutsideBaseDir", func(t *testing.T) {
		result, err := FilePath("/samples/x.mp4", "/storage/sidecar", "/originals", ".jpg")
		assert.NoError(t, err)
		assert.Equal(t, "/storage/sidecar/samples/x.mp4.jpg", result)
	})
	t.Run("SidecarInOriginals", func(t *testing.T) {
		result, err := FilePath("/originals/.photoprism/sidecar/2024/x.mp4", "/originals/.photoprism/sidecar", "/originals", ".jpg")
		assert.NoError(t, err)
		assert.Equal(t, "/originals/.photoprism/sidecar/.photoprism/sidecar/2024/x.mp4.jpg", result)
	})
}

func TestNameTooLong(t *testing.T) {
	name := strings.Repeat("n", NameMaxLen)
	t.Run("Empty", func(t *testing.T) {
		assert.False(t, NameTooLong(""))
		assert.False(t, NameTooLong("/"))
		assert.False(t, NameTooLong("//a//"))
	})
	t.Run("AtLimit", func(t *testing.T) {
		assert.False(t, NameTooLong(name))
		assert.False(t, NameTooLong("/"+name+"/"+name+"/"))
	})
	t.Run("OverLimit", func(t *testing.T) {
		assert.True(t, NameTooLong(name+"n"))
		assert.True(t, NameTooLong("/"+name+"n/file.jpg"))
		assert.True(t, NameTooLong("/folder/"+name+"n"))
	})
}

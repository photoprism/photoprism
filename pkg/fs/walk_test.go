package fs

import (
	"io/fs"
	"testing"

	"github.com/karrick/godirwalk"
	"github.com/stretchr/testify/assert"
)

func TestSkipWalk(t *testing.T) {
	t.Run("Done", func(t *testing.T) {
		done := make(Done)
		ignore := NewIgnoreList(".ppignore", true, false)

		done["foo.jpg"] = Found

		if skip, result := SkipWalk("testdata/directory", true, false, done, ignore); skip {
			assert.Nil(t, result)
		} else {
			t.Fatal("skip should be true because this is a directory and not a file")
		}

		assert.True(t, done["foo.jpg"].Exists())
		assert.True(t, done["testdata/directory"].Exists())
		assert.Equal(t, 2, len(done))
	})

	t.Run("Storage", func(t *testing.T) {
		done := make(Done)
		ignore := NewIgnoreList(".ppignore", true, false)

		if skip, result := SkipWalk("testdata/originals", true, false, done, ignore); skip {
			assert.Nil(t, result)
		} else {
			t.Fatal("skip should be true because this is a directory and not a file")
		}

		if skip, result := SkipWalk("testdata/originals/storage", true, false, done, ignore); skip {
			assert.Equal(t, fs.SkipDir, result)
		} else {
			t.Fatal("skip should be true because this is a directory and not a file")
		}
		assert.True(t, done["testdata/originals"].Exists())
		assert.True(t, done["testdata/originals/storage"].Exists())
		assert.Equal(t, 2, len(done))
	})

	t.Run("Symlink", func(t *testing.T) {
		done := make(Done)
		ignore := NewIgnoreList(".ppignore", true, false)

		if skip, result := SkipWalk("testdata/directory/subdirectory/symlink", false, true, done, ignore); skip {
			assert.Nil(t, result)
		} else {
			t.Fatal("should be skipped")
		}

		if skip, result := SkipWalk("testdata/directory/subdirectory/symlink/self", false, true, done, ignore); skip {
			assert.Error(t, result)
		} else {
			t.Fatal("should be skipped")
		}

		if skip, result := SkipWalk("testdata/directory/subdirectory/symlink/self/self", false, true, done, ignore); skip {
			assert.Error(t, result)
		} else {
			t.Fatal("should be skipped")
		}

		assert.True(t, done["testdata/linked"].Exists())
		assert.True(t, done["testdata/directory/subdirectory/symlink"].Exists())
		assert.True(t, done["testdata/directory/subdirectory/symlink/self"].Exists())
		assert.True(t, done["testdata/directory/subdirectory/symlink/self/self"].Exists())
		assert.Equal(t, 4, len(done))
	})

	t.Run("Godirwalk", func(t *testing.T) {
		done := make(Done)
		var skipped []string
		var dirs []string
		testPath := "testdata"
		ignore := NewIgnoreList(".ppignore", true, false)

		err := godirwalk.Walk(testPath, &godirwalk.Options{
			Callback: func(fileName string, info *godirwalk.Dirent) error {
				isDir, _ := info.IsDirOrSymlinkToDir()
				isSymlink := info.IsSymlink()

				if skip, result := SkipWalk(fileName, isDir, isSymlink, done, ignore); skip {
					if result != nil {
						dirs = append(dirs, fileName)
					} else {
						skipped = append(skipped, fileName)
					}
					return result
				}

				done[fileName] = Found

				if textName := SidecarText.Find(fileName, false); textName != "" {
					done[textName] = Found
				}

				return nil
			},
			Unsorted:            false,
			FollowSymbolicLinks: true,
		})

		assert.NoError(t, err)
		assert.Contains(t, dirs, "testdata/directory/subdirectory/.hiddendir")

		expected := []string{
			"testdata",
			"testdata/directory",
			"testdata/directory/.ppignore",
			"testdata/directory/bar.txt",
			"testdata/directory/baz.xml",
			"testdata/directory/subdirectory",
			"testdata/directory/subdirectory/.hiddenfile",
			"testdata/directory/subdirectory/.ppignore",
			"testdata/directory/subdirectory/animals",
			"testdata/directory/subdirectory/animals/.ppignore",
			"testdata/directory/subdirectory/animals/dog.json",
			"testdata/directory/subdirectory/animals/gopher.json",
			"testdata/directory/subdirectory/animals/gopher.md",
			"testdata/directory/subdirectory/example.txt",
			"testdata/directory/subdirectory/foo.txt",
			"testdata/directory/subdirectory/symlink",
			"testdata/directory/subdirectory/symlink/somefile.txt",
			"testdata/directory/subdirectory/symlink/test.md",
			"testdata/directory/subdirectory/symlink/test.txt",
			"testdata/originals",
		}

		assert.Equal(t, expected, skipped)
	})
}

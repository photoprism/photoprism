package fs

import (
	"path/filepath"
	"testing"

	"github.com/karrick/godirwalk"
	"github.com/stretchr/testify/assert"
)

func TestIgnoreList_Ignore(t *testing.T) {
	t.Run(".ppignore", func(t *testing.T) {
		list := NewIgnoreList(".ppignore", true, false)
		assert.False(t, list.Ignore("testdata"))
		assert.False(t, list.Ignore("testdata/directory"))
		assert.True(t, list.Ignore("testdata/directory/.ppignore"), ".ppignore should be ignored")
		assert.True(t, list.Ignore("testdata/directory/bar.txt"), "bar.txt should be ignored")
		assert.False(t, list.Ignore("testdata/directory/example.bmp"))
		assert.True(t, list.Ignore("testdata/directory/subdirectory/.hiddendir"), ".hiddendir should be ignored")
		assert.True(t, list.Ignore("testdata/directory/subdirectory/foo.txt"), "foo.txt should be ignored")
		assert.True(t, list.Ignore("testdata/directory/subdirectory/symlink/somefile.txt"), "somefile.txt should be ignored")
		assert.True(t, list.Ignore("testdata/directory/subdirectory/symlink/test.md"), "test.md should be ignored")
		assert.False(t, list.Ignore("testdata/directory/subdirectory/symlink/test.xml"))
		assert.False(t, list.Ignore("testdata/directory/subdirectory/symlink/test.yml"))
		assert.True(t, list.Ignore("testdata/directory/subdirectory/symlink/test.txt"), "test.txt should be ignored")
	})
}

func TestIgnoreList_Hidden(t *testing.T) {
	t.Run("ignore hidden", func(t *testing.T) {
		testPath := "testdata/directory/subdirectory"
		ignore := NewIgnoreList(".ppignore", true, false)

		err := godirwalk.Walk(testPath, &godirwalk.Options{
			Callback: func(fileName string, info *godirwalk.Dirent) error {
				if ignore.Ignore(fileName) && info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			},
			Unsorted:            false,
			FollowSymbolicLinks: false,
		})

		assert.NoError(t, err)

		expectHidden := []string{
			"testdata/directory/subdirectory/.hiddendir",
			"testdata/directory/subdirectory/.hiddenfile",
		}

		assert.Equal(t, expectHidden, ignore.Hidden())
	})

	t.Run("accept hidden", func(t *testing.T) {
		testPath := "testdata/directory/subdirectory"
		ignore := NewIgnoreList(".ppignore", false, false)

		err := godirwalk.Walk(testPath, &godirwalk.Options{
			Callback: func(fileName string, info *godirwalk.Dirent) error {
				if ignore.Ignore(fileName) && info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			},
			Unsorted:            false,
			FollowSymbolicLinks: false,
		})

		assert.NoError(t, err)

		assert.Equal(t, 0, len(ignore.Hidden()))
	})
}

func TestIgnoreList_Ignored(t *testing.T) {
	t.Run("has ignored", func(t *testing.T) {
		testPath := "testdata/directory"
		ignore := NewIgnoreList(".ppignore", true, false)

		err := godirwalk.Walk(testPath, &godirwalk.Options{
			Callback: func(fileName string, info *godirwalk.Dirent) error {
				if ignore.Ignore(fileName) && info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			},
			Unsorted:            false,
			FollowSymbolicLinks: false,
		})

		assert.NoError(t, err)

		expectIgnored := []string{
			"testdata/directory/bar.txt",
			"testdata/directory/baz.xml",
			"testdata/directory/subdirectory/animals/dog.json",
			"testdata/directory/subdirectory/animals/gopher.json",
			"testdata/directory/subdirectory/animals/gopher.md",
			"testdata/directory/subdirectory/bar",
			"testdata/directory/subdirectory/example.txt",
			"testdata/directory/subdirectory/foo.txt",
		}

		assert.Equal(t, expectIgnored, ignore.Ignored())
	})

	t.Run("no ignored", func(t *testing.T) {
		testPath := "testdata/directory"
		ignore := NewIgnoreList(".xyz", false, false)

		err := godirwalk.Walk(testPath, &godirwalk.Options{
			Callback: func(fileName string, info *godirwalk.Dirent) error {
				if ignore.Ignore(fileName) && info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			},
			Unsorted:            false,
			FollowSymbolicLinks: false,
		})

		assert.NoError(t, err)

		assert.Equal(t, 0, len(ignore.Ignored()))
	})
}

func TestNewIgnoreItem(t *testing.T) {
	t.Run("case sensitive false", func(t *testing.T) {
		ignore := NewIgnoreItem("testdata/directory", "Test_", false)
		assert.Equal(t, "test_", ignore.Pattern)
	})

	t.Run("case sensitive true", func(t *testing.T) {
		ignore := NewIgnoreItem("testdata/directory", "Test_", true)
		assert.Equal(t, "Test_", ignore.Pattern)
	})
}

func TestIgnoreList_AppendItems(t *testing.T) {
	t.Run("Error", func(t *testing.T) {
		ignoreList := NewIgnoreList(".xyz", false, false)
		assert.Error(t, ignoreList.AppendItems("", []string{"__test_"}))
	})

	t.Run("Success", func(t *testing.T) {
		ignoreList := NewIgnoreList(".xyz", false, false)
		assert.Nil(t, ignoreList.AppendItems("testdata/directory", []string{"__test_"}))
	})
}

func TestIgnoreList_Dir(t *testing.T) {
	t.Run("error empty directory name", func(t *testing.T) {
		ignoreList := NewIgnoreList(".xyz", false, false)
		assert.Error(t, ignoreList.Dir(""))
	})

	t.Run("error empty ignore file name", func(t *testing.T) {
		ignoreList := NewIgnoreList("", false, false)
		assert.Error(t, ignoreList.Dir("testdata"))
	})

	t.Run("no file found", func(t *testing.T) {
		ignoreList := NewIgnoreList(".xyz", false, false)
		assert.Error(t, ignoreList.Dir("./testdata"))
	})

	t.Run("nil", func(t *testing.T) {
		ignoreList := NewIgnoreList(".ppignore", false, false)
		assert.Nil(t, ignoreList.Dir("./testdata/directory"))
	})
}

func TestIgnoreList_Reset(t *testing.T) {
	ignoreList := NewIgnoreList(".xyz", false, false)
	ignoreList.AppendItems("testdata123/directory", []string{"__test_"})
	assert.Equal(t, "testdata123/directory/", ignoreList.items[0].Dir)
	ignoreList.Reset()
	assert.Len(t, ignoreList.items, 0)
}

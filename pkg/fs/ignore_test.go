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
	t.Run("IgnoreHidden", func(t *testing.T) {
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
	t.Run("AcceptHidden", func(t *testing.T) {
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
	t.Run("Directories", func(t *testing.T) {
		testPath := "testdata/directory"
		ignore := NewIgnoreList(".ppignore", true, false)
		ignore.Log = func(fileName string) {
			t.Logf(`TestIgnoreList_Ignored/Directories: ignored %s`, fileName)
		}

		err := godirwalk.Walk(testPath, &godirwalk.Options{
			Callback: func(name string, info *godirwalk.Dirent) error {
				if info.IsDir() {
					if err := ignore.Path(name); err != nil {
						t.Logf("configPath(%s) error: %s", name, err)
					}

					if ignore.Ignore(name) {
						return filepath.SkipDir
					}
				} else if ignore.Ignore(name) {
					return nil
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
			"testdata/directory/subdirectory/animals/private",
			"testdata/directory/subdirectory/animals/test files",
			"testdata/directory/subdirectory/bar",
			"testdata/directory/subdirectory/example.txt",
			"testdata/directory/subdirectory/foo.txt",
		}

		assert.Equal(t, expectIgnored, ignore.Ignored())
	})
	t.Run("NoIgnored", func(t *testing.T) {
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

func TestNewIgnorePattern(t *testing.T) {
	t.Run("CaseSensitiveFalse", func(t *testing.T) {
		ignore := NewIgnorePattern("testdata/directory", "Test_", false)
		assert.Equal(t, "test_", ignore.Pattern)
	})
	t.Run("CaseSensitiveTrue", func(t *testing.T) {
		ignore := NewIgnorePattern("testdata/directory", "Test_", true)
		assert.Equal(t, "Test_", ignore.Pattern)
	})
}

func TestIgnoreList_AddPatterns(t *testing.T) {
	t.Run("EmptyDir", func(t *testing.T) {
		ignoreList := NewIgnoreList(".xyz", false, false)
		assert.Error(t, ignoreList.AddPatterns("", []string{"__test_"}))
	})
	t.Run("Success", func(t *testing.T) {
		ignoreList := NewIgnoreList(".xyz", false, false)
		assert.Nil(t, ignoreList.AddPatterns("testdata/directory", []string{"__test_"}))
		result := ignoreList.ignore
		assert.Len(t, result, 1)
		assert.Equal(t, "__test_", result[0].Pattern)
	})
	t.Run("EmptyPattern", func(t *testing.T) {
		ignoreList := NewIgnoreList(".xyz", false, false)
		assert.Nil(t, ignoreList.AddPatterns("testdata/directory", []string{}))
		result := ignoreList.ignore
		assert.Len(t, result, 0)
	})
	t.Run("Trim", func(t *testing.T) {
		ignoreList := NewIgnoreList(".xyz", false, false)
		assert.Nil(t, ignoreList.AddPatterns("testdata/directory", []string{"/foo/bar/", "/", "\t\n\nbaz/\t\n\r", "test "}))
		result := ignoreList.ignore
		assert.Len(t, result, 3)
		assert.Equal(t, "foo/bar", result[0].Pattern)
		assert.Equal(t, "baz", result[1].Pattern)
		assert.Equal(t, "test ", result[2].Pattern)
	})
}

func TestIgnoreList_Path(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		ignoreList := NewIgnoreList(".ppignore", false, false)
		assert.Nil(t, ignoreList.Path("./testdata/directory"))
	})
	t.Run("FileNotFound", func(t *testing.T) {
		ignoreList := NewIgnoreList(".xyz", false, false)
		assert.Nil(t, ignoreList.Path("./testdata"))
	})
	t.Run("EmptyPathName", func(t *testing.T) {
		ignoreList := NewIgnoreList(".xyz", false, false)
		assert.Error(t, ignoreList.Path(""))
	})
	t.Run("EmptyFileName", func(t *testing.T) {
		ignoreList := NewIgnoreList("", false, false)
		assert.Error(t, ignoreList.Path("testdata"))
	})
}

func TestIgnoreList_File(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		ignoreList := NewIgnoreList(".ppignore", false, false)
		assert.Nil(t, ignoreList.File("./testdata/directory/subdirectory/.ppignore"))
		result := ignoreList.ignore
		assert.Len(t, result, 1)
		assert.Equal(t, "bar", result[0].Pattern)
	})
	t.Run("FileNotFound", func(t *testing.T) {
		ignoreList := NewIgnoreList(".ppignore", false, false)
		assert.Nil(t, ignoreList.File("./testdata/directory/subdirectory/.xxx"))
		result := ignoreList.ignore
		assert.Len(t, result, 0)
	})
	t.Run("EmptyFileName", func(t *testing.T) {
		ignoreList := NewIgnoreList(".ppignore", false, false)
		assert.Error(t, ignoreList.File(""))
	})
}

func TestIgnoreList_Reset(t *testing.T) {
	ignoreList := NewIgnoreList(".xyz", false, false)

	if err := ignoreList.AddPatterns("testdata123/directory", []string{"__test_"}); err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "testdata123/directory/", ignoreList.ignore[0].Dir)
	ignoreList.Reset()
	assert.Len(t, ignoreList.ignore, 0)
}

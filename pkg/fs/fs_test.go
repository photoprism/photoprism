package fs

import (
	"fmt"
	"os"
	"os/user"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	if insensitive, err := CaseInsensitive(os.TempDir()); err != nil {
		fmt.Println(err)
	} else if insensitive {
		IgnoreCase()
	}

	code := m.Run()

	os.Exit(code)
}

func TestFileExists(t *testing.T) {
	assert.True(t, FileExists("./testdata/test.jpg"))
	assert.True(t, FileExists("./testdata/test.jpg"))
	assert.True(t, FileExists("./testdata/empty.jpg"))
	assert.False(t, FileExists("./foo.jpg"))
	assert.False(t, FileExists(""))
}

func TestFileExistsNotEmpty(t *testing.T) {
	assert.True(t, FileExistsNotEmpty("./testdata/test.jpg"))
	assert.True(t, FileExistsNotEmpty("./testdata/test.jpg"))
	assert.False(t, FileExistsNotEmpty("./testdata/empty.jpg"))
	assert.False(t, FileExistsNotEmpty("./foo.jpg"))
	assert.False(t, FileExistsNotEmpty(""))
}

func TestPathExists(t *testing.T) {
	assert.True(t, PathExists("./testdata"))
	assert.False(t, PathExists("./testdata/test.jpg"))
	assert.False(t, PathExists("./testdata3ggdtgdg"))
	assert.False(t, PathExists(""))
}

func TestPathWritable(t *testing.T) {
	assert.True(t, PathWritable("./testdata"))
	assert.False(t, PathWritable("./testdata/test.jpg"))
	assert.False(t, PathWritable("./testdata3ggdtgdg"))
	assert.False(t, PathWritable(""))
}

func TestWritable(t *testing.T) {
	assert.True(t, Writable("./testdata"))
	assert.False(t, Writable("./testdata3ggdtgdg"))
	assert.False(t, Writable(""))
}

func TestOverwrite(t *testing.T) {
	data := make([]byte, 3)
	data[1] = 3
	data[2] = 8
	tmpPath := "./testdata/_tmp"
	os.Mkdir(tmpPath, 0777)

	defer os.RemoveAll(tmpPath)
	result := Overwrite("./testdata/_tmp/notyetexisting.jpg", data)
	assert.FileExists(t, "./testdata/_tmp/notyetexisting.jpg")
	assert.True(t, result)
}

func TestExpandedFilename(t *testing.T) {
	t.Run("test.jpg", func(t *testing.T) {
		filename := Abs("./testdata/test.jpg")
		assert.Contains(t, filename, "/testdata/test.jpg")
		assert.IsType(t, "", filename)
	})
	t.Run("empty filename", func(t *testing.T) {
		filename := Abs("")
		assert.Equal(t, "", filename)
		assert.IsType(t, "", filename)
	})
	t.Run("~ in filename", func(t *testing.T) {
		usr, _ := user.Current()
		expected := usr.HomeDir + "/test.jpg"
		filename := Abs("~/test.jpg")
		assert.Equal(t, expected, filename)
		assert.IsType(t, "", filename)
	})
}

func TestDirIsEmpty(t *testing.T) {
	t.Run("CurrentDir", func(t *testing.T) {
		assert.Equal(t, false, DirIsEmpty("."))
	})
	t.Run("Testdata", func(t *testing.T) {
		assert.Equal(t, false, DirIsEmpty("./testdata"))
	})
	t.Run("XXX", func(t *testing.T) {
		assert.Equal(t, false, DirIsEmpty("./xxx"))
	})
	t.Run("EmptyDir", func(t *testing.T) {
		if err := os.Mkdir("./testdata/emptyDir", 0777); err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll("./testdata/emptyDir")
		assert.Equal(t, true, DirIsEmpty("./testdata/emptyDir"))
	})
}

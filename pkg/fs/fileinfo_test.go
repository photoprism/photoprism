package fs

import (
	"os"
	"testing"
	"time"

	"github.com/emersion/go-webdav"

	"github.com/stretchr/testify/assert"
)

func TestNewFileInfo(t *testing.T) {
	info, err := os.Stat("testdata/test.jpg")

	if err != nil {
		t.Fatal(err)
	}

	result := NewFileInfo(info, "Photos/")

	assert.Equal(t, "test.jpg", result.Name)
	assert.Equal(t, "/Photos/test.jpg", result.Abs)
	assert.Equal(t, int64(10990), result.Size)
	assert.IsType(t, time.Time{}, result.Date)
	assert.Equal(t, false, result.Dir)
}

func TestNewFileInfos(t *testing.T) {
	dirEntries, err := os.ReadDir("testdata")

	if err != nil {
		t.Fatal(err)
	}

	infos := make([]os.FileInfo, 0, len(dirEntries))

	for _, entry := range dirEntries {
		info, err := entry.Info()
		if err != nil {
			t.Fatal(err)
		}
		infos = append(infos, info)
	}

	result := NewFileInfos(infos, PathSeparator)

	if len(result) < 1 {
		t.Fatal("empty result")
	}

	expected := map[string]FileInfo{
		"empty.jpg":     {Abs: PathSeparator + "empty.jpg", Size: 0, Dir: false},
		"test.jpg":      {Abs: PathSeparator + "test.jpg", Size: 10990, Dir: false},
		"CATYELLOW.jpg": {Abs: PathSeparator + "CATYELLOW.jpg", Size: 70790, Dir: false},
		"directory":     {Abs: PathSeparator + "directory", Size: 4096, Dir: true},
		"linked":        {Abs: PathSeparator + "linked", Size: 4096, Dir: true},
	}

	for _, file := range result {
		assert.NotEmpty(t, file.Name)
		assert.NotEmpty(t, file.Abs)
		assert.False(t, file.Date.IsZero())

		if info, ok := expected[file.Name]; ok {
			assert.Equal(t, info.Abs, file.Abs, "%s expected for %s", info.Abs, file.Name)
			if info.Dir {
				assert.GreaterOrEqual(t, info.Size, file.Size, "size < %d expected for %s", info.Size, file.Name)
			} else {
				assert.Equal(t, info.Size, file.Size, "%d expected for %s", info.Size, file.Name)
			}
			assert.Equal(t, info.Dir, file.Dir, "%t expected for %s", info.Dir, file.Name)
		}
	}
}

func TestWebFileInfo(t *testing.T) {

	infos := webdav.FileInfo{Path: "testdata", Size: 3, ModTime: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC), IsDir: true, MIMEType: "image"}

	result := WebFileInfo(infos, PathSeparator)

	assert.Equal(t, "testdata", result.Name)
	assert.Equal(t, "/testdata", result.Abs)
	assert.Equal(t, int64(3), result.Size)
	assert.Equal(t, "2020-01-01 00:00:00 +0000 UTC", result.Date.String())
	assert.Equal(t, true, result.Dir)
}

func TestFileInfos_Less(t *testing.T) {
	infos := FileInfos{
		{Name: "test.jpg", Abs: "/test.jpg", Size: 10990, Dir: false},
		{Name: "CATYELLOW.jpg", Abs: "/CATYELLOW.jpg", Size: 70790, Dir: false},
		{Name: "directory", Abs: "/directory", Size: 256, Dir: true},
		{Name: "linked", Abs: "/linked", Size: 256, Dir: true},
	}

	assert.Equal(t, false, infos.Less(0, 0))
	assert.Equal(t, false, infos.Less(0, 1))
	assert.Equal(t, true, infos.Less(1, 0))
	assert.Equal(t, true, infos.Less(2, 3))
	assert.Equal(t, false, infos.Less(3, 2))
}

func TestFileInfos_Swap(t *testing.T) {
	infos := FileInfos{
		{Name: "test.jpg", Abs: "/test.jpg", Size: 10990, Dir: false},
		{Name: "CATYELLOW.jpg", Abs: "/CATYELLOW.jpg", Size: 70790, Dir: false},
		{Name: "directory", Abs: "/directory", Size: 256, Dir: true},
		{Name: "linked", Abs: "/linked", Size: 256, Dir: true},
	}

	assert.Equal(t, "test.jpg", infos[0].Name)
	assert.Equal(t, "CATYELLOW.jpg", infos[1].Name)
	infos.Swap(0, 1)
	assert.Equal(t, "CATYELLOW.jpg", infos[0].Name)
	assert.Equal(t, "test.jpg", infos[1].Name)
}

func TestFileInfos_Len(t *testing.T) {
	infos := FileInfos{
		{Name: "test.jpg", Abs: "/test.jpg", Size: 10990, Dir: false},
		{Name: "CATYELLOW.jpg", Abs: "/CATYELLOW.jpg", Size: 70790, Dir: false},
		{Name: "directory", Abs: "/directory", Size: 256, Dir: true},
		{Name: "linked", Abs: "/linked", Size: 256, Dir: true},
	}

	assert.Equal(t, 4, infos.Len())
}

func TestFileInfos_Abs(t *testing.T) {
	infos := FileInfos{
		{Name: "test.jpg", Abs: "/test.jpg", Size: 10990, Dir: false},
		{Name: "CATYELLOW.jpg", Abs: "/CATYELLOW.jpg", Size: 70790, Dir: false},
		{Name: "directory", Abs: "/directory", Size: 256, Dir: true},
		{Name: "linked", Abs: "/linked", Size: 256, Dir: true},
	}

	assert.Equal(t, []string{"/test.jpg", "/CATYELLOW.jpg", "/directory", "/linked"}, infos.Abs())
}

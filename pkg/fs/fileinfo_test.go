package fs

import (
	"io/ioutil"
	"os"
	"testing"
	"time"

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
	infos, err := ioutil.ReadDir("testdata")

	if err != nil {
		t.Fatal(err)
	}

	result := NewFileInfos(infos, "/")

	if len(result) < 1 {
		t.Fatal("empty result")
	}

	assert.Equal(t, "test.jpg", result[2].Name)
	assert.Equal(t, "/test.jpg", result[2].Abs)
	assert.Equal(t, int64(10990), result[2].Size)
	assert.IsType(t, time.Time{}, result[2].Date)
	assert.Equal(t, false, result[2].Dir)
}

func TestFileInfos_Less(t *testing.T) {
	infos, err := ioutil.ReadDir("testdata")

	if err != nil {
		t.Fatal(err)
	}

	result := NewFileInfos(infos, "/")

	if len(result) < 1 {
		t.Fatal("empty result")
	}

	assert.Equal(t, 3, result.Len())
	assert.Equal(t, []string{"/CATYELLOW.jpg", "/chameleon_lime.jpg", "/test.jpg"}, result.Abs())
	assert.Equal(t, false, result.Less(0, 0))
	assert.Equal(t, true, result.Less(0, 1))

}

func TestFileInfos_Swap(t *testing.T) {
	infos, err := ioutil.ReadDir("testdata")

	if err != nil {
		t.Fatal(err)
	}

	result := NewFileInfos(infos, "/")

	if len(result) < 1 {
		t.Fatal("empty result")
	}

	assert.Equal(t, 3, result.Len())
	assert.Equal(t, []string{"/CATYELLOW.jpg", "/chameleon_lime.jpg", "/test.jpg"}, result.Abs())
	result.Swap(0, 1)
	assert.Equal(t, 3, result.Len())
	assert.Equal(t, []string{"/chameleon_lime.jpg", "/CATYELLOW.jpg", "/test.jpg"}, result.Abs())

}

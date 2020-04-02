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

	assert.Equal(t, "test.jpg", result[0].Name)
	assert.Equal(t, "/test.jpg", result[0].Abs)
	assert.Equal(t, int64(10990), result[0].Size)
	assert.IsType(t, time.Time{}, result[0].Date)
	assert.Equal(t, false, result[0].Dir)
}

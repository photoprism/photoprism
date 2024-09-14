package webdav

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/rnd"
)

const (
	testUrl  = "http://dummy-webdav/"
	testUser = "admin"
	testPass = "photoprism"
)

func TestClientUrl(t *testing.T) {
	result, err := clientUrl(testUrl, testUser, testPass)

	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "http://admin:photoprism@dummy-webdav/", result.String())
}

func TestNewClient(t *testing.T) {
	c, err := NewClient(testUrl, testUser, testPass, TimeoutLow)

	if err != nil {
		t.Fatal(err)
	}

	assert.IsType(t, &Client{}, c)
}

func TestClient_Files(t *testing.T) {
	c, err := NewClient(testUrl, testUser, testPass, TimeoutLow)

	if err != nil {
		t.Fatal(err)
	}

	assert.IsType(t, &Client{}, c)

	t.Run("NonRecursive", func(t *testing.T) {
		files, err := c.Files("Photos", false)

		if err != nil {
			t.Fatal(err)
		}

		if len(files) == 0 {
			t.Fatal("no files found")
		}
	})
	t.Run("Recursive", func(t *testing.T) {
		files, err := c.Files("Photos", true)

		if err != nil {
			t.Fatal(err)
		}

		if len(files) == 0 {
			t.Fatal("no files found")
		}
	})
}

func TestClient_Directories(t *testing.T) {
	c, err := NewClient(testUrl, testUser, testPass, TimeoutLow)

	if err != nil {
		t.Fatal(err)
	}

	assert.IsType(t, &Client{}, c)

	t.Run("NonRecursive", func(t *testing.T) {
		dirs, err := c.Directories("", false, MaxRequestDuration)

		if err != nil {
			t.Fatal(err)
		}

		if len(dirs) == 0 {
			t.Fatal("no directories found")
		}

		assert.IsType(t, fs.FileInfo{}, dirs[0])
		assert.Equal(t, "Photos", dirs[0].Name)
		assert.Equal(t, "/Photos", dirs[0].Abs)
		assert.Equal(t, true, dirs[0].Dir)
		assert.Equal(t, int64(0), dirs[0].Size)
	})
	t.Run("Recursive", func(t *testing.T) {
		dirs, err := c.Directories("", true, 0)

		if err != nil {
			t.Fatal(err)
		}

		if len(dirs) < 2 {
			t.Fatal("at least 2 directories expected")
		}
	})
}

func TestClient_UploadAndDelete(t *testing.T) {
	c, err := NewClient(testUrl, testUser, testPass, TimeoutLow)

	if err != nil {
		t.Fatal(err)
	}

	assert.IsType(t, &Client{}, c)

	tempName := rnd.UUID() + fs.ExtJPEG

	if err := c.Upload(fs.Abs("testdata/example.jpg"), tempName); err != nil {
		t.Fatal(err)
	}

	if err := c.Delete(tempName); err != nil {
		t.Fatal(err)
	}
}

func TestClient_Download(t *testing.T) {
	c, err := NewClient(testUrl, testUser, testPass, TimeoutDefault)

	if err != nil {
		t.Fatal(err)
	}

	assert.IsType(t, &Client{}, c)

	files, err := c.Files("Photos", false)

	if err != nil {
		t.Fatal(err)
	}

	tempDir := filepath.Join(os.TempDir(), rnd.UUID())
	tempFile := tempDir + "/foo.jpg"

	if len(files) == 0 {
		t.Fatal("no files to download")
	}

	if err := c.Download(files[0].Abs, tempFile, false); err != nil {
		t.Fatal(err)
	}

	if !fs.FileExists(tempFile) {
		t.Fatalf("%s does not exist", tempFile)
	}

	if err := os.RemoveAll(tempDir); err != nil {
		t.Fatal(err)
	}
}

func TestClient_DownloadDir(t *testing.T) {
	c, err := NewClient(testUrl, testUser, testPass, TimeoutLow)

	if err != nil {
		t.Fatal(err)
	}

	assert.IsType(t, &Client{}, c)

	t.Run("NonRecursive", func(t *testing.T) {
		tempDir := filepath.Join(os.TempDir(), rnd.UUID())

		if err = os.RemoveAll(tempDir); err != nil {
			t.Fatal(err)
		}

		if errs := c.DownloadDir("Photos", tempDir, false, false); len(errs) > 0 {
			t.Fatal(errs)
		}

		if err = os.RemoveAll(tempDir); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("Recursive", func(t *testing.T) {
		tempDir := filepath.Join(os.TempDir(), rnd.UUID())

		if err = os.RemoveAll(tempDir); err != nil {
			t.Fatal(err)
		}

		if errs := c.DownloadDir("Photos", tempDir, true, false); len(errs) > 0 {
			t.Fatal(errs)
		}

		if err = os.RemoveAll(tempDir); err != nil {
			t.Fatal(err)
		}
	})
}

package fs

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"os/user"
	"path/filepath"
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

func TestExists(t *testing.T) {
	assert.True(t, Exists("./testdata"))
	assert.True(t, Exists("./testdata/"))
	assert.True(t, Exists("./testdata/test.jpg"))
	assert.True(t, Exists("./testdata/test.jpg"))
	assert.True(t, Exists("./testdata/empty.jpg"))
	assert.False(t, Exists("./foo.jpg"))
	assert.False(t, Exists(""))
}

func TestFileExists(t *testing.T) {
	assert.False(t, FileExists("./testdata"))
	assert.False(t, FileExists("./testdata/"))
	assert.True(t, FileExists("./testdata/test.jpg"))
	assert.True(t, FileExists("./testdata/test.jpg"))
	assert.True(t, FileExists("./testdata/empty.jpg"))
	assert.False(t, FileExists("./foo.jpg"))
	assert.False(t, FileExists(""))
}

func TestFileExistsNotEmpty(t *testing.T) {
	assert.False(t, FileExistsNotEmpty("./testdata"))
	assert.False(t, FileExistsNotEmpty("./testdata/"))
	assert.True(t, FileExistsNotEmpty("./testdata/test.jpg"))
	assert.True(t, FileExistsNotEmpty("./testdata/test.jpg"))
	assert.False(t, FileExistsNotEmpty("./testdata/empty.jpg"))
	assert.False(t, FileExistsNotEmpty("./foo.jpg"))
	assert.False(t, FileExistsNotEmpty(""))
}

func TestFileExistsIsEmpty(t *testing.T) {
	assert.False(t, FileExistsIsEmpty("./testdata"))
	assert.False(t, FileExistsIsEmpty("./testdata/"))
	assert.False(t, FileExistsIsEmpty("./testdata/test.jpg"))
	assert.False(t, FileExistsIsEmpty("./testdata/test.jpg"))
	assert.True(t, FileExistsIsEmpty("./testdata/empty.jpg"))
	assert.False(t, FileExistsIsEmpty("./foo.jpg"))
	assert.False(t, FileExistsIsEmpty(""))
}

func TestFileSize(t *testing.T) {
	assert.Equal(t, 10990, int(FileSize("./testdata/test.jpg")))
	assert.Equal(t, 10990, int(FileSize("./testdata/test.jpg")))
	assert.Equal(t, 0, int(FileSize("./testdata/empty.jpg")))
	assert.Equal(t, -1, int(FileSize("./foo.jpg")))
	assert.Equal(t, -1, int(FileSize("")))
}

func TestPathExists(t *testing.T) {
	assert.True(t, PathExists("./testdata"))
	assert.False(t, PathExists("./testdata/test.jpg"))
	assert.False(t, PathExists("./testdata3ggdtgdg"))
	assert.False(t, PathExists(""))
}

func TestDeviceExists(t *testing.T) {
	assert.True(t, DeviceExists("/dev/null"))
	DeviceExists("/dev/nvidia0")
	assert.False(t, DeviceExists(""))
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

func TestExpandedFilename(t *testing.T) {
	t.Run("TestJpg", func(t *testing.T) {
		filename := Abs("./testdata/test.jpg")
		assert.Contains(t, filename, "/testdata/test.jpg")
		assert.IsType(t, "", filename)
	})
	t.Run("EmptyFilename", func(t *testing.T) {
		filename := Abs("")
		assert.Equal(t, "", filename)
		assert.IsType(t, "", filename)
	})
	t.Run("InFilename", func(t *testing.T) {
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
		if err := os.Mkdir("./testdata/emptyDir", 0o750); err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll("./testdata/emptyDir")
		assert.Equal(t, true, DirIsEmpty("./testdata/emptyDir"))
	})
}

func TestSocketExists(t *testing.T) {
	dir := t.TempDir()
	sock := filepath.Join(dir, "test.sock")

	ln, err := net.Listen("unix", sock)
	if err != nil {
		t.Skipf("unix sockets not supported: %v", err)
	}
	defer func() { _ = ln.Close(); _ = os.Remove(sock) }()

	assert.True(t, SocketExists(sock))
	assert.False(t, SocketExists(filepath.Join(dir, "missing.sock")))
}

func TestDownload_SuccessAndErrors(t *testing.T) {
	// Serve known content
	tsOK := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("hello world"))
	}))
	defer tsOK.Close()

	// Serve a failure status
	tsFail := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "nope", http.StatusBadRequest)
	}))
	defer tsFail.Close()

	dir := t.TempDir()
	goodPath := filepath.Join(dir, "sub", "file.txt")
	badPath := "file.txt" // invalid path according to Download

	// Success
	err := Download(goodPath, tsOK.URL)
	assert.NoError(t, err)
	b, rerr := os.ReadFile(goodPath) //nolint:gosec // test helper reads temp file
	assert.NoError(t, rerr)
	assert.Equal(t, "hello world", string(b))

	// Invalid target path
	err = Download(badPath, tsOK.URL)
	assert.Error(t, err)

	// Server error status
	anotherPath := filepath.Join(dir, "b", "x.txt")
	err = Download(anotherPath, tsFail.URL)
	assert.Error(t, err)
}

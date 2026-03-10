package webdav

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/rnd"
)

const (
	testUrl  = "http://dummy-webdav/"
	testUser = "admin"
	testPass = "photoprism"
)

// testWebDAVEntry represents a directory entry returned by the test WebDAV server.
type testWebDAVEntry struct {
	Href string
	Dir  bool
	Size int64
}

// testWebDAVServerOptions controls quirks exposed by the local WebDAV fixture.
type testWebDAVServerOptions struct {
	rejectInfinity    bool
	duplicateDepthOne bool
}

// testWebDAVTree defines the depth-1 PROPFIND responses returned by the test server.
var testWebDAVTree = map[string][]testWebDAVEntry{
	"/": {
		{Href: "/", Dir: true},
		{Href: "/.locks/", Dir: true},
		{Href: "/.partial-upload", Size: 1},
		{Href: "/Photos/", Dir: true},
		{Href: "/Shared/", Dir: true},
	},
	"/.locks": {
		{Href: "/.locks/", Dir: true},
		{Href: "/.locks/upload.tmp", Size: 2},
	},
	"/Photos": {
		{Href: "/Photos/", Dir: true},
		{Href: "/Photos/.upload.part", Size: 3},
		{Href: "/Photos/.staging/", Dir: true},
		{Href: "/Photos/cover.jpg", Size: 4},
		{Href: "/Photos/2020/", Dir: true},
		{Href: "/Photos/2021/", Dir: true},
	},
	"/Photos/.staging": {
		{Href: "/Photos/.staging/", Dir: true},
		{Href: "/Photos/.staging/incomplete.jpg", Size: 5},
	},
	"/Photos/2020": {
		{Href: "/Photos/2020/", Dir: true},
		{Href: "/Photos/2020/03/", Dir: true},
	},
	"/Photos/2020/03": {
		{Href: "/Photos/2020/03/", Dir: true},
	},
	"/Photos/2021": {
		{Href: "/Photos/2021/", Dir: true},
	},
	"/Shared": {
		{Href: "/Shared/", Dir: true},
	},
}

// newTestWebDAVServer returns a minimal WebDAV endpoint for recursive listing tests.
func newTestWebDAVServer(rejectInfinity bool) *httptest.Server {
	return newTestWebDAVServerWithOptions(testWebDAVServerOptions{rejectInfinity: rejectInfinity})
}

// newTestWebDAVServerWithOptions returns a minimal WebDAV endpoint with optional quirks for regression coverage.
func newTestWebDAVServerWithOptions(options testWebDAVServerOptions) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PROPFIND" {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		requestPath := testWebDAVCollectionPath(r.URL.Path)
		depth := strings.ToLower(strings.TrimSpace(r.Header.Get("Depth")))

		if depth == "" {
			depth = "infinity"
		}

		if depth == "infinity" && options.rejectInfinity {
			http.Error(w, `webdav: Depth: infinity is not supported`, http.StatusBadRequest)
			return
		}

		var entries []testWebDAVEntry

		switch depth {
		case "1":
			entries = testWebDAVDepthOneEntries(requestPath)

			if options.duplicateDepthOne && requestPath == "/Photos" {
				entries = append(entries,
					testWebDAVEntry{Href: "/Photos/2020/", Dir: true},
					testWebDAVEntry{Href: "/Photos/cover.jpg", Size: 4},
				)
			}
		case "infinity":
			entries = testWebDAVDepthInfinityEntries(requestPath)
		default:
			http.Error(w, "unsupported depth", http.StatusBadRequest)
			return
		}

		if len(entries) == 0 {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/xml; charset=utf-8")
		w.WriteHeader(http.StatusMultiStatus)
		//nolint:gosec // test fixture emits locally generated WebDAV XML only
		_, _ = w.Write([]byte(testWebDAVMultiStatus(entries)))
	}))
}

// newSlowWebDAVServer returns a WebDAV-like server that responds after a fixed delay.
func newSlowWebDAVServer(delay time.Duration, slowHeaders bool) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "MKCOL":
			time.Sleep(delay)
			w.WriteHeader(http.StatusCreated)
		case http.MethodDelete:
			time.Sleep(delay)
			w.WriteHeader(http.StatusNoContent)
		case "PROPFIND":
			time.Sleep(delay)
			w.Header().Set("Content-Type", "application/xml; charset=utf-8")
			w.WriteHeader(http.StatusMultiStatus)
			_, _ = io.WriteString(w, testWebDAVMultiStatus(testWebDAVDepthOneEntries("/")))
		case http.MethodGet:
			if slowHeaders {
				time.Sleep(delay)
			}
			w.WriteHeader(http.StatusOK)
			_, _ = io.WriteString(w, "test")
		case http.MethodPut:
			_, _ = io.Copy(io.Discard, r.Body)
			_ = r.Body.Close()
			if slowHeaders {
				time.Sleep(delay)
			}
			w.WriteHeader(http.StatusCreated)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}))
}

// testWebDAVCollectionPath normalizes collection paths used by the test server.
func testWebDAVCollectionPath(value string) string {
	if value = path.Clean("/" + strings.TrimSpace(value)); value != "." {
		return value
	}

	return "/"
}

// testWebDAVDepthOneEntries returns the direct child entries for a collection path.
func testWebDAVDepthOneEntries(collectionPath string) []testWebDAVEntry {
	return testWebDAVTree[testWebDAVCollectionPath(collectionPath)]
}

// testWebDAVDepthInfinityEntries returns the full subtree for a collection path.
func testWebDAVDepthInfinityEntries(collectionPath string) (result []testWebDAVEntry) {
	queue := []string{testWebDAVCollectionPath(collectionPath)}
	traversed := map[string]bool{testWebDAVCollectionPath(collectionPath): true}
	seen := map[string]bool{}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		for _, entry := range testWebDAVDepthOneEntries(current) {
			entryPath := testWebDAVCollectionPath(entry.Href)

			if !seen[entryPath] {
				seen[entryPath] = true
				result = append(result, entry)
			}

			if entry.Dir && entryPath != current && !traversed[entryPath] {
				traversed[entryPath] = true
				queue = append(queue, entryPath)
			}
		}
	}

	return result
}

// testWebDAVMultiStatus renders a simple 207 Multi-Status XML document for the test server.
func testWebDAVMultiStatus(entries []testWebDAVEntry) string {
	var builder strings.Builder

	builder.WriteString(`<?xml version="1.0" encoding="utf-8"?>`)
	builder.WriteString(`<D:multistatus xmlns:D="DAV:">`)

	for _, entry := range entries {
		builder.WriteString(`<D:response>`)
		builder.WriteString(`<D:href>`)
		builder.WriteString(entry.Href)
		builder.WriteString(`</D:href>`)
		builder.WriteString(`<D:propstat><D:prop><D:resourcetype>`)

		if entry.Dir {
			builder.WriteString(`<D:collection/>`)
		}

		builder.WriteString(`</D:resourcetype>`)
		builder.WriteString(`<D:getlastmodified>Mon, 02 Jan 2006 15:04:05 GMT</D:getlastmodified>`)

		if !entry.Dir {
			builder.WriteString(`<D:getcontentlength>`)
			builder.WriteString(strconv.FormatInt(entry.Size, 10))
			builder.WriteString(`</D:getcontentlength>`)
		}

		builder.WriteString(`</D:prop><D:status>HTTP/1.1 200 OK</D:status></D:propstat>`)
		builder.WriteString(`</D:response>`)
	}

	builder.WriteString(`</D:multistatus>`)

	return builder.String()
}

func TestClientUrl(t *testing.T) {
	result, err := clientUrl(testUrl, testUser, testPass)

	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "http://admin:photoprism@dummy-webdav/", result.String())
}

func TestNewClient(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		c, err := NewClient(testUrl, testUser, testPass, TimeoutLow, "")

		if err != nil {
			t.Fatal(err)
		}

		assert.IsType(t, &Client{}, c)
	})
	t.Run("InvalidCIDR", func(t *testing.T) {
		c, err := NewClient(testUrl, testUser, testPass, TimeoutLow, "not-a-cidr")
		assert.Nil(t, c)
		assert.Error(t, err)
	})
}

func TestClient_Files(t *testing.T) {
	c, err := NewClient(testUrl, testUser, testPass, TimeoutLow, "")

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
	c, err := NewClient(testUrl, testUser, testPass, TimeoutLow, "")

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

func TestClient_DirectoriesDepthFallback(t *testing.T) {
	expected := []string{"/", "/Photos", "/Photos/2020", "/Photos/2020/03", "/Photos/2021", "/Shared"}

	t.Run("RecursiveFastPath", func(t *testing.T) {
		server := newTestWebDAVServer(false)
		t.Cleanup(server.Close)

		client, err := NewClient(server.URL+"/", "", "", TimeoutLow, "")
		require.NoError(t, err)

		dirs, err := client.Directories("", true, MaxRequestDuration)
		require.NoError(t, err)
		assert.ElementsMatch(t, expected, dirs.Abs())
	})
	t.Run("DepthOneFallback", func(t *testing.T) {
		server := newTestWebDAVServer(true)
		t.Cleanup(server.Close)

		client, err := NewClient(server.URL+"/", "", "", TimeoutLow, "")
		require.NoError(t, err)

		dirs, err := client.Directories("", true, MaxRequestDuration)
		require.NoError(t, err)
		assert.ElementsMatch(t, expected, dirs.Abs())
	})
}

func TestClient_DirectoriesDepthFallback_DeduplicatesEntries(t *testing.T) {
	server := newTestWebDAVServerWithOptions(testWebDAVServerOptions{
		rejectInfinity:    true,
		duplicateDepthOne: true,
	})
	t.Cleanup(server.Close)

	client, err := NewClient(server.URL+"/", "", "", TimeoutLow, "")
	require.NoError(t, err)

	dirs, err := client.Directories("Photos", true, MaxRequestDuration)
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"/Photos", "/Photos/2020", "/Photos/2020/03", "/Photos/2021"}, dirs.Abs())
}

func TestClient_FilesExcludeHiddenEntries(t *testing.T) {
	server := newTestWebDAVServer(false)
	t.Cleanup(server.Close)

	client, err := NewClient(server.URL+"/", "", "", TimeoutLow, "")
	require.NoError(t, err)

	t.Run("NonRecursive", func(t *testing.T) {
		files, err := client.Files("Photos", false)
		require.NoError(t, err)
		require.Len(t, files, 1)
		assert.Equal(t, "/Photos/cover.jpg", files[0].Abs)
	})
	t.Run("Recursive", func(t *testing.T) {
		files, err := client.Files("", true)
		require.NoError(t, err)

		paths := files.Abs()
		assert.Contains(t, paths, "/Photos/cover.jpg")
		assert.NotContains(t, paths, "/.partial-upload")
		assert.NotContains(t, paths, "/.locks/upload.tmp")
		assert.NotContains(t, paths, "/Photos/.upload.part")
		assert.NotContains(t, paths, "/Photos/.staging/incomplete.jpg")
	})
}

func TestClient_OperationTimeouts(t *testing.T) {
	server := newSlowWebDAVServer(200*time.Millisecond, false)
	t.Cleanup(server.Close)

	client, err := NewClient(server.URL+"/", "", "", TimeoutLow, "")
	require.NoError(t, err)
	client.timeout = 25 * time.Millisecond

	t.Run("Mkdir", func(t *testing.T) {
		err := client.Mkdir("slow-dir")
		require.Error(t, err)
		assert.True(t, strings.Contains(err.Error(), "Client.Timeout") || strings.Contains(err.Error(), "context deadline exceeded"))
	})
	t.Run("Files", func(t *testing.T) {
		_, err := client.Files("", false)
		require.Error(t, err)
		assert.True(t, strings.Contains(err.Error(), "Client.Timeout") || strings.Contains(err.Error(), "context deadline exceeded"))
	})
	t.Run("Delete", func(t *testing.T) {
		err := client.Delete("/slow.txt")
		require.Error(t, err)
		assert.True(t, strings.Contains(err.Error(), "Client.Timeout") || strings.Contains(err.Error(), "context deadline exceeded"))
	})
}

func TestClient_TransferOperationsIgnoreServiceTimeout(t *testing.T) {
	server := newSlowWebDAVServer(200*time.Millisecond, false)
	t.Cleanup(server.Close)

	client, err := NewClient(server.URL+"/", "", "", TimeoutLow, "")
	require.NoError(t, err)
	client.timeout = 25 * time.Millisecond

	t.Run("Upload", func(t *testing.T) {
		err := client.Upload(fs.Abs("testdata/example.jpg"), "slow-upload.jpg")
		require.NoError(t, err)
	})
	t.Run("Download", func(t *testing.T) {
		dest := filepath.Join(t.TempDir(), "slow.txt")
		err := client.Download("/slow.txt", dest, false)
		require.NoError(t, err)

		//nolint:gosec // test reads back a file created in t.TempDir()
		data, readErr := os.ReadFile(dest)
		require.NoError(t, readErr)
		assert.Equal(t, "test", string(data))
	})
}

func TestClient_TransferOperationsIgnoreSlowResponseHeaders(t *testing.T) {
	server := newSlowWebDAVServer(200*time.Millisecond, true)
	t.Cleanup(server.Close)

	client, err := NewClient(server.URL+"/", "", "", TimeoutLow, "")
	require.NoError(t, err)
	client.timeout = 25 * time.Millisecond

	t.Run("UploadResponseHeaders", func(t *testing.T) {
		err := client.Upload(fs.Abs("testdata/example.jpg"), "slow-upload.jpg")
		require.NoError(t, err)
	})
	t.Run("DownloadResponseHeaders", func(t *testing.T) {
		dest := filepath.Join(t.TempDir(), "slow.txt")
		err := client.Download("/slow.txt", dest, false)
		require.NoError(t, err)
		assert.True(t, fs.FileExists(dest))
	})
}

func TestClient_UploadAndDelete(t *testing.T) {
	c, err := NewClient(testUrl, testUser, testPass, TimeoutLow, "")

	if err != nil {
		t.Fatal(err)
	}

	assert.IsType(t, &Client{}, c)

	tempName := rnd.UUID() + fs.ExtJpeg

	if err := c.Upload(fs.Abs("testdata/example.jpg"), tempName); err != nil {
		t.Fatal(err)
	}

	if err := c.Delete(tempName); err != nil {
		t.Fatal(err)
	}
}

func TestClient_Download(t *testing.T) {
	c, err := NewClient(testUrl, testUser, testPass, TimeoutDefault, "")

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
	c, err := NewClient(testUrl, testUser, testPass, TimeoutLow, "")

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

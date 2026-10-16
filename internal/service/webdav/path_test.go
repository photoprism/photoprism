package webdav

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsUnsafePath(t *testing.T) {
	t.Run("Safe", func(t *testing.T) {
		assert.False(t, isUnsafePath("/Photos/cover.jpg"))
		assert.False(t, isUnsafePath("Photos/2020/03/img.jpg"))
		assert.False(t, isUnsafePath("/"))
		assert.False(t, isUnsafePath(""))
		assert.False(t, isUnsafePath("/a..b/c.jpg"))
	})
	t.Run("RootedInteriorTraversal", func(t *testing.T) {
		assert.True(t, isUnsafePath("/sub/../../../../outside/target.jpg"))
	})
	t.Run("LeadingTraversal", func(t *testing.T) {
		assert.True(t, isUnsafePath("../outside.jpg"))
	})
	t.Run("TrailingTraversal", func(t *testing.T) {
		assert.True(t, isUnsafePath("/Photos/.."))
	})
	t.Run("BackslashTraversal", func(t *testing.T) {
		assert.True(t, isUnsafePath(`\sub\..\..\outside\target.jpg`))
	})
}

func TestHasControl(t *testing.T) {
	t.Run("Control", func(t *testing.T) {
		for _, name := range []string{"\x00", "a\x01b", "a\nb", "a\rb", "a\tb", "a\x1fb", "a\x7fb"} {
			assert.True(t, hasControl(name), "%q", name)
		}
	})
	t.Run("None", func(t *testing.T) {
		for _, name := range []string{"", "/Photos/cover.jpg", "/Fotos/Älter 2024 幸福.jpg", "/صور/a\u200Db.jpg", "a\u0085b", "a\xffb"} {
			assert.False(t, hasControl(name), "%q", name)
		}
	})
}

func TestIsHiddenPath(t *testing.T) {
	assert.True(t, isHiddenPath("/.locks/upload.tmp"))
	assert.True(t, isHiddenPath("/Photos/.staging/incomplete.jpg"))
	assert.False(t, isHiddenPath("/Photos/cover.jpg"))

	// Other names that are hidden locally are not hidden remotely.
	for _, name := range []string{"/@Family/x.jpg", "/Photos/@eaDir/x.jpg", "/__MACOSX/x.jpg", "/Photos/_.x.jpg", "/$RECYCLE.BIN/x.jpg", "~home"} {
		assert.False(t, isHiddenPath(name), name)
	}
}

// TestClient_FilesRejectTraversalEntries verifies that files whose remote href
// contains a parent-directory segment are excluded from listings so they never
// become a local download destination.
func TestClient_FilesRejectTraversalEntries(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PROPFIND" {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		entries := []testWebDAVEntry{
			{Href: "/Photos/", Dir: true},
			{Href: "/Photos/cover.jpg", Size: 4},
			{Href: "/Photos/sub/../../../../outside/target.jpg", Size: 5},
		}

		w.Header().Set("Content-Type", "application/xml; charset=utf-8")
		w.WriteHeader(http.StatusMultiStatus)
		//nolint:gosec // test fixture emits locally generated WebDAV XML only
		_, _ = w.Write([]byte(testWebDAVMultiStatus(entries)))
	}))
	t.Cleanup(server.Close)

	client, err := NewClient(server.URL+"/", "", "", TimeoutLow, "")
	require.NoError(t, err)

	files, err := client.Files("Photos", false)
	require.NoError(t, err)

	paths := files.Abs()
	assert.Contains(t, paths, "/Photos/cover.jpg")

	for _, p := range paths {
		assert.NotContains(t, p, "..", "traversal entry must be excluded from listing")
		assert.NotContains(t, p, "outside", "traversal entry must be excluded from listing")
	}
}

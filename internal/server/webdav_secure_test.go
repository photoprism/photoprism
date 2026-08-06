package server

import (
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/http/header"
)

func TestJoinUnderBase(t *testing.T) {
	base := t.TempDir()
	t.Run("NormalJoin", func(t *testing.T) {
		out, err := joinUnderBase(base, "a/b/c.txt")
		assert.NoError(t, err)
		assert.Equal(t, filepath.Join(base, "a/b/c.txt"), out)
	})
	t.Run("BackslashNormalized", func(t *testing.T) {
		out, err := joinUnderBase(base, "a\\b\\c.txt")
		assert.NoError(t, err)
		assert.Equal(t, filepath.Join(base, "a/b/c.txt"), out)
	})
	t.Run("CleansInsideBase", func(t *testing.T) {
		out, err := joinUnderBase(base, "dir/../evil.txt")
		assert.NoError(t, err)
		assert.Equal(t, filepath.Join(base, "evil.txt"), out)
	})
	t.Run("EmptyRejected", func(t *testing.T) {
		_, err := joinUnderBase(base, "")
		assert.Error(t, err)
	})
	t.Run("AbsoluteRejected", func(t *testing.T) {
		_, err := joinUnderBase(base, "/etc/passwd")
		assert.Error(t, err)
	})
	t.Run("RootedBackslashRejected", func(t *testing.T) {
		// "\rooted\path.txt" normalizes to "/rooted/path.txt", caught as absolute.
		_, err := joinUnderBase(base, "\\rooted\\path.txt")
		assert.Error(t, err)
	})
	t.Run("ParentTraversalRejected", func(t *testing.T) {
		_, err := joinUnderBase(base, "../../etc/passwd")
		assert.Error(t, err)
	})
	t.Run("BackslashTraversalRejected", func(t *testing.T) {
		_, err := joinUnderBase(base, "..\\..\\etc\\passwd")
		assert.Error(t, err)
	})
	t.Run("DriveLetterForwardSlashRejected", func(t *testing.T) {
		_, err := joinUnderBase(base, "C:/drive.txt")
		assert.Error(t, err)
	})
	t.Run("DriveLetterBackslashRejected", func(t *testing.T) {
		_, err := joinUnderBase(base, "C:\\drive.txt")
		assert.Error(t, err)
	})
	t.Run("LowercaseDriveLetterRejected", func(t *testing.T) {
		_, err := joinUnderBase(base, "c:/drive.txt")
		assert.Error(t, err)
	})
	t.Run("VolumeNameOnlyRejected", func(t *testing.T) {
		_, err := joinUnderBase(base, "C:")
		assert.Error(t, err)
	})
}

func TestWebDAVFileName_PathTraversalRejected(t *testing.T) {
	dir := t.TempDir()
	// Create a legitimate file inside base to ensure happy-path works later.
	insideFile := filepath.Join(dir, "ok.txt")
	assert.NoError(t, fs.WriteString(insideFile, "ok"))

	conf := newWebDAVTestConfig(t)
	conf.Options().OriginalsPath = dir

	r := gin.New()
	grp := r.Group(conf.BaseUri(WebDAVOriginals))

	// Attempt traversal to outside path.
	req := &http.Request{Method: header.MethodPut}
	req.URL = &url.URL{Path: conf.BaseUri(WebDAVOriginals) + "/../../etc/passwd"}
	got := WebDAVFileName(req, grp, conf)
	assert.Equal(t, "", got, "should reject traversal")

	// Happy path: file under base resolves and exists.
	req2 := &http.Request{Method: header.MethodPut}
	req2.URL = &url.URL{Path: conf.BaseUri(WebDAVOriginals) + "/ok.txt"}
	got = WebDAVFileName(req2, grp, conf)
	assert.Equal(t, insideFile, got)
}

func TestWebDAVFileName_DriveLetterRejected(t *testing.T) {
	conf := newWebDAVTestConfig(t)
	r := gin.New()
	grp := r.Group(conf.BaseUri(WebDAVOriginals))
	// A Windows drive-letter prefix must not resolve to an in-tree name.
	req := &http.Request{Method: header.MethodPut}
	req.URL = &url.URL{Path: conf.BaseUri(WebDAVOriginals) + "/C:/drive.txt"}
	got := WebDAVFileName(req, grp, conf)
	assert.Equal(t, "", got, "should reject drive-letter path")
}

func TestWebDAVFileName_MethodNotPut(t *testing.T) {
	conf := newWebDAVTestConfig(t)
	r := gin.New()
	grp := r.Group(conf.BaseUri(WebDAVOriginals))
	req := &http.Request{Method: header.MethodGet}
	req.URL = &url.URL{Path: conf.BaseUri(WebDAVOriginals) + "/anything.jpg"}
	got := WebDAVFileName(req, grp, conf)
	assert.Equal(t, "", got)
}

func TestWebDAVFileName_ImportBasePath(t *testing.T) {
	conf := newWebDAVTestConfig(t)
	r := gin.New()
	grp := r.Group(conf.BaseUri(WebDAVImport))
	// create a real file under import
	file := filepath.Join(conf.ImportPath(), "in.jpg")
	assert.NoError(t, fs.MkdirAll(filepath.Dir(file)))
	assert.NoError(t, fs.WriteString(file, "x"))
	req := &http.Request{Method: header.MethodPut}
	req.URL = &url.URL{Path: conf.BaseUri(WebDAVImport) + "/in.jpg"}
	got := WebDAVFileName(req, grp, conf)
	assert.Equal(t, file, got)
}

func TestWebDAVSetFileMtime_FutureIgnored(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "a.txt")
	assert.NoError(t, fs.WriteString(file, "x"))
	before, _ := os.Stat(file)
	future := time.Now().Add(2 * time.Hour).Unix()
	WebDAVSetFileMtime(file, future)
	after, _ := os.Stat(file)
	assert.Equal(t, before.ModTime().Unix(), after.ModTime().Unix())
}

func newWebDAVTestConfig(t *testing.T) *config.Config {
	return config.NewMinimalTestConfig(t.TempDir())
}

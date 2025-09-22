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
)

func TestJoinUnderBase(t *testing.T) {
	base := t.TempDir()
	// Normal join
	out, err := joinUnderBase(base, "a/b/c.txt")
	assert.NoError(t, err)
	assert.Equal(t, filepath.Join(base, "a/b/c.txt"), out)
	// Absolute rejected
	_, err = joinUnderBase(base, "/etc/passwd")
	assert.Error(t, err)
	// Parent traversal rejected
	_, err = joinUnderBase(base, "../../etc/passwd")
	assert.Error(t, err)
}

func TestWebDAVFileName_PathTraversalRejected(t *testing.T) {
	dir := t.TempDir()
	// Create a legitimate file inside base to ensure happy-path works later.
	insideFile := filepath.Join(dir, "ok.txt")
	assert.NoError(t, fs.WriteString(insideFile, "ok"))

	conf := config.NewTestConfig("server-webdav")
	conf.Options().OriginalsPath = dir

	r := gin.New()
	grp := r.Group(conf.BaseUri(WebDAVOriginals))

	// Attempt traversal to outside path.
	req := &http.Request{Method: http.MethodPut}
	req.URL = &url.URL{Path: conf.BaseUri(WebDAVOriginals) + "/../../etc/passwd"}
	got := WebDAVFileName(req, grp, conf)
	assert.Equal(t, "", got, "should reject traversal")

	// Happy path: file under base resolves and exists.
	req2 := &http.Request{Method: http.MethodPut}
	req2.URL = &url.URL{Path: conf.BaseUri(WebDAVOriginals) + "/ok.txt"}
	got = WebDAVFileName(req2, grp, conf)
	assert.Equal(t, insideFile, got)
}

func TestWebDAVFileName_MethodNotPut(t *testing.T) {
	conf := config.NewTestConfig("server-webdav")
	r := gin.New()
	grp := r.Group(conf.BaseUri(WebDAVOriginals))
	req := &http.Request{Method: http.MethodGet}
	req.URL = &url.URL{Path: conf.BaseUri(WebDAVOriginals) + "/anything.jpg"}
	got := WebDAVFileName(req, grp, conf)
	assert.Equal(t, "", got)
}

func TestWebDAVFileName_ImportBasePath(t *testing.T) {
	conf := config.NewTestConfig("server-webdav")
	r := gin.New()
	grp := r.Group(conf.BaseUri(WebDAVImport))
	// create a real file under import
	file := filepath.Join(conf.ImportPath(), "in.jpg")
	assert.NoError(t, fs.MkdirAll(filepath.Dir(file)))
	assert.NoError(t, fs.WriteString(file, "x"))
	req := &http.Request{Method: http.MethodPut}
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

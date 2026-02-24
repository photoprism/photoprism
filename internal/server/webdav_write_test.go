package server

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/http/header"
)

func setupWebDAVRouter(conf *config.Config) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	grp := r.Group(conf.BaseUri(WebDAVOriginals), WebDAVAuth(conf))
	WebDAV(conf.OriginalsPath(), grp, conf)
	return r
}

func authBearer(req *http.Request) {
	sess := entity.SessionFixtures.Get("alice_token_webdav")
	header.SetAuthorization(req, sess.AuthToken())
}

func authBasic(req *http.Request) {
	sess := entity.SessionFixtures.Get("alice_token_webdav")
	basic := fmt.Appendf(nil, "alice:%s", sess.AuthToken())
	req.Header.Set(header.Auth, fmt.Sprintf("%s %s", header.AuthBasic, base64.StdEncoding.EncodeToString(basic)))
}

func TestWebDAVWrite_MKCOL_PUT(t *testing.T) {
	conf := newWebDAVTestConfig(t)
	if err := conf.CreateDirectories(); err != nil {
		t.Fatalf("failed to create test directories: %v", err)
	}
	r := setupWebDAVRouter(conf)

	// MKCOL
	w := httptest.NewRecorder()
	req := httptest.NewRequest(header.MethodMkcol, conf.BaseUri(WebDAVOriginals)+"/wdvdir", nil)
	authBearer(req)
	r.ServeHTTP(w, req)
	assert.InDelta(t, 201, w.Code, 1) // Created
	// PUT file
	w = httptest.NewRecorder()
	req = httptest.NewRequest(header.MethodPut, conf.BaseUri(WebDAVOriginals)+"/wdvdir/hello.txt", bytes.NewBufferString("hello"))
	authBearer(req)
	r.ServeHTTP(w, req)
	assert.InDelta(t, 201, w.Code, 1)
	// file exists
	path := filepath.Join(conf.OriginalsPath(), "wdvdir", "hello.txt")
	// #nosec G304 -- test reads file created under controlled temp directory.
	b, err := os.ReadFile(path)
	assert.NoError(t, err)
	assert.Equal(t, "hello", string(b))
}

func TestWebDAV_NoTrailingSlashRedirectOnBasePath(t *testing.T) {
	testCases := []struct {
		name    string
		siteURL string
	}{
		{name: "DefaultBasePath", siteURL: "http://localhost:2342/"},
		{name: "PrefixedBasePath", siteURL: "https://app.localssl.dev/i/pro-1/"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			conf := newWebDAVTestConfig(t)
			conf.Options().SiteUrl = tc.siteURL
			if err := conf.CreateDirectories(); err != nil {
				t.Fatalf("failed to create test directories: %v", err)
			}

			r := setupWebDAVRouter(conf)
			basePath := conf.BaseUri(WebDAVOriginals)

			for _, method := range []string{header.MethodOptions, header.MethodPropfind} {
				w := httptest.NewRecorder()
				req := httptest.NewRequest(method, basePath, nil)
				if method == header.MethodPropfind {
					req.Header.Set("Depth", "0")
				}
				authBasic(req)
				r.ServeHTTP(w, req)

				if w.Code >= 300 && w.Code < 400 {
					t.Fatalf("expected no redirect for %s %s, got %d (%s)", method, basePath, w.Code, w.Header().Get("Location"))
				}
			}
		})
	}
}

func TestWebDAVWrite_MOVE_COPY(t *testing.T) {
	conf := newWebDAVTestConfig(t)
	if err := conf.CreateDirectories(); err != nil {
		t.Fatalf("failed to create test directories: %v", err)
	}
	r := setupWebDAVRouter(conf)

	// Ensure source and destination directories via MKCOL
	w := httptest.NewRecorder()
	req := httptest.NewRequest(header.MethodMkcol, conf.BaseUri(WebDAVOriginals)+"/src", nil)
	authBasic(req)
	r.ServeHTTP(w, req)
	assert.InDelta(t, 201, w.Code, 1)
	w = httptest.NewRecorder()
	req = httptest.NewRequest(header.MethodMkcol, conf.BaseUri(WebDAVOriginals)+"/dst", nil)
	authBasic(req)
	r.ServeHTTP(w, req)
	assert.InDelta(t, 201, w.Code, 1)
	// Create source file via PUT
	w = httptest.NewRecorder()
	req = httptest.NewRequest(header.MethodPut, conf.BaseUri(WebDAVOriginals)+"/src/a.txt", bytes.NewBufferString("A"))
	authBasic(req)
	r.ServeHTTP(w, req)
	assert.InDelta(t, 201, w.Code, 1)

	// MOVE /originals/src/a.txt -> /originals/dst/b.txt
	w = httptest.NewRecorder()
	req = httptest.NewRequest(header.MethodMove, conf.BaseUri(WebDAVOriginals)+"/src/a.txt", nil)
	req.Header.Set("Destination", conf.BaseUri(WebDAVOriginals)+"/dst/b.txt")
	authBasic(req)
	r.ServeHTTP(w, req)
	assert.InDelta(t, 201, w.Code, 1)
	// Verify moved
	assert.NoFileExists(t, filepath.Join(conf.OriginalsPath(), "src", "a.txt"))
	assert.FileExists(t, filepath.Join(conf.OriginalsPath(), "dst", "b.txt"))

	// COPY /originals/dst/b.txt -> /originals/dst/c.txt
	w = httptest.NewRecorder()
	req = httptest.NewRequest(header.MethodCopy, conf.BaseUri(WebDAVOriginals)+"/dst/b.txt", nil)
	req.Header.Set("Destination", conf.BaseUri(WebDAVOriginals)+"/dst/c.txt")
	authBasic(req)
	r.ServeHTTP(w, req)
	assert.InDelta(t, 201, w.Code, 1)
	// Verify copy
	assert.FileExists(t, filepath.Join(conf.OriginalsPath(), "dst", "b.txt"))
	assert.FileExists(t, filepath.Join(conf.OriginalsPath(), "dst", "c.txt"))
}

func TestWebDAVWrite_OverwriteSemantics(t *testing.T) {
	conf := newWebDAVTestConfig(t)
	if err := conf.CreateDirectories(); err != nil {
		t.Fatalf("failed to create test directories: %v", err)
	}
	r := setupWebDAVRouter(conf)

	// Prepare src and dst
	_ = os.MkdirAll(filepath.Join(conf.OriginalsPath(), "src"), 0o700)
	_ = os.MkdirAll(filepath.Join(conf.OriginalsPath(), "dst"), 0o700)
	_ = os.WriteFile(filepath.Join(conf.OriginalsPath(), "src", "f.txt"), []byte("NEW"), 0o600)
	_ = os.WriteFile(filepath.Join(conf.OriginalsPath(), "dst", "f.txt"), []byte("OLD"), 0o600)

	// COPY with Overwrite: F -> should not overwrite existing
	w := httptest.NewRecorder()
	req := httptest.NewRequest(header.MethodCopy, conf.BaseUri(WebDAVOriginals)+"/src/f.txt", nil)
	req.Header.Set("Destination", conf.BaseUri(WebDAVOriginals)+"/dst/f.txt")
	req.Header.Set("Overwrite", "F")
	authBasic(req)
	r.ServeHTTP(w, req)
	// Expect not successful (commonly 412 Precondition Failed)
	if w.Code == 201 || w.Code == 204 {
		t.Fatalf("expected failure when Overwrite=F, got %d", w.Code)
	}
	// Content remains OLD
	b, _ := os.ReadFile(filepath.Join(conf.OriginalsPath(), "dst", "f.txt"))
	assert.Equal(t, "OLD", string(b))

	// COPY with Overwrite: T -> must overwrite
	w = httptest.NewRecorder()
	req = httptest.NewRequest(header.MethodCopy, conf.BaseUri(WebDAVOriginals)+"/src/f.txt", nil)
	req.Header.Set("Destination", conf.BaseUri(WebDAVOriginals)+"/dst/f.txt")
	req.Header.Set("Overwrite", "T")
	authBasic(req)
	r.ServeHTTP(w, req)
	// Success (201/204 acceptable)
	if w.Code != http.StatusCreated && w.Code != http.StatusNoContent {
		t.Fatalf("expected success for Overwrite=T, got %d", w.Code)
	}
	b, _ = os.ReadFile(filepath.Join(conf.OriginalsPath(), "dst", "f.txt"))
	assert.Equal(t, "NEW", string(b))

	// MOVE with Overwrite: F to existing file -> expect failure
	_ = os.WriteFile(filepath.Join(conf.OriginalsPath(), "src", "g.txt"), []byte("GNEW"), 0o600)
	_ = os.WriteFile(filepath.Join(conf.OriginalsPath(), "dst", "g.txt"), []byte("GOLD"), 0o600)
	w = httptest.NewRecorder()
	req = httptest.NewRequest(header.MethodMove, conf.BaseUri(WebDAVOriginals)+"/src/g.txt", nil)
	req.Header.Set("Destination", conf.BaseUri(WebDAVOriginals)+"/dst/g.txt")
	req.Header.Set("Overwrite", "F")
	authBasic(req)
	r.ServeHTTP(w, req)
	if w.Code == 201 || w.Code == 204 {
		t.Fatalf("expected failure when Overwrite=F for MOVE, got %d", w.Code)
	}
	// MOVE with Overwrite: T -> overwrites and removes source
	w = httptest.NewRecorder()
	req = httptest.NewRequest(header.MethodMove, conf.BaseUri(WebDAVOriginals)+"/src/g.txt", nil)
	req.Header.Set("Destination", conf.BaseUri(WebDAVOriginals)+"/dst/g.txt")
	req.Header.Set("Overwrite", "T")
	authBasic(req)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated && w.Code != http.StatusNoContent {
		t.Fatalf("expected success for MOVE Overwrite=T, got %d", w.Code)
	}
	assert.NoFileExists(t, filepath.Join(conf.OriginalsPath(), "src", "g.txt"))
	gb, _ := os.ReadFile(filepath.Join(conf.OriginalsPath(), "dst", "g.txt"))
	assert.Equal(t, "GNEW", string(gb))
}

func TestWebDAVWrite_MoveMissingDestination(t *testing.T) {
	conf := newWebDAVTestConfig(t)
	if err := conf.CreateDirectories(); err != nil {
		t.Fatalf("failed to create test directories: %v", err)
	}
	r := setupWebDAVRouter(conf)
	// Ensure src exists
	_ = os.MkdirAll(filepath.Join(conf.OriginalsPath(), "mv"), 0o700)
	_ = os.WriteFile(filepath.Join(conf.OriginalsPath(), "mv", "file.txt"), []byte("X"), 0o600)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(header.MethodMove, conf.BaseUri(WebDAVOriginals)+"/mv/file.txt", nil)
	// no Destination header
	authBasic(req)
	r.ServeHTTP(w, req)
	// Expect failure (not 201/204)
	if w.Code == http.StatusCreated || w.Code == http.StatusNoContent {
		t.Fatalf("expected failure when Destination header missing, got %d", w.Code)
	}
	// Source remains
	assert.FileExists(t, filepath.Join(conf.OriginalsPath(), "mv", "file.txt"))
}

func TestWebDAVWrite_CopyInvalidDestinationPrefix(t *testing.T) {
	conf := newWebDAVTestConfig(t)
	if err := conf.CreateDirectories(); err != nil {
		t.Fatalf("failed to create test directories: %v", err)
	}
	r := setupWebDAVRouter(conf)
	// Ensure src exists
	_ = os.MkdirAll(filepath.Join(conf.OriginalsPath(), "cp"), 0o700)
	_ = os.WriteFile(filepath.Join(conf.OriginalsPath(), "cp", "a.txt"), []byte("A"), 0o600)

	// COPY to a destination outside the handler prefix
	w := httptest.NewRecorder()
	req := httptest.NewRequest(header.MethodCopy, conf.BaseUri(WebDAVOriginals)+"/cp/a.txt", nil)
	req.Header.Set("Destination", "/notwebdav/d.txt")
	authBasic(req)
	r.ServeHTTP(w, req)
	// Expect failure
	if w.Code == http.StatusCreated || w.Code == http.StatusNoContent {
		t.Fatalf("expected failure for invalid Destination prefix, got %d", w.Code)
	}
	// Destination not created
	assert.NoFileExists(t, filepath.Join(conf.OriginalsPath(), "notwebdav", "d.txt"))
}

func TestWebDAVWrite_MoveNonExistentSource(t *testing.T) {
	conf := newWebDAVTestConfig(t)
	if err := conf.CreateDirectories(); err != nil {
		t.Fatalf("failed to create test directories: %v", err)
	}
	r := setupWebDAVRouter(conf)
	// Ensure destination dir exists
	_ = os.MkdirAll(filepath.Join(conf.OriginalsPath(), "dst2"), 0o700)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(header.MethodMove, conf.BaseUri(WebDAVOriginals)+"/nosuch/file.txt", nil)
	req.Header.Set("Destination", conf.BaseUri(WebDAVOriginals)+"/dst2/file.txt")
	authBasic(req)
	r.ServeHTTP(w, req)
	// Expect failure (e.g., 404)
	if w.Code == http.StatusCreated || w.Code == http.StatusNoContent {
		t.Fatalf("expected failure moving non-existent source, got %d", w.Code)
	}
	assert.NoFileExists(t, filepath.Join(conf.OriginalsPath(), "dst2", "file.txt"))
}

func TestWebDAVWrite_CopyTraversalDestination(t *testing.T) {
	conf := newWebDAVTestConfig(t)
	if err := conf.CreateDirectories(); err != nil {
		t.Fatalf("failed to create test directories: %v", err)
	}
	r := setupWebDAVRouter(conf)

	// Create source file via PUT
	_ = os.MkdirAll(filepath.Join(conf.OriginalsPath(), "travsrc"), 0o700)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(header.MethodPut, conf.BaseUri(WebDAVOriginals)+"/travsrc/a.txt", bytes.NewBufferString("A"))
	authBasic(req)
	r.ServeHTTP(w, req)
	assert.InDelta(t, 201, w.Code, 1)

	// Attempt COPY with traversal in Destination
	w = httptest.NewRecorder()
	req = httptest.NewRequest(header.MethodCopy, conf.BaseUri(WebDAVOriginals)+"/travsrc/a.txt", nil)
	req.Header.Set("Destination", conf.BaseUri(WebDAVOriginals)+"/../evil.txt")
	authBasic(req)
	r.ServeHTTP(w, req)
	// Expect success with sanitized destination inside base
	if w.Code != http.StatusCreated && w.Code != http.StatusNoContent {
		t.Fatalf("expected success (sanitized), got %d", w.Code)
	}
	// Not created above originals; created as /originals/evil.txt
	parent := filepath.Dir(conf.OriginalsPath())
	assert.NoFileExists(t, filepath.Join(parent, "evil.txt"))
	assert.FileExists(t, filepath.Join(conf.OriginalsPath(), "evil.txt"))
}

func TestWebDAVWrite_MoveTraversalDestination(t *testing.T) {
	conf := newWebDAVTestConfig(t)
	if err := conf.CreateDirectories(); err != nil {
		t.Fatalf("failed to create test directories: %v", err)
	}
	r := setupWebDAVRouter(conf)

	// Create source file via PUT
	_ = os.MkdirAll(filepath.Join(conf.OriginalsPath(), "travsrc2"), 0o700)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(header.MethodPut, conf.BaseUri(WebDAVOriginals)+"/travsrc2/a.txt", bytes.NewBufferString("A"))
	authBasic(req)
	r.ServeHTTP(w, req)
	assert.InDelta(t, 201, w.Code, 1)

	// Attempt MOVE with traversal in Destination
	w = httptest.NewRecorder()
	req = httptest.NewRequest(header.MethodMove, conf.BaseUri(WebDAVOriginals)+"/travsrc2/a.txt", nil)
	req.Header.Set("Destination", conf.BaseUri(WebDAVOriginals)+"/../evil2.txt")
	authBasic(req)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated && w.Code != http.StatusNoContent {
		t.Fatalf("expected success (sanitized) for MOVE, got %d", w.Code)
	}
	// Source removed; destination created inside base, not outside
	assert.NoFileExists(t, filepath.Join(conf.OriginalsPath(), "travsrc2", "a.txt"))
	parent := filepath.Dir(conf.OriginalsPath())
	assert.NoFileExists(t, filepath.Join(parent, "evil2.txt"))
	assert.FileExists(t, filepath.Join(conf.OriginalsPath(), "evil2.txt"))
}

func TestWebDAVWrite_ReadOnlyMethodNotAllowed(t *testing.T) {
	conf := newWebDAVTestConfig(t)
	conf.Options().ReadOnly = true
	if err := conf.CreateDirectories(); err != nil {
		t.Fatalf("failed to create test directories: %v", err)
	}
	r := setupWebDAVRouter(conf)

	for _, method := range []string{header.MethodMkcol, header.MethodLock, header.MethodUnlock} {
		t.Run(method, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest(method, conf.BaseUri(WebDAVOriginals)+"/ro", nil)
			authBearer(req)
			r.ServeHTTP(w, req)
			assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
		})
	}
}

func TestWebDAVWrite_PatchMethodNotAllowed(t *testing.T) {
	conf := newWebDAVTestConfig(t)
	if err := conf.CreateDirectories(); err != nil {
		t.Fatalf("failed to create test directories: %v", err)
	}
	r := setupWebDAVRouter(conf)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(header.MethodPatch, conf.BaseUri(WebDAVOriginals)+"/wdvdir/hello.txt", bytes.NewBufferString("{}"))
	authBearer(req)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestWebDAVWrite_PostNotBlockedByReadOnly(t *testing.T) {
	readWriteConf := newWebDAVTestConfig(t)
	if err := readWriteConf.CreateDirectories(); err != nil {
		t.Fatalf("failed to create test directories: %v", err)
	}
	readWriteRouter := setupWebDAVRouter(readWriteConf)

	readOnlyConf := newWebDAVTestConfig(t)
	readOnlyConf.Options().ReadOnly = true
	if err := readOnlyConf.CreateDirectories(); err != nil {
		t.Fatalf("failed to create test directories: %v", err)
	}
	readOnlyRouter := setupWebDAVRouter(readOnlyConf)

	readWriteResp := httptest.NewRecorder()
	readWriteReq := httptest.NewRequest(header.MethodPost, readWriteConf.BaseUri(WebDAVOriginals)+"/post-test", nil)
	authBearer(readWriteReq)
	readWriteRouter.ServeHTTP(readWriteResp, readWriteReq)

	readOnlyResp := httptest.NewRecorder()
	readOnlyReq := httptest.NewRequest(header.MethodPost, readOnlyConf.BaseUri(WebDAVOriginals)+"/post-test", nil)
	authBearer(readOnlyReq)
	readOnlyRouter.ServeHTTP(readOnlyResp, readOnlyReq)

	assert.Equal(t, readWriteResp.Code, readOnlyResp.Code)
	assert.NotEqual(t, http.StatusForbidden, readOnlyResp.Code)
}

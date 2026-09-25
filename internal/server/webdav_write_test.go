package server

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/pkg/fs"
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

// logCapture is a logrus hook that records emitted entries for assertions.
type logCapture struct{ entries []*logrus.Entry }

// Levels reports the log levels the capture hook fires on.
func (h *logCapture) Levels() []logrus.Level { return logrus.AllLevels }

// Fire records the given log entry.
func (h *logCapture) Fire(e *logrus.Entry) error {
	h.entries = append(h.entries, e)
	return nil
}

func TestWebDAVWrite_MKCOL_Exists(t *testing.T) {
	conf := newWebDAVTestConfig(t)
	if err := conf.CreateDirectories(); err != nil {
		t.Fatalf("failed to create test directories: %v", err)
	}
	r := setupWebDAVRouter(conf)

	// First MKCOL creates the collection.
	w := httptest.NewRecorder()
	req := httptest.NewRequest(header.MethodMkcol, conf.BaseUri(WebDAVOriginals)+"/exists", nil)
	authBearer(req)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)

	// Capture the console-only system log while repeating the MKCOL against the
	// existing collection. Errors from x/net/webdav are routed through
	// event.SystemLog (never the browser log.* stream), so we hook it here.
	hook := &logCapture{}
	event.SystemLog.ReplaceHooks(logrus.LevelHooks{})
	event.SystemLog.AddHook(hook)
	defer event.SystemLog.ReplaceHooks(logrus.LevelHooks{})

	w = httptest.NewRecorder()
	req = httptest.NewRequest(header.MethodMkcol, conf.BaseUri(WebDAVOriginals)+"/exists", nil)
	authBearer(req)
	r.ServeHTTP(w, req)
	// Probing an existing collection is a client-side no-op; x/net/webdav returns 405.
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)

	// The benign probe must not be logged as an error, and must not leak the server path.
	var found bool
	for _, e := range hook.entries {
		if e.Level == logrus.DebugLevel && assert.Contains(t, e.Message, "already exists") {
			found = true
			assert.NotContains(t, e.Message, conf.OriginalsPath())
		}
		assert.NotEqual(t, logrus.ErrorLevel, e.Level, "MKCOL on existing collection must not log an error")
	}
	assert.True(t, found, "expected a debug entry for the existing collection")
}

// webDAVTestBody returns a body of the given size filled with a position-dependent pattern, so
// a copy that is short or padded differs in content as well as in length.
func webDAVTestBody(size int) []byte {
	body := make([]byte, size)

	for i := range body {
		body[i] = byte(i % 251)
	}

	return body
}

// webDAVPut uploads body to name and returns the recorder. Passing declareLength=false hands
// httptest a reader type it does not measure, leaving ContentLength at -1: the shape a chunked
// client sends, and the only one the declared-length precheck cannot see.
func webDAVPut(r *gin.Engine, conf *config.Config, name string, body []byte, declareLength bool) *httptest.ResponseRecorder {
	var req *http.Request

	target := conf.BaseUri(WebDAVOriginals) + "/" + name

	if declareLength {
		req = httptest.NewRequest(header.MethodPut, target, bytes.NewReader(body))
	} else {
		req = httptest.NewRequest(header.MethodPut, target, io.NopCloser(bytes.NewReader(body)))
	}

	w := httptest.NewRecorder()
	authBearer(req)
	r.ServeHTTP(w, req)

	return w
}

func TestWebDAVWrite_SeparatorRefused(t *testing.T) {
	conf := newWebDAVTestConfig(t)
	if err := conf.CreateDirectories(); err != nil {
		t.Fatalf("failed to create test directories: %v", err)
	}
	r := setupWebDAVRouter(conf)

	base := conf.BaseUri(WebDAVOriginals)
	stored := func(elem ...string) string {
		return filepath.Join(append([]string{conf.OriginalsPath()}, elem...)...)
	}
	send := func(method, target string, body []byte, destination string) *httptest.ResponseRecorder {
		var req *http.Request
		if body == nil {
			req = httptest.NewRequest(method, target, nil)
		} else {
			req = httptest.NewRequest(method, target, bytes.NewReader(body))
		}
		if destination != "" {
			req.Header.Set("Destination", destination)
		}
		w := httptest.NewRecorder()
		authBearer(req)
		r.ServeHTTP(w, req)
		return w
	}

	// Give every resolution of "dir\name" a writable target: the literal name resolves beside
	// the mount root, the normalized one inside an existing directory holding a file of known
	// bytes. Without this the normalized assertions cannot fail, because a sanitizing write
	// would stop at the missing parent rather than at the refusal.
	if err := os.MkdirAll(stored("dir"), fs.ModeDir); err != nil {
		t.Fatal(err)
	}
	occupied := webDAVTestBody(24)
	if err := os.WriteFile(stored("dir", "file.bin"), occupied, fs.ModeFile); err != nil {
		t.Fatal(err)
	}

	t.Run("PutRefused", func(t *testing.T) {
		hook := &logCapture{}
		event.SystemLog.ReplaceHooks(logrus.LevelHooks{})
		event.SystemLog.AddHook(hook)
		defer event.SystemLog.ReplaceHooks(logrus.LevelHooks{})

		assert.Equal(t, http.StatusBadRequest, send(header.MethodPut, base+"/dir%5Cfile.bin", webDAVTestBody(8), "").Code)
		assert.NoFileExists(t, stored(`dir\file.bin`))
		// #nosec G304 -- test reads file created under controlled temp directory.
		b, err := os.ReadFile(stored("dir", "file.bin"))
		assert.NoError(t, err)
		assert.Equal(t, occupied, b, "the normalized resolution must be untouched")

		var reported bool
		for _, e := range hook.entries {
			if e.Level == logrus.WarnLevel && strings.Contains(e.Message, "contains a path separator") {
				reported = true
				assert.NotContains(t, e.Message, conf.OriginalsPath())
			}
		}
		assert.True(t, reported, "expected a warning for the refused name")
	})
	t.Run("MkcolRefused", func(t *testing.T) {
		// A collection is a create too, and one that could never receive an upload.
		assert.Equal(t, http.StatusBadRequest, send(header.MethodMkcol, base+"/made%5Cup", nil, "").Code)
		assert.NoDirExists(t, stored(`made\up`))
		assert.NoDirExists(t, stored("made", "up"))
	})
	t.Run("LockCreatesNothing", func(t *testing.T) {
		// A LOCK carrying lockinfo asks for a placeholder when the name does not exist, which
		// is the one create that happens past the method gate.
		w := send(header.MethodLock, base+"/locked%5Cname.bin", []byte(lockInfoBody(4)), "")
		assert.NotEqual(t, http.StatusCreated, w.Code)
		assert.NoFileExists(t, stored(`locked\name.bin`))
		assert.NoFileExists(t, stored("locked", "name.bin"))
	})
	t.Run("MoveDestinationRefused", func(t *testing.T) {
		body := webDAVTestBody(8)
		assert.Equal(t, http.StatusCreated, send(header.MethodPut, base+"/movable.bin", body, "").Code)

		assert.Equal(t, http.StatusBadRequest, send(header.MethodMove, base+"/movable.bin", nil, base+"/dir%5Cfile.bin").Code)
		assert.FileExists(t, stored("movable.bin"), "the source must survive a refused move")
		assert.NoFileExists(t, stored(`dir\file.bin`))
		// #nosec G304 -- test reads file created under controlled temp directory.
		b, err := os.ReadFile(stored("dir", "file.bin"))
		assert.NoError(t, err)
		assert.Equal(t, occupied, b)
	})
	t.Run("CopyDestinationRefused", func(t *testing.T) {
		body := webDAVTestBody(8)
		assert.Equal(t, http.StatusCreated, send(header.MethodPut, base+"/copyable.bin", body, "").Code)

		assert.Equal(t, http.StatusBadRequest, send(header.MethodCopy, base+"/copyable.bin", nil, base+"/dir%5Cfile.bin").Code)
		assert.FileExists(t, stored("copyable.bin"))
		assert.NoFileExists(t, stored(`dir\file.bin`))
	})
	t.Run("RawSeparatorInDestinationRefused", func(t *testing.T) {
		// A client may send the character unencoded; it survives url.Parse either way.
		body := webDAVTestBody(8)
		assert.Equal(t, http.StatusCreated, send(header.MethodPut, base+"/rawmove.bin", body, "").Code)

		assert.Equal(t, http.StatusBadRequest, send(header.MethodMove, base+"/rawmove.bin", nil, base+`/dir\raw.bin`).Code)
		assert.FileExists(t, stored("rawmove.bin"))
		assert.NoFileExists(t, stored(`dir\raw.bin`))
		assert.NoFileExists(t, stored("dir", "raw.bin"))
	})
	t.Run("OrdinaryNamesAccepted", func(t *testing.T) {
		// Positive control, and the record of what stays accepted on purpose: characters that
		// are legal on POSIX and that both resolvers agree on. "%255C" decodes to the literal
		// text "%5C", which holds no separator, so widening the refusal to it would be wrong.
		for _, name := range []string{"a file.bin", "Fotos Größe.bin", "naïve+name.bin", "%5C literal.bin"} {
			body := webDAVTestBody(16)
			w := send(header.MethodPut, base+"/"+url.PathEscape(name), body, "")
			assert.Equal(t, http.StatusCreated, w.Code, name)
			// #nosec G304 -- test reads file created under controlled temp directory.
			b, err := os.ReadFile(stored(name))
			assert.NoError(t, err, name)
			assert.Equal(t, body, b, name)
		}
	})
}

func TestWebDAVSeparatorInName(t *testing.T) {
	t.Run("PutWithSeparator", func(t *testing.T) {
		assert.True(t, WebDAVSeparatorInName(httptest.NewRequest(header.MethodPut, "/originals/a%5Cb.bin", nil)))
	})
	t.Run("PutWithoutSeparator", func(t *testing.T) {
		assert.False(t, WebDAVSeparatorInName(httptest.NewRequest(header.MethodPut, "/originals/a/b.bin", nil)))
	})
	t.Run("MoveDestinationWithSeparator", func(t *testing.T) {
		req := httptest.NewRequest(header.MethodMove, "/originals/a.bin", nil)
		req.Header.Set("Destination", "/originals/c%5Cd.bin")
		assert.True(t, WebDAVSeparatorInName(req))
	})
	t.Run("CopyDestinationWithoutSeparator", func(t *testing.T) {
		req := httptest.NewRequest(header.MethodCopy, "/originals/a.bin", nil)
		req.Header.Set("Destination", "/originals/c/d.bin")
		assert.False(t, WebDAVSeparatorInName(req))
	})
	t.Run("ReadMethodIgnored", func(t *testing.T) {
		// A name that already holds one stays readable and removable.
		assert.False(t, WebDAVSeparatorInName(httptest.NewRequest(header.MethodGet, "/originals/a%5Cb.bin", nil)))
		assert.False(t, WebDAVSeparatorInName(httptest.NewRequest(header.MethodDelete, "/originals/a%5Cb.bin", nil)))
	})
}

func TestWebDAVWrite_PUT_OriginalsLimit(t *testing.T) {
	conf := newWebDAVTestConfig(t)
	conf.Options().OriginalsLimit = 1 // cap uploaded files at 1 MB
	if err := conf.CreateDirectories(); err != nil {
		t.Fatalf("failed to create test directories: %v", err)
	}
	r := setupWebDAVRouter(conf)

	limit := int(conf.OriginalsLimitBytes())
	stored := func(name string) string { return filepath.Join(conf.OriginalsPath(), name) }

	t.Run("UnderLimitAccepted", func(t *testing.T) {
		body := webDAVTestBody(limit / 2)
		assert.Equal(t, http.StatusCreated, webDAVPut(r, conf, "small.bin", body, true).Code)
		// #nosec G304 -- test reads file created under controlled temp directory.
		b, err := os.ReadFile(stored("small.bin"))
		assert.NoError(t, err)
		assert.Equal(t, body, b)
	})
	t.Run("AtLimitAccepted", func(t *testing.T) {
		body := webDAVTestBody(limit)
		assert.Equal(t, http.StatusCreated, webDAVPut(r, conf, "exact.bin", body, true).Code)
		// #nosec G304 -- test reads file created under controlled temp directory.
		b, err := os.ReadFile(stored("exact.bin"))
		assert.NoError(t, err)
		assert.Equal(t, body, b)
	})
	t.Run("UnderLimitUnknownLengthAccepted", func(t *testing.T) {
		body := webDAVTestBody(limit / 2)
		assert.Equal(t, http.StatusCreated, webDAVPut(r, conf, "chunked.bin", body, false).Code)
		// #nosec G304 -- test reads file created under controlled temp directory.
		b, err := os.ReadFile(stored("chunked.bin"))
		assert.NoError(t, err)
		assert.Equal(t, body, b)
	})
	t.Run("OverLimitDeclaredLengthRefused", func(t *testing.T) {
		hook := &logCapture{}
		event.SystemLog.ReplaceHooks(logrus.LevelHooks{})
		event.SystemLog.AddHook(hook)
		defer event.SystemLog.ReplaceHooks(logrus.LevelHooks{})

		w := webDAVPut(r, conf, "big.bin", webDAVTestBody(limit+1), true)
		assert.Equal(t, http.StatusRequestEntityTooLarge, w.Code)
		_, err := os.Stat(stored("big.bin"))
		assert.True(t, os.IsNotExist(err), "the destination must not be created")

		// The refusal returns before the handler's own logger, so it reports itself, at the
		// level that logger gives a write method and without the server path.
		var reported bool
		for _, e := range hook.entries {
			if e.Level == logrus.WarnLevel && strings.Contains(e.Message, "exceeds the originals limit") {
				reported = true
				assert.NotContains(t, e.Message, conf.OriginalsPath())
			}
		}
		assert.True(t, reported, "expected a warning for the refused upload")
	})
	t.Run("UnrelatedFailureKeepsFile", func(t *testing.T) {
		// The cleanup is keyed on the size bound, so a PUT failing for any other reason leaves
		// the destination alone. A lock token matching nothing fails confirmLocks without
		// opening the file, and the stored body is the size the cleanup would otherwise accept.
		body := webDAVTestBody(limit)
		assert.Equal(t, http.StatusCreated, webDAVPut(r, conf, "locked.bin", body, true).Code)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(header.MethodPut, conf.BaseUri(WebDAVOriginals)+"/locked.bin", bytes.NewReader(webDAVTestBody(16)))
		req.Header.Set("If", "(<opaquelocktoken:nonexistent>)")
		authBearer(req)
		r.ServeHTTP(w, req)
		assert.NotEqual(t, http.StatusCreated, w.Code)

		// #nosec G304 -- test reads file created under controlled temp directory.
		b, err := os.ReadFile(stored("locked.bin"))
		assert.NoError(t, err)
		assert.Equal(t, body, b)
	})
	t.Run("OverLimitUnknownLengthLeavesNoFile", func(t *testing.T) {
		// The handler answers with its own status, so the refusal shows as a non-2xx, not 413.
		// The two assertions carry the contract together: the log line shows the bound error
		// reached the matching branch, the absent destination shows it acted. Keep both.
		hook := &logCapture{}
		event.SystemLog.ReplaceHooks(logrus.LevelHooks{})
		event.SystemLog.AddHook(hook)
		defer event.SystemLog.ReplaceHooks(logrus.LevelHooks{})

		w := webDAVPut(r, conf, "bigchunked.bin", webDAVTestBody(limit+1), false)
		assert.NotEqual(t, http.StatusCreated, w.Code)
		_, err := os.Stat(stored("bigchunked.bin"))
		assert.True(t, os.IsNotExist(err), "the partial upload must be removed")

		// The logged error is the one the cleanup matched, so it names the exceeded bound.
		var reported bool
		for _, e := range hook.entries {
			if strings.Contains(e.Message, "request body too large") {
				reported = true
			}
		}
		assert.True(t, reported, "expected the size bound to be reported to the operator")
	})
	t.Run("OverLimitDeclaredLengthKeepsExistingFile", func(t *testing.T) {
		// The declared-length refusal returns before the handler opens the destination, which
		// is what leaves an existing file untouched.
		body := webDAVTestBody(limit / 2)
		assert.Equal(t, http.StatusCreated, webDAVPut(r, conf, "keep.bin", body, true).Code)

		assert.Equal(t, http.StatusRequestEntityTooLarge, webDAVPut(r, conf, "keep.bin", webDAVTestBody(limit+1), true).Code)
		// #nosec G304 -- test reads file created under controlled temp directory.
		b, err := os.ReadFile(stored("keep.bin"))
		assert.NoError(t, err)
		assert.Equal(t, body, b)
	})
	t.Run("OverLimitUnknownLengthLeavesNoFileAtUsedName", func(t *testing.T) {
		// The guarantee holds for a name already in use: an upload over the bound leaves no
		// file at the destination, whether or not the name was free when it started.
		assert.Equal(t, http.StatusCreated, webDAVPut(r, conf, "reused.bin", webDAVTestBody(limit/2), true).Code)

		assert.NotEqual(t, http.StatusCreated, webDAVPut(r, conf, "reused.bin", webDAVTestBody(limit+1), false).Code)
		_, err := os.Stat(stored("reused.bin"))
		assert.True(t, os.IsNotExist(err), "the destination must be left empty")
	})
}

func TestWebDAVWrite_PUT_OriginalsLimitDisabled(t *testing.T) {
	conf := newWebDAVTestConfig(t)
	conf.Options().OriginalsLimit = -1
	if err := conf.CreateDirectories(); err != nil {
		t.Fatalf("failed to create test directories: %v", err)
	}
	r := setupWebDAVRouter(conf)

	body := webDAVTestBody(2 * 1024 * 1024)
	assert.Equal(t, http.StatusCreated, webDAVPut(r, conf, "unbounded.bin", body, true).Code)
	// #nosec G304 -- test reads file created under controlled temp directory.
	b, err := os.ReadFile(filepath.Join(conf.OriginalsPath(), "unbounded.bin"))
	assert.NoError(t, err)
	assert.Equal(t, body, b)
}

func TestWebDAVRemovePartialUpload(t *testing.T) {
	// write creates a file of the given size and returns its name.
	write := func(t *testing.T, dir, name string, size int) string {
		t.Helper()
		fileName := filepath.Join(dir, name)
		if err := os.WriteFile(fileName, webDAVTestBody(size), fs.ModeFile); err != nil {
			t.Fatal(err)
		}
		return fileName
	}

	t.Run("Removed", func(t *testing.T) {
		fileName := write(t, t.TempDir(), "partial.bin", 64)
		WebDAVRemovePartialUpload(fileName, 64)
		_, err := os.Stat(fileName)
		assert.True(t, os.IsNotExist(err))
	})
	t.Run("OtherSizeKept", func(t *testing.T) {
		// A name that no longer holds a file of exactly the bound is a different file.
		fileName := write(t, t.TempDir(), "complete.bin", 32)
		WebDAVRemovePartialUpload(fileName, 64)
		_, err := os.Stat(fileName)
		assert.NoError(t, err)
	})
	t.Run("SymlinkKept", func(t *testing.T) {
		dir := t.TempDir()
		target := write(t, dir, "target.bin", 64)
		linkName := filepath.Join(dir, "link.bin")
		if err := os.Symlink(target, linkName); err != nil {
			t.Skipf("symlinks unsupported: %v", err)
		}
		// Lstat reports a link's size as the length of its target path, so passing that length
		// is what makes the regular-file check the only thing declining here.
		WebDAVRemovePartialUpload(linkName, int64(len(target)))
		_, err := os.Lstat(linkName)
		assert.NoError(t, err, "a link must keep its own name")
		_, err = os.Stat(target)
		assert.NoError(t, err, "a link target must not be removed")
	})
	t.Run("Missing", func(t *testing.T) {
		hook := &logCapture{}
		event.SystemLog.ReplaceHooks(logrus.LevelHooks{})
		event.SystemLog.AddHook(hook)
		defer event.SystemLog.ReplaceHooks(logrus.LevelHooks{})

		WebDAVRemovePartialUpload(filepath.Join(t.TempDir(), "absent.bin"), 64)
		// An absent destination declines at the Lstat and is never an operator-visible failure.
		for _, e := range hook.entries {
			assert.NotEqual(t, logrus.ErrorLevel, e.Level)
		}
	})
	t.Run("Directory", func(t *testing.T) {
		dirName := filepath.Join(t.TempDir(), "dir")
		if err := os.Mkdir(dirName, fs.ModeDir); err != nil {
			t.Fatal(err)
		}
		write(t, dirName, "child.bin", 64)
		WebDAVRemovePartialUpload(dirName, 64)
		_, err := os.Stat(dirName)
		assert.NoError(t, err)
	})
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

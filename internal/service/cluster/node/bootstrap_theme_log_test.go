package node

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/sirupsen/logrus"
	logtest "github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/service/cluster"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// captureSystemHook records the system log entries, including those of this package, for the duration of a test.
func captureSystemHook(t *testing.T) *logtest.Hook {
	t.Helper()

	origSystem, origLog := event.SystemLog, log
	logger, hook := logtest.NewNullLogger()
	logger.SetLevel(logrus.TraceLevel)
	event.SystemLog = logger
	log = logger

	t.Cleanup(func() {
		event.SystemLog = origSystem
		log = origLog
	})

	return hook
}

// logMessages returns the messages of the captured entries with the given level.
func logMessages(hook *logtest.Hook, level logrus.Level) (messages []string) {
	for _, entry := range hook.AllEntries() {
		if entry.Level == level {
			messages = append(messages, entry.Message)
		}
	}

	return messages
}

// newThemeArchive returns a theme archive with app.js and version.txt, plus the entries added by extra.
func newThemeArchive(t *testing.T, version string, extra func(zw *zip.Writer)) []byte {
	t.Helper()

	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	appJS, err := zw.Create(fs.AppJsFile)
	require.NoError(t, err)
	_, _ = appJS.Write([]byte("console.log('theme');\n"))
	versionTxt, err := zw.Create(fs.VersionTxtFile)
	require.NoError(t, err)
	_, _ = versionTxt.Write([]byte(version))

	if extra != nil {
		extra(zw)
	}

	require.NoError(t, zw.Close())

	return buf.Bytes()
}

// themePortal is a fake Portal that advertises a theme version and counts theme downloads.
type themePortal struct {
	*httptest.Server
	hits atomic.Int32
}

// newThemePortal returns a Portal that advertises the theme version and answers theme downloads with status and archive.
func newThemePortal(t *testing.T, version string, status int, archive []byte) *themePortal {
	t.Helper()

	return newThemePortalWith(t, version, http.StatusOK, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/zip")
		w.WriteHeader(status)
		_, _ = w.Write(archive)
	})
}

// newThemePortalWith returns a Portal that advertises the theme version, answers token requests with tokenStatus,
// and theme downloads with theme.
func newThemePortalWith(t *testing.T, version string, tokenStatus int, theme http.HandlerFunc) *themePortal {
	t.Helper()

	p := &themePortal{}
	p.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/cluster/nodes/register":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(cluster.RegisterResponse{
				UUID:    rnd.UUID(),
				Node:    cluster.Node{ClientID: "cs5gfen1bgxz7s9i", Name: "pp-node-01", Theme: version},
				Secrets: &cluster.RegisterSecrets{ClientSecret: cluster.ExampleClientSecret},
				Theme:   version,
			})
		case "/api/v1/oauth/token":
			if tokenStatus != http.StatusOK {
				w.WriteHeader(tokenStatus)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]string{"access_token": "tok", "token_type": "Bearer"})
		case "/api/v1/cluster/theme":
			p.hits.Add(1)
			theme(w, r)
		default:
			http.NotFound(w, r)
		}
	}))

	t.Cleanup(p.Close)

	return p
}

// logContains reports whether a captured entry of any level contains text.
func logContains(hook *logtest.Hook, text string) bool {
	for _, entry := range hook.AllEntries() {
		if strings.Contains(entry.Message, text) {
			return true
		}
	}

	return false
}

func TestThemeInstall_Warning(t *testing.T) {
	overLimit := newThemeArchive(t, "theme-v1", func(zw *zip.Writer) {
		// The declared size exceeds the entry limit, so the entry is refused before its data is read.
		w, err := zw.CreateRaw(&zip.FileHeader{Name: "large.css", Method: zip.Store, CompressedSize64: 4, UncompressedSize64: 33 * fs.MB})
		require.NoError(t, err)
		_, _ = w.Write([]byte("body"))
	})
	parentDir := newThemeArchive(t, "theme-v1", func(zw *zip.Writer) {
		w, err := zw.Create("css/../style.css")
		require.NoError(t, err)
		_, _ = w.Write([]byte("body{}\n"))
	})

	cases := []struct {
		name    string
		db      string
		status  int
		archive []byte
		cause   string
	}{
		{name: "OverLimitEntry", db: "bootstrap-theme-warn-limit", status: http.StatusOK, archive: overLimit, cause: "large.css cannot be extracted"},
		{name: "ParentDirEntry", db: "bootstrap-theme-warn-parent", status: http.StatusOK, archive: parentDir, cause: "entry css/../style.css cannot be extracted"},
		{name: "Forbidden", db: "bootstrap-theme-warn-forbidden", status: http.StatusForbidden, archive: nil, cause: "403 Forbidden"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			srv := newThemePortal(t, "theme-v1", tc.status, tc.archive)
			c := newBootstrapTestConfig(t, tc.db)
			c.Options().PortalUrl = srv.URL
			c.Options().JoinToken = cluster.ExampleJoinToken
			nodeThemeDir := c.NodeThemePath()
			require.NoError(t, os.RemoveAll(nodeThemeDir))
			hook := captureSystemHook(t)

			require.NoError(t, InitConfig(c))

			warnings := logMessages(hook, logrus.WarnLevel)
			require.Len(t, warnings, 1, strings.Join(warnings, "\n"))
			assert.True(t, strings.HasPrefix(warnings[0], "cluster: theme theme-v1 install failed › "), warnings[0])
			assert.Contains(t, warnings[0], tc.cause)
			assert.False(t, logContains(hook, "theme download skipped"))
			assert.Equal(t, int32(1), srv.hits.Load())
			assert.NoFileExists(t, filepath.Join(nodeThemeDir, fs.AppJsFile))
			assert.Empty(t, c.NodeThemeVersion())
			assert.NotEqual(t, nodeThemeDir, c.ThemePath())
		})
	}
}

func TestThemeInstall_UpdateFailed(t *testing.T) {
	// installThemeV1 writes an installed theme-v1 into the node theme directory.
	installThemeV1 := func(t *testing.T, dir string) {
		require.NoError(t, os.MkdirAll(dir, fs.ModeDir))
		require.NoError(t, os.WriteFile(filepath.Join(dir, fs.AppJsFile), []byte("console.log('theme-v1');\n"), fs.ModeFile))
		require.NoError(t, os.WriteFile(filepath.Join(dir, fs.VersionTxtFile), []byte("theme-v1"), fs.ModeFile))
		require.NoError(t, os.WriteFile(filepath.Join(dir, "v1-only.css"), []byte("body{}\n"), fs.ModeFile))
	}

	t.Run("RefusedKeepsPreviousTheme", func(t *testing.T) {
		srv := newThemePortal(t, "theme-v2", http.StatusForbidden, nil)
		c := newBootstrapTestConfig(t, "bootstrap-theme-update-refused")
		c.Options().PortalUrl = srv.URL
		c.Options().JoinToken = cluster.ExampleJoinToken
		nodeThemeDir := c.NodeThemePath()
		installThemeV1(t, nodeThemeDir)
		hook := captureSystemHook(t)

		require.NoError(t, InitConfig(c))

		require.Len(t, logMessages(hook, logrus.WarnLevel), 1)
		assert.FileExists(t, filepath.Join(nodeThemeDir, "v1-only.css"))
		assert.Equal(t, "theme-v1", c.NodeThemeVersion())
		assert.Equal(t, nodeThemeDir, c.ThemePath())
	})
	t.Run("InvalidArchiveUsesDefaultTheme", func(t *testing.T) {
		archive := newThemeArchive(t, "theme-v2", func(zw *zip.Writer) {
			w, err := zw.Create("css/../style.css")
			require.NoError(t, err)
			_, _ = w.Write([]byte("body{}\n"))
		})
		srv := newThemePortal(t, "theme-v2", http.StatusOK, archive)
		c := newBootstrapTestConfig(t, "bootstrap-theme-update-invalid")
		c.Options().PortalUrl = srv.URL
		c.Options().JoinToken = cluster.ExampleJoinToken
		nodeThemeDir := c.NodeThemePath()
		installThemeV1(t, nodeThemeDir)
		hook := captureSystemHook(t)

		require.NoError(t, InitConfig(c))

		require.Len(t, logMessages(hook, logrus.WarnLevel), 1)
		assert.NoFileExists(t, filepath.Join(nodeThemeDir, fs.AppJsFile))
		assert.NoFileExists(t, filepath.Join(nodeThemeDir, "v1-only.css"))
		assert.Empty(t, c.NodeThemeVersion())
		assert.NotEqual(t, nodeThemeDir, c.ThemePath())
	})
}

func TestThemeInstall_NoWarning(t *testing.T) {
	t.Run("NoTheme", func(t *testing.T) {
		srv := newThemePortal(t, "", http.StatusOK, nil)
		c := newBootstrapTestConfig(t, "bootstrap-theme-none")
		c.Options().PortalUrl = srv.URL
		c.Options().JoinToken = cluster.ExampleJoinToken
		hook := captureSystemHook(t)

		require.NoError(t, InitConfig(c))

		assert.Empty(t, logMessages(hook, logrus.WarnLevel))
		assert.Empty(t, c.NodeThemeVersion())
	})
	t.Run("AlreadyCurrent", func(t *testing.T) {
		srv := newThemePortal(t, "theme-v1", http.StatusForbidden, nil)
		c := newBootstrapTestConfig(t, "bootstrap-theme-current")
		c.Options().PortalUrl = srv.URL
		c.Options().JoinToken = cluster.ExampleJoinToken
		nodeThemeDir := c.NodeThemePath()
		require.NoError(t, os.MkdirAll(nodeThemeDir, fs.ModeDir))
		require.NoError(t, os.WriteFile(filepath.Join(nodeThemeDir, fs.AppJsFile), []byte("console.log('theme');\n"), fs.ModeFile))
		require.NoError(t, os.WriteFile(filepath.Join(nodeThemeDir, fs.VersionTxtFile), []byte("theme-v1"), fs.ModeFile))
		hook := captureSystemHook(t)

		require.NoError(t, InitConfig(c))

		assert.Empty(t, logMessages(hook, logrus.WarnLevel))
		assert.Equal(t, "theme-v1", c.NodeThemeVersion())
		assert.Equal(t, nodeThemeDir, c.ThemePath())
	})
	t.Run("NotFound", func(t *testing.T) {
		srv := newThemePortal(t, "theme-v1", http.StatusNotFound, nil)
		c := newBootstrapTestConfig(t, "bootstrap-theme-notfound")
		c.Options().PortalUrl = srv.URL
		c.Options().JoinToken = cluster.ExampleJoinToken
		hook := captureSystemHook(t)

		require.NoError(t, InitConfig(c))

		assert.Empty(t, logMessages(hook, logrus.WarnLevel))
		assert.Empty(t, c.NodeThemeVersion())
		assert.Equal(t, int32(1), srv.hits.Load())
	})
	t.Run("TokenRefused", func(t *testing.T) {
		srv := newThemePortalWith(t, "theme-v1", http.StatusUnauthorized, func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusForbidden)
		})
		c := newBootstrapTestConfig(t, "bootstrap-theme-token")
		c.Options().PortalUrl = srv.URL
		c.Options().JoinToken = cluster.ExampleJoinToken
		hook := captureSystemHook(t)

		require.NoError(t, InitConfig(c))

		assert.Empty(t, logMessages(hook, logrus.WarnLevel))
		assert.True(t, logContains(hook, "portal access token request failed"))
		assert.Equal(t, int32(0), srv.hits.Load())
		assert.Empty(t, c.NodeThemeVersion())
	})
	t.Run("ConnectionClosed", func(t *testing.T) {
		srv := newThemePortalWith(t, "theme-v1", http.StatusOK, func(w http.ResponseWriter, _ *http.Request) {
			conn, _, err := w.(http.Hijacker).Hijack()

			if !assert.NoError(t, err) {
				return
			}

			_ = conn.Close()
		})
		c := newBootstrapTestConfig(t, "bootstrap-theme-closed")
		c.Options().PortalUrl = srv.URL
		c.Options().JoinToken = cluster.ExampleJoinToken
		hook := captureSystemHook(t)

		require.NoError(t, InitConfig(c))

		assert.Empty(t, logMessages(hook, logrus.WarnLevel))
		assert.True(t, logContains(hook, "cluster: theme download skipped ("))
		// The HTTP client retries an idempotent request once after the connection was closed.
		assert.GreaterOrEqual(t, srv.hits.Load(), int32(1))
		assert.Empty(t, c.NodeThemeVersion())
	})
}

func TestLogThemeSyncError(t *testing.T) {
	t.Run("Nil", func(t *testing.T) {
		hook := captureSystemHook(t)
		logThemeSyncError(nil)
		assert.Empty(t, hook.AllEntries())
	})
	t.Run("Unreachable", func(t *testing.T) {
		hook := captureSystemHook(t)
		logThemeSyncError(errors.New("dial tcp 192.0.2.1:443: connect: connection refused"))
		assert.Empty(t, logMessages(hook, logrus.WarnLevel))
		debug := logMessages(hook, logrus.DebugLevel)
		require.Len(t, debug, 1)
		assert.Equal(t, "cluster: theme download skipped (dial tcp 192.0.2.1:443: connect: connection refused)", debug[0])
	})
	t.Run("InstallFailed", func(t *testing.T) {
		hook := captureSystemHook(t)
		logThemeSyncError(&themeInstallError{Version: "theme-v2", Err: errors.New("zip entry exceeds limit")})
		warnings := logMessages(hook, logrus.WarnLevel)
		require.Len(t, warnings, 1)
		assert.Equal(t, "cluster: theme theme-v2 install failed › zip entry exceeds limit", warnings[0])
	})
	t.Run("InstallFailedUnknownVersion", func(t *testing.T) {
		hook := captureSystemHook(t)
		logThemeSyncError(&themeInstallError{Err: errors.New("500 Internal Server Error")})
		warnings := logMessages(hook, logrus.WarnLevel)
		require.Len(t, warnings, 1)
		assert.Equal(t, "cluster: theme install failed › 500 Internal Server Error", warnings[0])
	})
	t.Run("Sanitized", func(t *testing.T) {
		hook := captureSystemHook(t)
		logThemeSyncError(&themeInstallError{Version: "v1\nsecond", Err: errors.New("bad\nline")})
		logThemeSyncError(&themeInstallError{Err: errors.New("bad\nline")})
		warnings := logMessages(hook, logrus.WarnLevel)
		require.Len(t, warnings, 2)
		assert.NotContains(t, warnings[0], "\n")
		assert.Equal(t, "cluster: theme install failed › bad line", warnings[1])
	})
}

func TestThemeInstallError(t *testing.T) {
	t.Run("Error", func(t *testing.T) {
		cause := errors.New("403 Forbidden")
		err := &themeInstallError{Version: "theme-v1", Err: cause}
		assert.Equal(t, "403 Forbidden", err.Error())
		assert.ErrorIs(t, err, cause)
	})
}

func TestInstallNodeTheme(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		c := newBootstrapTestConfig(t, "bootstrap-theme-install")
		dir := filepath.Join(t.TempDir(), "theme")

		require.NoError(t, installNodeTheme(c, bytes.NewReader(newThemeArchive(t, "theme-v1", nil)), dir, false))

		assert.FileExists(t, filepath.Join(dir, fs.AppJsFile))
	})
	t.Run("InvalidArchive", func(t *testing.T) {
		c := newBootstrapTestConfig(t, "bootstrap-theme-install-invalid")
		dir := filepath.Join(t.TempDir(), "theme")

		assert.Error(t, installNodeTheme(c, strings.NewReader("not a zip"), dir, false))
		assert.NoFileExists(t, filepath.Join(dir, fs.AppJsFile))
	})
}

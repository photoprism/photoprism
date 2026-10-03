package webdav

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"testing"

	dav "github.com/emersion/go-webdav"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
)

// downloadTestClient returns a client for a server that serves body on GET, and calls onGet before
// answering so a test can change the destination while the request is in flight.
func downloadTestClient(t *testing.T, body string, onGet func()) *Client {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		if onGet != nil {
			onGet()
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(body))
	}))

	t.Cleanup(server.Close)

	c, err := NewClient(server.URL+"/", "", "", TimeoutLow, "")
	require.NoError(t, err)

	return c
}

func TestClient_DownloadCreatesExclusively(t *testing.T) {
	t.Run("DestinationAppearingDuringTheRequestIsKept", func(t *testing.T) {
		dest := filepath.Join(t.TempDir(), "local.bin")

		// A file created while the request is in flight must be kept: the destination is claimed
		// only at the end, and only when the name is still free.
		client := downloadTestClient(t, "remote", func() {
			require.NoError(t, os.WriteFile(dest, []byte("local"), fs.ModeFile))
		})

		err := client.Download("/local.bin", dest, false)
		// The sync worker reads the collision from the error type, and counts it against neither its
		// retry budget nor its error log.
		assert.ErrorIs(t, err, os.ErrExist)

		got, readErr := os.ReadFile(dest) //nolint:gosec // test fixture reads a test-owned temporary path
		require.NoError(t, readErr)
		assert.Equal(t, "local", string(got), "the local file must not be replaced")
	})
	t.Run("ExistingDestinationIsRefused", func(t *testing.T) {
		// Guards the early existence check rather than the publish; the subtest above is what
		// discriminates it.
		dest := filepath.Join(t.TempDir(), "local.bin")
		require.NoError(t, os.WriteFile(dest, []byte("local"), fs.ModeFile))

		client := downloadTestClient(t, "remote", nil)
		assert.ErrorIs(t, client.Download("/local.bin", dest, false), os.ErrExist)

		got, readErr := os.ReadFile(dest) //nolint:gosec // test fixture reads a test-owned temporary path
		require.NoError(t, readErr)
		assert.Equal(t, "local", string(got))
	})
	t.Run("NewDestinationIsWritten", func(t *testing.T) {
		dest := filepath.Join(t.TempDir(), "new.bin")
		client := downloadTestClient(t, "remote", nil)
		require.NoError(t, client.Download("/new.bin", dest, false))

		got, readErr := os.ReadFile(dest) //nolint:gosec // test fixture reads a test-owned temporary path
		require.NoError(t, readErr)
		assert.Equal(t, "remote", string(got))
	})
	t.Run("ForcedReplacementSucceeds", func(t *testing.T) {
		dir := t.TempDir()
		dest := filepath.Join(dir, "local.bin")
		require.NoError(t, os.WriteFile(dest, []byte("local"), fs.ModeFile))

		client := downloadTestClient(t, "remote", nil)
		require.NoError(t, client.Download("/local.bin", dest, true))

		got, readErr := os.ReadFile(dest) //nolint:gosec // test fixture reads a test-owned temporary path
		require.NoError(t, readErr)
		assert.Equal(t, "remote", string(got))
		assert.Equal(t, []string{"local.bin"}, dirEntries(t, dir), "no temporary file may be left behind")
	})
}

func TestClient_DownloadLeavesNothingBehindOnFailure(t *testing.T) {
	// A response that promises more than it delivers, so the copy fails part way through.
	truncating := func(t *testing.T) *Client {
		t.Helper()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}

			w.Header().Set("Content-Length", "64")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("short"))

			if hijacker, ok := w.(http.Hijacker); ok {
				conn, _, hijackErr := hijacker.Hijack()

				if hijackErr == nil {
					_ = conn.Close()
				}
			}
		}))

		t.Cleanup(server.Close)

		c, err := NewClient(server.URL+"/", "", "", TimeoutLow, "")
		require.NoError(t, err)

		return c
	}

	t.Run("ForcedReplacementKeepsTheLocalFile", func(t *testing.T) {
		dir := t.TempDir()
		dest := filepath.Join(dir, "local.bin")
		require.NoError(t, os.WriteFile(dest, []byte("local"), fs.ModeFile))

		require.Error(t, truncating(t).Download("/local.bin", dest, true))

		got, readErr := os.ReadFile(dest) //nolint:gosec // test fixture reads a test-owned temporary path
		require.NoError(t, readErr)
		assert.Equal(t, "local", string(got), "a failed replacement must leave the local file intact")
		assert.Equal(t, []string{"local.bin"}, dirEntries(t, dir))
	})
	t.Run("NewDestinationIsRemoved", func(t *testing.T) {
		dir := t.TempDir()
		dest := filepath.Join(dir, "new.bin")

		err := truncating(t).Download("/new.bin", dest, false)
		// Anchored on the copy stage, so a failure before the file is created cannot satisfy it.
		require.ErrorContains(t, err, "failed writing")
		assert.Empty(t, dirEntries(t, dir), "a failed download must leave nothing behind")
	})
	t.Run("OversizeKeepsTheLocalFile", func(t *testing.T) {
		dir := t.TempDir()
		dest := filepath.Join(dir, "local.bin")
		require.NoError(t, os.WriteFile(dest, []byte("local"), fs.ModeFile))

		client := downloadTestClient(t, "a much longer remote body than the limit allows", nil)
		client.SetDownloadLimit(8)

		require.Error(t, client.Download("/local.bin", dest, true))

		got, readErr := os.ReadFile(dest) //nolint:gosec // test fixture reads a test-owned temporary path
		require.NoError(t, readErr)
		assert.Equal(t, "local", string(got))
		assert.Equal(t, []string{"local.bin"}, dirEntries(t, dir))
	})
}

// roundTripFunc adapts a function to http.RoundTripper.
type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

// failingBody fails on the first read, calling onRead before it does so a test can change the
// destination while the response is being consumed.
type failingBody struct {
	onRead func()
	read   bool
}

func (b *failingBody) Read([]byte) (int, error) {
	if !b.read {
		b.read = true
		b.onRead()
	}

	return 0, io.ErrUnexpectedEOF
}

func (b *failingBody) Close() error { return nil }

func TestClient_DownloadFailureKeepsAnotherWritersFile(t *testing.T) {
	dir := t.TempDir()
	dest := filepath.Join(dir, "local.bin")

	written := filepath.Join(dir, "another.bin")
	require.NoError(t, os.WriteFile(written, []byte("another writer"), fs.ModeFile))

	// A file that takes the destination while the response is being read belongs to whoever put it
	// there.
	body := &failingBody{onRead: func() {
		assert.NoFileExists(t, dest, "the destination is claimed only once the download is complete")
		require.NoError(t, os.Rename(written, dest))
	}}

	client, err := NewClient("http://127.0.0.1/", "", "", TimeoutLow, "")
	require.NoError(t, err)

	client.client, err = dav.NewClient(&http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: body, ContentLength: -1, Request: r}, nil
	})}, "http://127.0.0.1/")
	require.NoError(t, err)

	require.ErrorContains(t, client.Download("/local.bin", dest, false), "failed writing")

	got, readErr := os.ReadFile(dest) //nolint:gosec // test fixture reads a test-owned temporary path
	require.NoError(t, readErr, "failure cleanup must remove only the file this call created")
	assert.Equal(t, "another writer", string(got))
	assert.Equal(t, []string{"local.bin"}, dirEntries(t, dir), "no staging file may survive")
}

func TestPublishSink(t *testing.T) {
	stage := func(t *testing.T) (sink, dest string) {
		t.Helper()

		dir := t.TempDir()
		sink = filepath.Join(dir, "staged.tmp")
		require.NoError(t, os.WriteFile(sink, []byte("remote"), fs.ModeFile))

		return sink, filepath.Join(dir, "local.bin")
	}

	t.Run("Success", func(t *testing.T) {
		sink, dest := stage(t)
		require.NoError(t, publishSink(sink, dest, false))

		got, err := os.ReadFile(dest) //nolint:gosec // test fixture reads a test-owned temporary path
		require.NoError(t, err)
		assert.Equal(t, "remote", string(got))
		assert.Equal(t, []string{"local.bin"}, dirEntries(t, filepath.Dir(dest)), "the staging file is gone")
	})
	t.Run("TakenName", func(t *testing.T) {
		sink, dest := stage(t)
		require.NoError(t, os.WriteFile(dest, []byte("local"), fs.ModeFile))
		assert.ErrorIs(t, publishSink(sink, dest, false), os.ErrExist)

		got, err := os.ReadFile(dest) //nolint:gosec // test fixture reads a test-owned temporary path
		require.NoError(t, err)
		assert.Equal(t, "local", string(got), "a taken name keeps its file")
	})
	t.Run("TakenNameWithForce", func(t *testing.T) {
		sink, dest := stage(t)
		require.NoError(t, os.WriteFile(dest, []byte("local"), fs.ModeFile))
		require.NoError(t, publishSink(sink, dest, true))

		got, err := os.ReadFile(dest) //nolint:gosec // test fixture reads a test-owned temporary path
		require.NoError(t, err)
		assert.Equal(t, "remote", string(got))
	})
	t.Run("MissingSink", func(t *testing.T) {
		_, dest := stage(t)
		assert.Error(t, publishSink(filepath.Join(filepath.Dir(dest), "absent.tmp"), dest, false))
		assert.NoFileExists(t, dest)
	})
	t.Run("WithoutHardLinks", func(t *testing.T) {
		// The path a filesystem without hard links takes, with the same two outcomes.
		unsupported := func(string, string) error { return &os.LinkError{Op: "link", Err: syscall.EPERM} }
		orig := linkFile
		linkFile = unsupported
		t.Cleanup(func() { linkFile = orig })

		sink, dest := stage(t)
		require.NoError(t, publishSink(sink, dest, false))
		got, err := os.ReadFile(dest) //nolint:gosec // test fixture reads a test-owned temporary path
		require.NoError(t, err)
		assert.Equal(t, "remote", string(got), "the staged bytes are published, not an empty file")
		assert.Equal(t, []string{"local.bin"}, dirEntries(t, filepath.Dir(dest)))

		taken, occupied := stage(t)
		require.NoError(t, os.WriteFile(occupied, []byte("local"), fs.ModeFile))
		assert.ErrorIs(t, publishSink(taken, occupied, false), os.ErrExist)
		kept, err := os.ReadFile(occupied) //nolint:gosec // test fixture reads a test-owned temporary path
		require.NoError(t, err)
		assert.Equal(t, "local", string(kept), "a taken name keeps its file")
	})
}

// dirEntries returns the sorted names of the entries in dir.
func dirEntries(t *testing.T, dir string) []string {
	t.Helper()

	entries, err := os.ReadDir(dir)
	require.NoError(t, err)

	names := make([]string, 0, len(entries))

	for _, entry := range entries {
		names = append(names, entry.Name())
	}

	return names
}

func TestClient_DownloadDoesNotFollowSymlinks(t *testing.T) {
	t.Run("DanglingSymlinkIsRefused", func(t *testing.T) {
		dir := t.TempDir()
		target := filepath.Join(dir, "target.bin")
		dest := filepath.Join(dir, "link.bin")
		require.NoError(t, os.Symlink(target, dest))

		// A symlink is not a destination this call may write through, and the target must not be
		// created on its behalf.
		require.Error(t, downloadTestClient(t, "remote", nil).Download("/link.bin", dest, false))
		assert.NoFileExists(t, target)
	})
	t.Run("ForcedReplacementLeavesTheTargetAlone", func(t *testing.T) {
		dir := t.TempDir()
		target := filepath.Join(dir, "target.bin")
		dest := filepath.Join(dir, "link.bin")
		require.NoError(t, os.WriteFile(target, []byte("target"), fs.ModeFile))
		require.NoError(t, os.Symlink(target, dest))

		require.NoError(t, downloadTestClient(t, "remote", nil).Download("/link.bin", dest, true))

		got, readErr := os.ReadFile(target) //nolint:gosec // test fixture reads a test-owned temporary path
		require.NoError(t, readErr)
		assert.Equal(t, "target", string(got), "the link target keeps its contents")

		written, readErr := os.ReadFile(dest) //nolint:gosec // test fixture reads a test-owned temporary path
		require.NoError(t, readErr)
		assert.Equal(t, "remote", string(written))
	})
}

func TestClient_DownloadConcurrentDestination(t *testing.T) {
	t.Run("OnlyOneCreates", func(t *testing.T) {
		dir := t.TempDir()
		dest := filepath.Join(dir, "shared.bin")
		client := downloadTestClient(t, "remote", nil)

		var wg sync.WaitGroup
		results := make([]error, 8)

		for i := range results {
			wg.Add(1)

			go func() {
				defer wg.Done()
				results[i] = client.Download("/shared.bin", dest, false)
			}()
		}

		wg.Wait()

		created := 0

		for _, err := range results {
			if err == nil {
				created++
			}
		}

		assert.Equal(t, 1, created, "exactly one call may create the destination")
		assert.Equal(t, []string{"shared.bin"}, dirEntries(t, dir))
	})
	t.Run("ForcedReplacementsLeaveOneCompleteFile", func(t *testing.T) {
		dir := t.TempDir()
		dest := filepath.Join(dir, "shared.bin")
		client := downloadTestClient(t, "a complete remote body", nil)

		var wg sync.WaitGroup

		for range 8 {
			wg.Add(1)

			go func() {
				defer wg.Done()
				_ = client.Download("/shared.bin", dest, true)
			}()
		}

		wg.Wait()

		got, readErr := os.ReadFile(dest) //nolint:gosec // test fixture reads a test-owned temporary path
		require.NoError(t, readErr)
		assert.Equal(t, "a complete remote body", string(got), "the destination is never a partial write")
		assert.Equal(t, []string{"shared.bin"}, dirEntries(t, dir), "no staging file may survive")
	})
}

func TestClient_DownloadKeepsDestinationMode(t *testing.T) {
	dir := t.TempDir()
	dest := filepath.Join(dir, "restricted.bin")
	require.NoError(t, os.WriteFile(dest, []byte("local"), 0o600))
	require.NoError(t, os.Chmod(dest, 0o600))

	require.NoError(t, downloadTestClient(t, "remote", nil).Download("/restricted.bin", dest, true))

	info, err := os.Stat(dest)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o600), info.Mode().Perm(), "replacing a file must not widen its mode")
}

// captureLog redirects the package logger for the duration of the test and returns its entries.
func captureLog(t *testing.T) *test.Hook {
	t.Helper()

	orig := log
	logger, hook := test.NewNullLogger()
	logger.SetLevel(logrus.TraceLevel)
	log = logger

	t.Cleanup(func() { log = orig })

	return hook
}

// connectLog returns everything NewClient logged while connecting to the given endpoint.
func connectLog(t *testing.T, serverUrl, user, pass string) string {
	t.Helper()

	hook := captureLog(t)

	if _, err := NewClient(serverUrl, user, pass, TimeoutLow, ""); err != nil {
		t.Fatal(err)
	}

	var logged strings.Builder

	for _, entry := range hook.AllEntries() {
		logged.WriteString(entry.Message)
		logged.WriteString("\n")
	}

	return logged.String()
}

func TestNewClientLogsEndpointWithoutCredentials(t *testing.T) {
	// Deliberately made of characters the sanitizers pass through unchanged, so the absence of the
	// secrets below means they were redacted rather than merely escaped out of recognition.
	const password = "sup3rs3cr3t-w3bdav-p4ss"
	const token = "sup3rs3cr3t-w3bdav-t0k3n"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(server.Close)

	t.Run("Userinfo", func(t *testing.T) {
		logged := connectLog(t, server.URL+"/", "webdav-user", password)
		assert.NotContains(t, logged, password, "the configured password must not reach the log")
		assert.Contains(t, logged, clean.UriRedactedValue, "the credential is marked as removed")
		assert.Contains(t, logged, "webdav-user", "the account name identifies the connection")
	})
	t.Run("QueryCredential", func(t *testing.T) {
		logged := connectLog(t, server.URL+"/?token="+token, "webdav-user", password)
		assert.NotContains(t, logged, token, "a credential parameter must not reach the log")
		assert.NotContains(t, logged, password, "the configured password must not reach the log")
	})
	t.Run("MalformedQueryCredential", func(t *testing.T) {
		for _, query := range []string{"?token=" + token + ";tail", "?token=" + token + "%zz", "?other=ok&token=" + token + ";tail"} {
			logged := connectLog(t, server.URL+"/"+query, "webdav-user", password)
			assert.NotContainsf(t, logged, token, "%s must not reach the log", query)
			assert.NotContains(t, logged, password, "the configured password must not reach the log")
		}
	})
}

package webdav

import (
	"bytes"
	"context"
	"io"
	"math"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	logtest "github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// stallSettings sets the stall timeout, the data that restarts it, and the response rate until the test ends.
func stallSettings(t *testing.T, timeout time.Duration, minBytes, rate int64) {
	t.Helper()

	prevTimeout, prevBytes, prevRate := transferStallTimeout, transferStallBytes, transferResponseRate
	transferStallTimeout, transferStallBytes, transferResponseRate = timeout, minBytes, rate
	t.Cleanup(func() {
		transferStallTimeout, transferStallBytes, transferResponseRate = prevTimeout, prevBytes, prevRate
	})
}

// stallTestServer returns a client for a server that answers with handler, whose blocked requests are released
// when the test ends.
func stallTestServer(t *testing.T, handler func(w http.ResponseWriter, r *http.Request, release <-chan struct{})) *Client {
	t.Helper()

	release := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handler(w, r, release)
	}))
	t.Cleanup(server.Close)
	t.Cleanup(func() { close(release) })

	c, err := NewClient(server.URL+"/", "", "", TimeoutLow, "")
	require.NoError(t, err)

	return c
}

// within runs f and fails the test if it does not return within the specified time.
func within(t *testing.T, limit time.Duration, f func() error) error {
	t.Helper()

	done := make(chan error, 1)
	go func() { done <- f() }()

	select {
	case err := <-done:
		return err
	case <-time.After(limit):
		t.Fatalf("transfer did not end within %s", limit)
		return nil
	}
}

func TestStallWatch(t *testing.T) {
	stallSettings(t, 400*time.Millisecond, 4, 1<<20)

	t.Run("Stalls", func(t *testing.T) {
		w := newStallWatch(context.Background(), 20*time.Millisecond)
		defer w.Stop()
		<-w.Context().Done()
		assert.True(t, w.Stalled())
	})
	t.Run("MovedKeepsItOpen", func(t *testing.T) {
		w := newStallWatch(context.Background(), transferStallTimeout)
		defer w.Stop()
		for i := 0; i < 30; i++ {
			time.Sleep(20 * time.Millisecond)
			w.Moved(4)
		}
		assert.NoError(t, w.Context().Err())
		assert.False(t, w.Stalled())
	})
	t.Run("BelowFloor", func(t *testing.T) {
		// Less data than the floor within the timeout does not restart it.
		w := newStallWatch(context.Background(), 100*time.Millisecond)
		defer w.Stop()
		for i := 0; i < 5; i++ {
			time.Sleep(60 * time.Millisecond)
			w.Moved(1)
		}
		assert.True(t, w.Stalled())
	})
	t.Run("ReaderCounts", func(t *testing.T) {
		w := newStallWatch(context.Background(), transferStallTimeout)
		defer w.Stop()
		r := w.Reader(strings.NewReader(strings.Repeat("a", 120)), nil)
		buf := make([]byte, 4)
		for i := 0; i < 30; i++ {
			time.Sleep(20 * time.Millisecond)
			_, err := r.Read(buf)
			require.NoError(t, err)
		}
		assert.False(t, w.Stalled())
	})
	t.Run("AwaitAfterEOF", func(t *testing.T) {
		// Once the body is read, the wait for the response replaces the timeout.
		w := newStallWatch(context.Background(), 50*time.Millisecond)
		defer w.Stop()
		r := w.Reader(strings.NewReader("body"), func() time.Duration { return 600 * time.Millisecond })
		_, err := io.ReadAll(r)
		require.NoError(t, err)
		time.Sleep(200 * time.Millisecond)
		w.Moved(4)
		time.Sleep(100 * time.Millisecond)
		assert.False(t, w.Stalled())
		<-w.Context().Done()
		assert.True(t, w.Stalled())
	})
	t.Run("Stop", func(t *testing.T) {
		w := newStallWatch(context.Background(), 20*time.Millisecond)
		w.Stop()
		w.Moved(4)
		time.Sleep(60 * time.Millisecond)
		assert.False(t, w.Stalled())
		assert.Error(t, w.Context().Err())
	})
	t.Run("Disabled", func(t *testing.T) {
		w := newStallWatch(context.Background(), 0)
		defer w.Stop()
		time.Sleep(20 * time.Millisecond)
		w.Moved(4)
		assert.False(t, w.Stalled())
		assert.NoError(t, w.Context().Err())
	})
}

func TestResponseWait(t *testing.T) {
	stallSettings(t, 5*time.Minute, 1<<20, 1<<20)
	assert.Equal(t, 5*time.Minute, responseWait(0))
	assert.Equal(t, 5*time.Minute, responseWait(1000))
	assert.Equal(t, 5*time.Minute+4*time.Second, responseWait(4<<20))
	assert.Equal(t, 5*time.Minute+4096*time.Second, responseWait(4<<30))

	// A size whose wait would overflow a duration gets the longest one.
	stallSettings(t, 5*time.Minute, 1<<20, 1)
	assert.Greater(t, responseWait(math.MaxInt64), time.Duration(math.MaxInt64)-time.Second)
}

func TestClient_DownloadStall(t *testing.T) {
	stallSettings(t, 300*time.Millisecond, 4, 1<<20)

	t.Run("BodyStalls", func(t *testing.T) {
		c := stallTestServer(t, func(w http.ResponseWriter, r *http.Request, release <-chan struct{}) {
			w.Header().Set("Content-Length", "64")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("abc"))
			w.(http.Flusher).Flush()
			<-release
		})
		dir := t.TempDir()
		start := time.Now()
		err := within(t, 5*time.Second, func() error { return c.Download("/stall.jpg", filepath.Join(dir, "stall.jpg"), false) })
		require.Error(t, err)
		assert.Contains(t, err.Error(), "stalled")
		assert.Less(t, time.Since(start), 5*time.Second)
		entries, readErr := os.ReadDir(dir)
		require.NoError(t, readErr)
		assert.Empty(t, entries)
	})
	t.Run("ResponseStalls", func(t *testing.T) {
		c := stallTestServer(t, func(w http.ResponseWriter, r *http.Request, release <-chan struct{}) {
			<-release
		})
		start := time.Now()
		dest := filepath.Join(t.TempDir(), "stall.jpg")
		err := within(t, 5*time.Second, func() error { return c.Download("/stall.jpg", dest, false) })
		require.Error(t, err)
		assert.Contains(t, err.Error(), "stalled")
		assert.Less(t, time.Since(start), 5*time.Second)
	})
	t.Run("SlowButSteady", func(t *testing.T) {
		// Sending a byte every 20ms takes longer than the timeout in total, which is fine while enough data moves.
		c := stallTestServer(t, func(w http.ResponseWriter, r *http.Request, release <-chan struct{}) {
			w.Header().Set("Content-Length", "24")
			w.WriteHeader(http.StatusOK)
			for _, b := range []byte("slow-content-slow-conten") {
				time.Sleep(20 * time.Millisecond)
				_, _ = w.Write([]byte{b})
				w.(http.Flusher).Flush()
			}
		})
		dest := filepath.Join(t.TempDir(), "slow.jpg")
		require.NoError(t, c.Download("/slow.jpg", dest, false))
		data, err := os.ReadFile(dest) //nolint:gosec // test-owned temporary file
		require.NoError(t, err)
		assert.Equal(t, "slow-content-slow-conten", string(data))
	})
	t.Run("Trickle", func(t *testing.T) {
		// A byte every 120ms is less data than the floor within the timeout.
		c := stallTestServer(t, func(w http.ResponseWriter, r *http.Request, release <-chan struct{}) {
			w.Header().Set("Content-Length", "64")
			w.WriteHeader(http.StatusOK)
			for i := 0; i < 64; i++ {
				select {
				case <-release:
					return
				case <-r.Context().Done():
					return
				case <-time.After(120 * time.Millisecond):
				}
				_, _ = w.Write([]byte("t"))
				w.(http.Flusher).Flush()
			}
		})
		dest := filepath.Join(t.TempDir(), "trickle.jpg")
		err := within(t, 5*time.Second, func() error { return c.Download("/trickle.jpg", dest, false) })
		require.Error(t, err)
		assert.Contains(t, err.Error(), "stalled")
	})
}

func TestClient_UploadStall(t *testing.T) {
	// The response wait grows by a second per 4 MiB of the body.
	stallSettings(t, 300*time.Millisecond, 64<<10, 4<<20)

	src := filepath.Join(t.TempDir(), "upload.jpg")
	require.NoError(t, os.WriteFile(src, bytes.Repeat([]byte("u"), 4<<20), 0o600))

	t.Run("ResponseStalls", func(t *testing.T) {
		c := stallTestServer(t, func(w http.ResponseWriter, r *http.Request, release <-chan struct{}) {
			_, _ = io.Copy(io.Discard, r.Body)
			<-release
		})
		start := time.Now()
		err := within(t, 5*time.Second, func() error { return c.Upload(src, "/stall.jpg") })
		require.Error(t, err)
		assert.Contains(t, err.Error(), "stalled")
		assert.Less(t, time.Since(start), 5*time.Second)
	})
	t.Run("ReadingStalls", func(t *testing.T) {
		c := stallTestServer(t, func(w http.ResponseWriter, r *http.Request, release <-chan struct{}) {
			<-release
		})
		err := within(t, 5*time.Second, func() error { return c.Upload(src, "/stall.jpg") })
		require.Error(t, err)
		assert.Contains(t, err.Error(), "stalled")
	})
	t.Run("SlowResponse", func(t *testing.T) {
		// The server answers after longer than the timeout, but within the wait for a body of this size.
		c := stallTestServer(t, func(w http.ResponseWriter, r *http.Request, release <-chan struct{}) {
			_, _ = io.Copy(io.Discard, r.Body)
			time.Sleep(600 * time.Millisecond)
			w.WriteHeader(http.StatusCreated)
		})
		require.NoError(t, within(t, 5*time.Second, func() error { return c.Upload(src, "/slow-response.jpg") }))
	})
	t.Run("SlowButSteady", func(t *testing.T) {
		// Reading in steps takes longer than the timeout in total, which is fine while data keeps moving. The body
		// is larger than the socket buffers, so the client keeps reading it while the server does.
		large := filepath.Join(t.TempDir(), "large.jpg")
		require.NoError(t, os.WriteFile(large, bytes.Repeat([]byte("l"), 32<<20), 0o600))
		c := stallTestServer(t, func(w http.ResponseWriter, r *http.Request, release <-chan struct{}) {
			buf := make([]byte, 512<<10)
			for {
				time.Sleep(10 * time.Millisecond)
				if _, err := io.ReadFull(r.Body, buf); err != nil {
					break
				}
			}
			w.WriteHeader(http.StatusCreated)
		})
		start := time.Now()
		require.NoError(t, c.Upload(large, "/slow.jpg"))
		assert.Greater(t, time.Since(start), 200*time.Millisecond)
	})
}

func TestClient_timeoutContext(t *testing.T) {
	for _, c := range []struct {
		name     string
		timeout  Timeout
		arg      time.Duration
		deadline time.Duration
	}{
		{"Low", TimeoutLow, 0, 30 * time.Second},
		{"Override", TimeoutLow, 2 * time.Minute, 2 * time.Minute},
		// Without a configured timeout, requests are still bounded.
		{"None", TimeoutNone, 0, MaxRequestDuration},
		{"NoOverride", TimeoutLow, -1, 0},
	} {
		t.Run(c.name, func(t *testing.T) {
			client, err := NewClient("http://127.0.0.1:1/", "", "", c.timeout, "")
			require.NoError(t, err)
			ctx, cancel := client.timeoutContext(c.arg)
			defer cancel()
			deadline, ok := ctx.Deadline()
			if c.deadline == 0 {
				assert.False(t, ok)
				return
			}
			require.True(t, ok)
			assert.WithinDuration(t, time.Now().Add(c.deadline), deadline, 5*time.Second)
		})
	}
}

func TestClient_RecoveredPanic(t *testing.T) {
	// A client without its HTTP clients panics on a transfer, which is reported briefly, with the stack in the debug log.
	endpoint, err := url.Parse("http://127.0.0.1:1/")
	require.NoError(t, err)
	c := &Client{ctx: context.Background(), endpoint: endpoint}

	logger, hook := logtest.NewNullLogger()
	logger.SetLevel(logrus.DebugLevel)
	prev := log
	log = logger
	t.Cleanup(func() { log = prev })

	// stackLogged reports whether the panic was logged as an error without its stack, and its stack at debug level.
	stackLogged := func(kind string) bool {
		var logged, stack bool
		for _, e := range hook.AllEntries() {
			switch {
			case e.Level == logrus.ErrorLevel && strings.Contains(e.Message, "panic while "+kind):
				logged = !strings.Contains(e.Message, "goroutine")
			case e.Level == logrus.DebugLevel && strings.Contains(e.Message, "panic while "+kind) && strings.Contains(e.Message, "goroutine"):
				stack = true
			}
		}
		return logged && stack
	}

	t.Run("Upload", func(t *testing.T) {
		src := filepath.Join(t.TempDir(), "panic.jpg")
		require.NoError(t, os.WriteFile(src, []byte("panic"), 0o600))
		err := c.Upload(src, "/panic.jpg")
		require.Error(t, err)
		assert.Equal(t, "webdav: failed to upload panic.jpg (unexpected error)", err.Error())
		assert.True(t, stackLogged("uploading"))
	})
	t.Run("Files", func(t *testing.T) {
		// Without a context, the request context cannot be created.
		_, err := (&Client{endpoint: endpoint}).Files("/panic", false)
		require.Error(t, err)
		assert.NotContains(t, err.Error(), "\n")
		assert.Contains(t, err.Error(), "failed to list files in")
		assert.Contains(t, err.Error(), "(unexpected error)")
		assert.True(t, stackLogged("listing files"))
	})
	t.Run("Download", func(t *testing.T) {
		err := c.Download("/panic.jpg", filepath.Join(t.TempDir(), "panic.jpg"), false)
		require.Error(t, err)
		assert.NotContains(t, err.Error(), "\n")
		assert.Contains(t, err.Error(), "unexpected error while downloading")
		assert.True(t, stackLogged("downloading"))
	})
}

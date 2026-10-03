package photoprism

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/ffmpeg/encode"
	"github.com/photoprism/photoprism/pkg/fs"
)

// transportRelease supplies a barrier that is also released if a test assertion stops the caller.
func transportRelease(t *testing.T) chan struct{} {
	t.Helper()
	release := make(chan struct{})
	t.Cleanup(func() {
		select {
		case <-release:
		default:
			close(release)
		}
	})
	return release
}

// awaitTransport waits for a finite test operation before reading its completion.
func awaitTransport(t *testing.T, call *transportCall) (string, error) {
	t.Helper()
	select {
	case <-call.done:
	case <-time.After(3 * time.Second):
		t.Fatal("transport operation did not complete")
	}
	return call.wait()
}

// TestTransportGroup_Start checks shared completions, independent targets and request ordering.
func TestTransportGroup_Start(t *testing.T) {
	t.Run("SharedSuccess", func(t *testing.T) {
		var group transportGroup
		var runs atomic.Int32
		release := transportRelease(t)
		request := transportRequest{source: "source.mts"}
		run := func() (string, error) { runs.Add(1); <-release; return "result.avc", nil }
		calls := make([]*transportCall, 6)
		for i := range calls {
			calls[i] = group.start("target", request, run)
		}
		close(release)
		for _, call := range calls {
			path, err := awaitTransport(t, call)
			require.NoError(t, err)
			assert.Equal(t, "result.avc", path)
		}
		assert.EqualValues(t, 1, runs.Load())
		group.mutex.Lock()
		assert.Empty(t, group.calls)
		group.mutex.Unlock()
	})
	t.Run("SharedFailureAndRetry", func(t *testing.T) {
		var group transportGroup
		var runs atomic.Int32
		failure := errors.New("example failure")
		release := transportRelease(t)
		run := func() (string, error) { runs.Add(1); <-release; return "", failure }
		first := group.start("target", transportRequest{}, run)
		second := group.start("target", transportRequest{}, run)
		close(release)
		for _, call := range []*transportCall{first, second} {
			path, err := awaitTransport(t, call)
			assert.Empty(t, path)
			assert.ErrorIs(t, err, failure)
		}
		assert.EqualValues(t, 1, runs.Load())
		retry := group.start("target", transportRequest{}, func() (string, error) { runs.Add(1); return "retry.avc", nil })
		path, err := awaitTransport(t, retry)
		require.NoError(t, err)
		assert.Equal(t, "retry.avc", path)
		assert.EqualValues(t, 2, runs.Load())
	})
	t.Run("IndependentTargets", func(t *testing.T) {
		var group transportGroup
		release := transportRelease(t)
		first := group.start("one", transportRequest{}, func() (string, error) { <-release; return "one.avc", nil })
		second := group.start("two", transportRequest{}, func() (string, error) { return "two.avc", nil })
		path, err := awaitTransport(t, second)
		require.NoError(t, err)
		assert.Equal(t, "two.avc", path)
		select {
		case <-first.done:
			t.Fatal("first operation must still be waiting")
		default:
		}
		close(release)
		_, err = awaitTransport(t, first)
		require.NoError(t, err)
	})
	t.Run("DifferentRequestsQueue", func(t *testing.T) {
		variants := []struct {
			name    string
			request transportRequest
		}{
			{"Source", transportRequest{source: "other.mts"}},
			{"Config", transportRequest{conf: &config.Config{}}},
			{"Encoder", transportRequest{encoder: encode.SoftwareAvc}},
			{"Exclude", transportRequest{exclude: "avi"}},
			{"NoMutex", transportRequest{noMutex: true}},
			{"Force", transportRequest{force: true}},
		}
		for _, variant := range variants {
			t.Run(variant.name, func(t *testing.T) {
				var group transportGroup
				var runs atomic.Int32
				release := transportRelease(t)
				entered := make(chan struct{})
				first := group.start("target", transportRequest{}, func() (string, error) { close(entered); <-release; return "first.avc", nil })
				<-entered
				run := func() (string, error) { runs.Add(1); return "second.avc", nil }
				second := group.start("target", variant.request, run)
				third := group.start("target", variant.request, run)
				select {
				case <-second.done:
					t.Fatal("different request ran before its predecessor")
				case <-time.After(30 * time.Millisecond):
				}
				close(release)
				_, err := awaitTransport(t, first)
				require.NoError(t, err)
				for _, call := range []*transportCall{second, third} {
					path, err := awaitTransport(t, call)
					require.NoError(t, err)
					assert.Equal(t, "second.avc", path)
				}
				assert.EqualValues(t, 1, runs.Load())
				group.mutex.Lock()
				assert.Empty(t, group.calls)
				group.mutex.Unlock()
			})
		}
	})
}

// TestTransportGroup_Execute verifies that exceptional completions release the destination.
func TestTransportGroup_Execute(t *testing.T) {
	t.Run("Panic", func(t *testing.T) {
		var group transportGroup
		release := transportRelease(t)
		run := func() (string, error) { <-release; panic("example panic") }
		first := group.start("target", transportRequest{}, run)
		second := group.start("target", transportRequest{}, run)
		close(release)
		for _, call := range []*transportCall{first, second} {
			assert.PanicsWithValue(t, "example panic", func() { _, _ = awaitTransport(t, call) })
		}
		retry := group.start("target", transportRequest{}, func() (string, error) { return "ready", nil })
		path, err := awaitTransport(t, retry)
		require.NoError(t, err)
		assert.Equal(t, "ready", path)
	})
	t.Run("Goexit", func(t *testing.T) {
		var group transportGroup
		call := group.start("target", transportRequest{}, func() (string, error) { runtime.Goexit(); return "", nil })
		_, err := awaitTransport(t, call)
		assert.ErrorIs(t, err, errTransportInterrupted)
		retry := group.start("target", transportRequest{}, func() (string, error) { return "ready", nil })
		_, err = awaitTransport(t, retry)
		require.NoError(t, err)
	})
}

// TestTransportCall_Wait returns a completed result without modifying its shared state.
func TestTransportCall_Wait(t *testing.T) {
	done := make(chan struct{})
	close(done)
	expected := errors.New("example result")
	call := &transportCall{done: done, path: "result.avc", err: expected}
	for range 2 {
		path, err := call.wait()
		assert.Equal(t, "result.avc", path)
		assert.ErrorIs(t, err, expected)
	}
}

// TestTransportRequest_Matches distinguishes the observed version from a reused pathname.
func TestTransportRequest_Matches(t *testing.T) {
	path := filepath.Join(t.TempDir(), "source.mts")
	require.NoError(t, os.WriteFile(path, []byte("first"), fs.ModeFile))
	first, err := os.Stat(path)
	require.NoError(t, err)
	again, err := os.Stat(path)
	require.NoError(t, err)
	a := transportRequest{source: path, info: first}
	assert.True(t, a.matches(transportRequest{source: path, info: again}))
	assert.False(t, a.matches(transportRequest{source: path}))
	require.NoError(t, os.WriteFile(path, []byte("different size"), fs.ModeFile))
	changed, err := os.Stat(path)
	require.NoError(t, err)
	assert.False(t, a.matches(transportRequest{source: path, info: changed}))
	require.NoError(t, os.WriteFile(path, []byte("first"), fs.ModeFile))
	require.NoError(t, os.Chtimes(path, first.ModTime(), first.ModTime().Add(time.Second)))
	changed, err = os.Stat(path)
	require.NoError(t, err)
	assert.False(t, a.matches(transportRequest{source: path, info: changed}))
	replacement := filepath.Join(filepath.Dir(path), "replacement")
	require.NoError(t, os.WriteFile(replacement, []byte("first"), fs.ModeFile))
	require.NoError(t, os.Chtimes(replacement, first.ModTime(), first.ModTime()))
	require.NoError(t, os.Rename(replacement, path))
	changed, err = os.Stat(path)
	require.NoError(t, err)
	assert.False(t, a.matches(transportRequest{source: path, info: changed}))
}

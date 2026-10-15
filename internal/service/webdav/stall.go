package webdav

import (
	"context"
	"errors"
	"io"
	"math"
	"sync/atomic"
	"time"
)

// Transfers have no total deadline, so large files on slow links finish, but end when too little data moves.
// These are variables, so tests can change them.
var (
	// transferStallTimeout ends an upload or download when less than transferStallBytes moved within it.
	transferStallTimeout = 5 * time.Minute
	// transferStallBytes is the data that must move to restart the timeout.
	transferStallBytes int64 = 1 << 20
	// transferResponseRate is the rate in bytes per second at which the response to an upload may lag behind its
	// body, as servers may scan or move a file before they answer.
	transferResponseRate int64 = 1 << 20
)

// SetTransferStallTimeout sets the stall timeout of uploads and downloads and returns a function that restores the
// previous one, so that tests in other packages can shorten it. It must not be called while a transfer runs.
func SetTransferStallTimeout(timeout time.Duration) (restore func()) {
	prev := transferStallTimeout
	transferStallTimeout = timeout

	return func() { transferStallTimeout = prev }
}

// responseWait returns how long to wait for the response once an upload body of the specified size was sent.
func responseWait(size int64) time.Duration {
	if size <= 0 || transferResponseRate <= 0 {
		return transferStallTimeout
	}

	// The wait is bounded so that the duration cannot overflow.
	secs := min(size/transferResponseRate, int64((math.MaxInt64-transferStallTimeout)/time.Second))

	return transferStallTimeout + time.Duration(secs)*time.Second
}

// stallWatch cancels its context when too little data moves within its timeout.
type stallWatch struct {
	ctx     context.Context
	cancel  context.CancelFunc
	timer   *time.Timer
	timeout time.Duration
	moved   atomic.Int64
	stalled atomic.Bool
	waiting atomic.Bool
	stopped atomic.Bool
}

// newStallWatch returns a stallWatch whose context is derived from parent; a timeout of zero or less disables it.
func newStallWatch(parent context.Context, timeout time.Duration) *stallWatch {
	ctx, cancel := context.WithCancel(parent)
	w := &stallWatch{ctx: ctx, cancel: cancel, timeout: timeout}

	if timeout > 0 {
		w.timer = time.AfterFunc(timeout, func() {
			w.stalled.Store(true)
			cancel()
		})
	}

	return w
}

// Context returns the context to use for the transfer.
func (w *stallWatch) Context() context.Context {
	return w.ctx
}

// Moved counts n transferred bytes and restarts the timeout once transferStallBytes moved since the last restart.
func (w *stallWatch) Moved(n int) {
	if n > 0 && w.moved.Add(int64(n)) >= transferStallBytes && !w.waiting.Load() {
		w.moved.Store(0)
		w.restart(w.timeout)
	}
}

// Await restarts the timeout with the specified wait, which data moving afterwards no longer shortens.
func (w *stallWatch) Await(wait time.Duration) {
	w.waiting.Store(true)
	w.restart(wait)
}

// restart restarts the timer unless the transfer stalled or ended.
func (w *stallWatch) restart(wait time.Duration) {
	if w.timer == nil || w.stalled.Load() || w.stopped.Load() {
		return
	}

	w.timer.Reset(wait)
}

// Stalled reports whether the timeout canceled the transfer.
func (w *stallWatch) Stalled() bool {
	return w.stalled.Load()
}

// Stop releases the timer and the context.
func (w *stallWatch) Stop() {
	w.stopped.Store(true)

	if w.timer != nil {
		w.timer.Stop()
	}

	w.cancel()
}

// Reader returns a reader that counts the data read from r, and awaits the response for the duration returned
// by onEOF once r is exhausted, if onEOF is not nil.
func (w *stallWatch) Reader(r io.Reader, onEOF func() time.Duration) io.Reader {
	return &stallReader{r: r, watch: w, onEOF: onEOF}
}

// stallReader counts the data it reads for a stallWatch.
type stallReader struct {
	r     io.Reader
	watch *stallWatch
	onEOF func() time.Duration
}

// Read reads from the underlying reader and counts the data read.
func (s *stallReader) Read(p []byte) (int, error) {
	n, err := s.r.Read(p)
	s.watch.Moved(n)

	if errors.Is(err, io.EOF) && s.onEOF != nil {
		s.watch.Await(s.onEOF())
		s.onEOF = nil
	}

	return n, err
}

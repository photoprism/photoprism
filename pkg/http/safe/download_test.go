package safe

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestSafeDownload_OK(t *testing.T) {
	ts := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, "hello")
	})
	dir := t.TempDir()
	dest := filepath.Join(dir, "ok.txt")
	if err := Download(dest, ts.URL, &Options{Timeout: 5 * time.Second, MaxSizeBytes: 1024, AllowPrivate: true}); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(dest) //nolint:gosec // test reads temp file
	if err != nil || string(b) != "hello" {
		t.Fatalf("unexpected content: %v %q", err, string(b))
	}
}

func TestSafeDownload_TooLarge(t *testing.T) {
	ts := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		// 2KiB
		_, _ = w.Write(make([]byte, 2048))
	})
	dir := t.TempDir()
	dest := filepath.Join(dir, "big.bin")
	if err := Download(dest, ts.URL, &Options{Timeout: 5 * time.Second, MaxSizeBytes: 1024, AllowPrivate: true}); err == nil {
		t.Fatalf("expected ErrSizeExceeded")
	}
}

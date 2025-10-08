package avatar

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/photoprism/photoprism/pkg/service/http/safe"
)

func TestSafeDownload_InvalidScheme(t *testing.T) {
	dir := t.TempDir()
	dest := filepath.Join(dir, "x")
	if err := SafeDownload(dest, "file:///etc/passwd", nil); err == nil {
		t.Fatal("expected error for invalid scheme")
	}
}

func TestSafeDownload_PrivateIPBlocked(t *testing.T) {
	dir := t.TempDir()
	dest := filepath.Join(dir, "x")
	if err := SafeDownload(dest, "http://127.0.0.1/test.png", nil); err == nil {
		t.Fatal("expected SSRF private IP block")
	}
}

func TestSafeDownload_MaxSizeExceeded(t *testing.T) {
	// Local server; allow private for test.
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")
		// 2MB body
		w.WriteHeader(http.StatusOK)
		buf := make([]byte, 2<<20)
		_, _ = w.Write(buf)
	}))
	defer ts.Close()

	dir := t.TempDir()
	dest := filepath.Join(dir, "big")
	err := SafeDownload(dest, ts.URL, &safe.Options{Timeout: 5 * time.Second, MaxSizeBytes: 1 << 20, AllowPrivate: true})
	if err == nil {
		t.Fatal("expected size exceeded error")
	}
	if _, statErr := os.Stat(dest); !os.IsNotExist(statErr) {
		t.Fatalf("expected no output file on error, got stat err=%v", statErr)
	}
}

func TestSafeDownload_Succeeds(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, "ok")
	}))
	defer ts.Close()

	dir := t.TempDir()
	dest := filepath.Join(dir, "ok")
	if err := SafeDownload(dest, ts.URL, &safe.Options{Timeout: 5 * time.Second, MaxSizeBytes: 1 << 20, AllowPrivate: true}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	b, err := os.ReadFile(dest)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if string(b) != "ok" {
		t.Fatalf("unexpected content: %q", string(b))
	}
}

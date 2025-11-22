package safe

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// Redirect to a private IP must be blocked when AllowPrivate=false.
func TestDownload_BlockRedirectToPrivate(t *testing.T) {
	// Public-looking server that redirects to 127.0.0.1
	redirectTarget := "http://127.0.0.1:65535/secret"
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, redirectTarget, http.StatusFound)
	}))
	defer ts.Close()

	dir := t.TempDir()
	dest := filepath.Join(dir, "out")
	err := Download(dest, ts.URL, &Options{Timeout: 5 * time.Second, MaxSizeBytes: 1 << 20, AllowPrivate: false})
	if err == nil {
		t.Fatalf("expected redirect SSRF to be blocked")
	}
	if _, statErr := os.Stat(dest); !os.IsNotExist(statErr) {
		t.Fatalf("expected no output file on error, got stat err=%v", statErr)
	}
}

// With AllowPrivate=true, redirects to a local httptest server should succeed.
func TestDownload_AllowRedirectToPrivate(t *testing.T) {
	// Local private target that serves content.
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, "ok")
	}))
	defer target.Close()

	// Public-looking server that redirects to the private target.
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, target.URL, http.StatusFound)
	}))
	defer ts.Close()

	dir := t.TempDir()
	dest := filepath.Join(dir, "ok")
	if err := Download(dest, ts.URL, &Options{Timeout: 5 * time.Second, MaxSizeBytes: 1 << 20, AllowPrivate: true}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	b, err := os.ReadFile(dest)
	if err != nil || string(b) != "ok" {
		t.Fatalf("unexpected content: %v %q", err, string(b))
	}
}

package safe

import (
	"io"
	"net"
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

func TestIsPrivateOrDisallowedIP(t *testing.T) {
	disallowed := []string{
		"0.0.0.0",          // 0.0.0.0/8 this network
		"0.1.2.3",          // 0.0.0.0/8 this network
		"10.0.0.1",         // RFC1918
		"100.64.0.1",       // CGNAT RFC6598
		"100.127.255.254",  // CGNAT upper bound
		"172.16.0.1",       // RFC1918
		"192.168.1.1",      // RFC1918
		"169.254.169.254",  // link-local / cloud metadata
		"127.0.0.1",        // loopback
		"224.0.0.1",        // multicast
		"fc00::1",          // IPv6 ULA
		"::1",              // IPv6 loopback
		"fe80::1",          // IPv6 link-local
		"::",               // IPv6 unspecified
		"::ffff:127.0.0.1", // IPv4-mapped loopback
		"::a00:1",          // IPv4-compatible
		"64:ff9b::a00:1",   // NAT64 of 10.0.0.1
		"64:ff9b::7f00:1",  // NAT64 of 127.0.0.1
		"64:ff9b:1::1",     // local-use NAT64
		"2002:a00:1::",     // 6to4 of 10.0.0.1
		"2002:7f00:1::1",   // 6to4 of 127.0.0.1
		"fec0::1",          // site-local
		"2001::1",          // Teredo
		"100::1",           // discard-only
		"ff02::1",          // multicast
		"240.0.0.1",        // reserved
		"255.255.255.255",  // broadcast
	}
	for _, s := range disallowed {
		ip := net.ParseIP(s)
		if ip == nil {
			t.Fatalf("failed to parse %q", s)
		}
		if !isPrivateOrDisallowedIP(ip) {
			t.Errorf("expected %q to be disallowed", s)
		}
	}

	allowed := []string{
		"8.8.8.8",              // public
		"100.63.255.255",       // just below CGNAT range
		"100.128.0.1",          // just above CGNAT range
		"1.1.1.1",              // public
		"2606:4700:4700::1111", // public IPv6
		"::ffff:8.8.8.8",       // IPv4-mapped public
		"64:ff9b::808:808",     // NAT64 of 8.8.8.8
		"2002:808:808::1",      // 6to4 of 8.8.8.8
		"2a00:1450:4001::1",    // public IPv6
	}
	for _, s := range allowed {
		ip := net.ParseIP(s)
		if ip == nil {
			t.Fatalf("failed to parse %q", s)
		}
		if isPrivateOrDisallowedIP(ip) {
			t.Errorf("expected %q to be allowed", s)
		}
	}

	if !isPrivateOrDisallowedIP(nil) {
		t.Error("expected nil IP to be disallowed")
	}

	if !isPrivateOrDisallowedIP(net.IP{1, 2, 3}) {
		t.Error("expected invalid IP to be disallowed")
	}

	if !isPrivateOrDisallowedIP(net.ParseIP("fe80::1").To16()) {
		t.Error("expected link-local IP to be disallowed")
	}
}

func TestSafeDownload_Status(t *testing.T) {
	ts := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		hijacker, ok := w.(http.Hijacker)
		if !ok {
			t.Fatal("response writer does not support hijacking")
		}
		conn, buf, err := hijacker.Hijack()
		if err != nil {
			t.Fatal(err)
		}
		defer conn.Close()
		_, _ = buf.WriteString("HTTP/1.1 404 Gone Fishing\r\nContent-Length: 0\r\nConnection: close\r\n\r\n")
		_ = buf.Flush()
	})
	dir := t.TempDir()
	err := Download(filepath.Join(dir, "missing.bin"), ts.URL, &Options{Timeout: 5 * time.Second, MaxSizeBytes: 1024, AllowPrivate: true})
	if err == nil || err.Error() != "unexpected status 404" {
		t.Fatalf("unexpected error: %v", err)
	}
}

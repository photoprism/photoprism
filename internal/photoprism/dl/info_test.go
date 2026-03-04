package dl

import "testing"

func TestNewExternalGetRequest(t *testing.T) {
	t.Run("ValidHttps", func(t *testing.T) {
		req, err := newExternalGetRequest("https://example.com/vision")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if req.URL.Host != "example.com" {
			t.Fatalf("expected host example.com, got %s", req.URL.Host)
		}
	})
	t.Run("RejectMissingHost", func(t *testing.T) {
		if _, err := newExternalGetRequest("/relative/path"); err == nil {
			t.Fatal("expected error for missing host")
		}
	})
	t.Run("RejectUnsupportedScheme", func(t *testing.T) {
		if _, err := newExternalGetRequest("file:///tmp/secret"); err == nil {
			t.Fatal("expected error for unsupported scheme")
		}
	})
}

func TestSanitizeDownloadPath(t *testing.T) {
	t.Run("AcceptRelativePath", func(t *testing.T) {
		got, err := sanitizeDownloadPath(" clips/test.mp4 ")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if got != "clips/test.mp4" {
			t.Fatalf("unexpected path: %s", got)
		}
	})
	t.Run("RejectParentTraversal", func(t *testing.T) {
		if _, err := sanitizeDownloadPath("../secrets.txt"); err == nil {
			t.Fatal("expected parent traversal error")
		}
	})
}

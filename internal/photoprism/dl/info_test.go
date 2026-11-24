package dl

import "testing"

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

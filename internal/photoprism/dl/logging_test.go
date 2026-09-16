package dl

import (
	"testing"

	"github.com/photoprism/photoprism/pkg/txt"
)

func TestRedactArgs(t *testing.T) {
	in := []string{"--add-header", "Authorization: Bearer secret", "--add-header", "Origin: https://example.com", "--other", "v"}
	out := redactArgs(in)
	if out[1] != "Authorization: "+txt.Masked {
		t.Fatalf("expected redaction for Authorization, got %q", out[1])
	}
	if out[3] != "Origin: "+txt.Masked {
		t.Fatalf("expected redaction for Origin, got %q", out[3])
	}
	if in[1] == out[1] {
		t.Fatalf("redaction should not modify input slice in-place")
	}
}

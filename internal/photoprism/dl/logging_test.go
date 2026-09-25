package dl

import (
	"testing"

	"github.com/stretchr/testify/assert"

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
	t.Run("UrlCredentials", func(t *testing.T) {
		in := []string{"--proxy", "https://alice:hunter2@proxy.example.com:8443", "--batch-file", "-"}
		out := redactArgs(in)
		assert.Equal(t, "https://alice:"+txt.Masked+"@proxy.example.com:8443", out[1])
		assert.Equal(t, "-", out[3])
		assert.Contains(t, in[1], "hunter2")
	})
	t.Run("UrlQueryToken", func(t *testing.T) {
		out := redactArgs([]string{"https://videos.example.com/v/1?token=capabilityT0ken&x=1"})
		assert.Equal(t, "https://videos.example.com/v/1?token="+txt.Masked+"&x=1", out[0])
	})
	t.Run("UnparsableUrl", func(t *testing.T) {
		out := redactArgs([]string{"https://alice:hunter2@exa mple.com/ ?a=b"})
		assert.NotContains(t, out[0], "hunter2")
	})
	t.Run("HeaderLastArgument", func(t *testing.T) {
		assert.Equal(t, []string{"-f", "best", "--add-header"}, redactArgs([]string{"-f", "best", "--add-header"}))
	})
	t.Run("RepeatedHeadersAndUri", func(t *testing.T) {
		in := []string{"--add-header", "X-A: 1", "--add-header", "X-B: 2", "https://alice:hunter2@h/v"}
		out := redactArgs(in)
		assert.Equal(t, []string{"--add-header", "X-A: " + txt.Masked, "--add-header", "X-B: " + txt.Masked,
			"https://alice:" + txt.Masked + "@h/v"}, out)
	})
	t.Run("PlainArguments", func(t *testing.T) {
		in := []string{"--cookies", "/tmp/cookies.txt", "-f", "bv*+ba/b"}
		assert.Equal(t, in, redactArgs(in))
	})
}

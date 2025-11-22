package dns

import (
	"testing"
)

func Test_IsDNSLabel(t *testing.T) {
	good := []string{"a", "node1", "pp-node-01", "n32", "a234567890123456789012345678901"}
	bad := []string{"", "A", "node_1", "-bad", "bad-", stringsRepeat("a", 33)}
	for _, s := range good {
		if !IsLabel(s) {
			t.Fatalf("expected valid label: %q", s)
		}
	}
	for _, s := range bad {
		if IsLabel(s) {
			t.Fatalf("expected invalid label: %q", s)
		}
	}
}

func Test_IsDNSDomain(t *testing.T) {
	good := []string{"example.dev", "sub.domain.dev", "a.b"}
	bad := []string{"localdomain", "localhost", "a", "EXAMPLE.com", "example.com", "invalid", "test", "x.local"}
	for _, s := range good {
		if !IsDomain(s) {
			t.Fatalf("expected valid domain: %q", s)
		}
	}
	for _, s := range bad {
		if IsDomain(s) {
			t.Fatalf("expected invalid domain: %q", s)
		}
	}
}

// helper: fast string repeat without importing strings just for tests
func stringsRepeat(s string, n int) string {
	b := make([]byte, 0, len(s)*n)
	for i := 0; i < n; i++ {
		b = append(b, s...)
	}
	return string(b)
}

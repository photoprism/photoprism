package dns

import "testing"

func TestIsLocalSuffix(t *testing.T) {
	cases := []struct {
		name   string
		input  string
		expect bool
	}{
		{"Local mDNS", "local", true},
		{"Nested local", "sub.local", true},
		{"Regular domain", "example.dev", false},
		{"Empty", "", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := IsLocalSuffix(tc.input); got != tc.expect {
				t.Fatalf("IsLocalSuffix(%q) = %v, want %v", tc.input, got, tc.expect)
			}
		})
	}
}

func TestIsLoopbackHost(t *testing.T) {
	cases := []struct {
		name   string
		input  string
		expect bool
	}{
		{"Empty", "", false},
		{"Localhost", "localhost", true},
		{"Mixed case host", "LOCALHOST", true},
		{"Loopback IPv4", "127.0.0.42", true},
		{"Loopback IPv6", "::1", true},
		{"Trim whitespace", " 127.0.0.1 ", true},
		{"Regular IPv4", "192.168.0.1", false},
		{"Regular host", "node.example", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := IsLoopbackHost(tc.input); got != tc.expect {
				t.Fatalf("IsLoopbackHost(%q) = %v, want %v", tc.input, got, tc.expect)
			}
		})
	}
}

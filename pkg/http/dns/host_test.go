package dns

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTrimBrackets(t *testing.T) {
	assert.Equal(t, "2001:db8::1", TrimBrackets("[2001:db8::1]"))
	assert.Equal(t, "2001:db8::1", TrimBrackets("2001:db8::1"))
	assert.Equal(t, "example.com", TrimBrackets("example.com"))
	assert.Equal(t, "[", TrimBrackets("["))
	assert.Equal(t, "", TrimBrackets("[]"))
	assert.Equal(t, "", TrimBrackets(""))
}

func TestJoinHostPort(t *testing.T) {
	t.Run("Hostname", func(t *testing.T) {
		assert.Equal(t, "mariadb:3306", JoinHostPort("mariadb", 3306))
	})
	t.Run("IPv4", func(t *testing.T) {
		assert.Equal(t, "0.0.0.0:2342", JoinHostPort("0.0.0.0", 2342))
	})
	t.Run("IPv6", func(t *testing.T) {
		assert.Equal(t, "[fd00::10]:3306", JoinHostPort("fd00::10", 3306))
		assert.Equal(t, "[::1]:2342", JoinHostPort("[::1]", 2342))
	})
	t.Run("RoundTrip", func(t *testing.T) {
		for _, host := range []string{"mariadb", "10.0.0.5", "fd00::10", "[fd00::10]"} {
			h, p, err := net.SplitHostPort(JoinHostPort(host, 3306))
			assert.NoError(t, err, host)
			assert.Equal(t, TrimBrackets(host), h)
			assert.Equal(t, "3306", p)
		}
	})
}

func TestBracketHost(t *testing.T) {
	assert.Equal(t, "[2001:db8::1]", BracketHost("2001:db8::1"))
	assert.Equal(t, "[2001:db8::1]", BracketHost("[2001:db8::1]"))
	assert.Equal(t, "photos.example.com", BracketHost("photos.example.com"))
	assert.Equal(t, "192.0.2.1", BracketHost("192.0.2.1"))
}

func TestJoinHostPort_Listen(t *testing.T) {
	hosts := []string{"127.0.0.1"}

	if probe, err := net.Listen("tcp", "[::1]:0"); err == nil {
		_ = probe.Close()
		hosts = append(hosts, "::1", "[::1]")
	}

	for _, host := range hosts {
		l, err := net.Listen("tcp", JoinHostPort(host, 0))

		if assert.NoError(t, err, host) {
			_ = l.Close()
		}
	}
}

func TestIsASCII(t *testing.T) {
	assert.True(t, IsASCII("photos.example.io:443"))
	assert.True(t, IsASCII(""))
	assert.False(t, IsASCII("photos.example.\u0130o"))
	assert.False(t, IsASCII("\u212a.example"))
}

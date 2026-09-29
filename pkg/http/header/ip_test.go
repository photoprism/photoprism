package header

import (
	"net"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIP(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		assert.Equal(t, "0.0.0.0", IP("", "0.0.0.0"))
	})
	t.Run("Unknown", func(t *testing.T) {
		assert.Equal(t, "0.0.0.0", IP("0.0.0.0", "0.0.0.0"))
	})
	t.Run("Localhost", func(t *testing.T) {
		assert.Equal(t, "127.0.0.1", IP("127.0.0.1", "0.0.0.0"))
	})
	t.Run("IPv6", func(t *testing.T) {
		assert.Equal(t, "2001:0:130f::9c0:876a:130b", IP("2001:0000:130F:0000:0000:09C0:876A:130B", "0.0.0.0"))
	})
	t.Run("IPv6", func(t *testing.T) {
		assert.Equal(t, "2001:0:130f::9c0:876a:130b", IP("    2001:0000:130F:0000:0000:09C0:876A:130B    ", "0.0.0.0"))
	})
	t.Run("PublicIPv4", func(t *testing.T) {
		assert.Equal(t, "8.8.8.8", IP("8.8.8.8", "0.0.0.0"))
	})
	t.Run("PrivateIPv4", func(t *testing.T) {
		assert.Equal(t, "192.168.1.128", IP("192.168.1.128", "0.0.0.0"))
	})
	t.Run("UUID", func(t *testing.T) {
		assert.Equal(t, "0.0.0.0", IP("123e4567-e89b-12d3-A456-426614174000", "0.0.0.0"))
	})
	t.Run("Hello", func(t *testing.T) {
		assert.Equal(t, "0.0.0.0", IP("Hello", "0.0.0.0"))
	})
	t.Run("Default", func(t *testing.T) {
		assert.Equal(t, "default", IP("Hello", "default"))
	})
	t.Run("EmptyDefault", func(t *testing.T) {
		assert.Equal(t, "", IP("Hello", ""))
	})
	t.Run("List", func(t *testing.T) {
		assert.Equal(t, "0.0.0.0", IP("198.51.100.7, 203.0.113.5", "0.0.0.0"))
		assert.Equal(t, "0.0.0.0", IP("2001:0db8:0000:0000:0000:0000:0000:0001, 203.0.113.5", "0.0.0.0"))
	})
	t.Run("Port", func(t *testing.T) {
		assert.Equal(t, "2001:db8::1", IP("[2001:db8::1]:5678", "0.0.0.0"))
		assert.Equal(t, "203.0.113.5", IP("203.0.113.5:5678", "0.0.0.0"))
	})
	t.Run("Zone", func(t *testing.T) {
		assert.Equal(t, "fe80::1", IP("fe80::1%2", "0.0.0.0"))
	})
	t.Run("Prefix", func(t *testing.T) {
		assert.Equal(t, "0.0.0.0", IP("2001:db8::1/64", "0.0.0.0"))
	})
}

func TestParseIP(t *testing.T) {
	t.Run("IPv4", func(t *testing.T) {
		ip, err := ParseIP("203.0.113.5")
		assert.NoError(t, err)
		assert.Equal(t, "203.0.113.5", ip)
	})
	t.Run("IPv6", func(t *testing.T) {
		ip, err := ParseIP(" 2001:0DB8:0000:0000:0000:0000:0000:0001 ")
		assert.NoError(t, err)
		assert.Equal(t, "2001:db8::1", ip)
	})
	t.Run("Bracketed", func(t *testing.T) {
		ip, err := ParseIP("[2001:db8::1]")
		assert.NoError(t, err)
		assert.Equal(t, "2001:db8::1", ip)
	})
	t.Run("BracketedPort", func(t *testing.T) {
		ip, err := ParseIP("[2001:db8::1]:5678")
		assert.NoError(t, err)
		assert.Equal(t, "2001:db8::1", ip)
	})
	t.Run("IPv4Port", func(t *testing.T) {
		ip, err := ParseIP("203.0.113.5:443")
		assert.NoError(t, err)
		assert.Equal(t, "203.0.113.5", ip)
	})
	t.Run("Zone", func(t *testing.T) {
		ip, err := ParseIP("[fe80::1%eth0]:80")
		assert.NoError(t, err)
		assert.Equal(t, "fe80::1", ip)
	})
	t.Run("Mapped", func(t *testing.T) {
		ip, err := ParseIP("::ffff:203.0.113.5")
		assert.NoError(t, err)
		assert.Equal(t, "203.0.113.5", ip)
	})
	t.Run("MappedFullLength", func(t *testing.T) {
		s := "0000:0000:0000:0000:0000:ffff:192.168.100.200"
		assert.Len(t, s, 45)
		ip, err := ParseIP(s)
		assert.NoError(t, err)
		assert.Equal(t, "192.168.100.200", ip)
	})
	t.Run("Invalid", func(t *testing.T) {
		for _, s := range []string{
			"",
			" ",
			"<nil>",
			"Hello",
			"198.51.100.7, 203.0.113.5",
			"198.51.100.7,203.0.113.5",
			"2001:0db8:0000:0000:0000:0000:0000:0001, 203.0.113.5",
			"2001:db8::5/64",
			"2001:db8::1/64",
			"203.0.113.0/24",
			"203.0.113.5:",
			"203.0.113.5:http",
			"203.0.113.5:123456",
			"[2001:db8::1]:",
			"[2001:db8::1",
			"2001:db8::1]",
			"[[2001:db8::1]]",
			"[203.0.113.5]x",
			"203.0.113.5%eth0",
			"01.2.3.4",
			"example.com",
			"example.com:80",
			"1.2.3.4.5",
			"2001:db8::1\x00",
			"0000:0000:0000:0000:0000:0000:0000:0000:0000:0000:0000:0000:0000:1",
			"fe80::1%" + strings.Repeat("a", 57),
			strings.Repeat(" ", 58) + "1.2.3.4",
		} {
			ip, err := ParseIP(s)
			assert.ErrorIs(t, err, ErrInvalidIP, s)
			assert.Empty(t, ip, s)
		}
	})
	t.Run("OutputUnchanged", func(t *testing.T) {
		for _, s := range []string{
			"0.0.0.0", "127.0.0.1", "255.255.255.255", "::", "::1", "::1.2.3.4", "::ffff:102:304",
			"2001:db8::", "2001:db8:0:0:1:0:0:1", "2001:0:130F:0:0:9C0:876A:130B", "64:ff9b::1.2.3.4", "fe80::1",
		} {
			ip, err := ParseIP(s)
			assert.NoError(t, err, s)
			assert.Equal(t, net.ParseIP(s).String(), ip, s)
		}
	})
}

func TestIsPort(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		assert.True(t, isPort("0"))
		assert.True(t, isPort("443"))
		assert.True(t, isPort("65535"))
	})
	t.Run("Invalid", func(t *testing.T) {
		assert.False(t, isPort(""))
		assert.False(t, isPort("123456"))
		assert.False(t, isPort("http"))
		assert.False(t, isPort("-1"))
		assert.False(t, isPort("4 3"))
	})
}

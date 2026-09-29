package header

import (
	"net"
	"net/netip"
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

func TestClientNetwork(t *testing.T) {
	t.Run("IPv4", func(t *testing.T) {
		assert.Equal(t, "203.0.113.5", ClientNetwork("203.0.113.5"))
		assert.Equal(t, "203.0.113.5", ClientNetwork("::ffff:203.0.113.5"))
		assert.Equal(t, "203.0.113.5", ClientNetwork("203.0.113.5:443"))
	})
	t.Run("GlobalIPv6", func(t *testing.T) {
		assert.Equal(t, "2001:db8:1:2::/64", ClientNetwork("2001:db8:1:2:a:b:c:d"))
		assert.Equal(t, "2001:db8:1:2::/64", ClientNetwork("2001:0db8:0001:0002:ffff:ffff:ffff:ffff"))
		assert.Equal(t, "2001:db8:1:2::/64", ClientNetwork("[2001:db8:1:2::1]:443"))
		assert.Equal(t, "2001:db8:1:3::/64", ClientNetwork("2001:db8:1:3::1"))
	})
	t.Run("Nat64", func(t *testing.T) {
		assert.Equal(t, "192.0.2.1", ClientNetwork("64:ff9b::c000:201"))
		assert.Equal(t, "198.51.100.7", ClientNetwork("64:ff9b::198.51.100.7"))
		assert.Equal(t, "64:ff9b::7f00:1", ClientNetwork("64:ff9b::127.0.0.1"))
		assert.Equal(t, "64:ff9b::ac11:1", ClientNetwork("64:ff9b::172.17.0.1"))
		assert.Equal(t, "64:ff9b::", ClientNetwork("64:ff9b::0.0.0.0"))
		assert.Equal(t, "64:ff9b::6440:1", ClientNetwork("64:ff9b::100.64.0.1"))
	})
	t.Run("Teredo", func(t *testing.T) {
		// Server 192.0.2.10, client 203.0.113.5 inverted in the last 32 bits, any flags and port.
		assert.Equal(t, "teredo:203.0.113.5", ClientNetwork("2001:0:c000:20a:8000:63bf:34ff:8efa"))
		assert.Equal(t, "teredo:203.0.113.5", ClientNetwork("2001:0:c000:20a:0:1:34ff:8efa"))
		assert.Equal(t, "teredo:0.0.0.0", ClientNetwork("2001:0:c000:20a:0:1:ffff:ffff"))
	})
	t.Run("6to4", func(t *testing.T) {
		assert.Equal(t, "6to4:203.0.113.5", ClientNetwork("2002:cb00:7105::1"))
		assert.Equal(t, "6to4:203.0.113.5", ClientNetwork("2002:cb00:7105:ffff::1"))
		assert.Equal(t, "6to4:127.0.0.1", ClientNetwork("2002:7f00:1::1"))
	})
	t.Run("PerAddress", func(t *testing.T) {
		assert.Equal(t, "fd00:1::2", ClientNetwork("fd00:1::2"))
		assert.Equal(t, "fd00:1::3", ClientNetwork("fd00:1::3"))
		assert.Equal(t, "fe80::1", ClientNetwork("fe80::1%eth0"))
		assert.Equal(t, "::1", ClientNetwork("::1"))
		assert.Equal(t, "::c000:201", ClientNetwork("::192.0.2.1"))
		assert.Equal(t, "64:ff9b:1::1", ClientNetwork("64:ff9b:1::1"))
	})
	t.Run("Invalid", func(t *testing.T) {
		assert.Equal(t, "", ClientNetwork(""))
		assert.Equal(t, "", ClientNetwork("2001:db8::/64"))
		assert.Equal(t, "", ClientNetwork("example.com"))
	})
}

func TestParseAddr(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		addr, err := parseAddr("[::ffff:203.0.113.5]:80")
		assert.NoError(t, err)
		assert.True(t, addr.Is4())
		assert.Equal(t, "203.0.113.5", addr.String())
	})
	t.Run("Invalid", func(t *testing.T) {
		addr, err := parseAddr("203.0.113.5, 198.51.100.7")
		assert.ErrorIs(t, err, ErrInvalidIP)
		assert.False(t, addr.IsValid())
	})
}

func TestIsGlobalIPv4(t *testing.T) {
	t.Run("Global", func(t *testing.T) {
		assert.True(t, isGlobalIPv4(netip.MustParseAddr("203.0.113.5")))
		assert.True(t, isGlobalIPv4(netip.MustParseAddr("8.8.8.8")))
	})
	t.Run("NotGlobal", func(t *testing.T) {
		for _, s := range []string{"0.0.0.0", "127.0.0.1", "10.0.0.1", "172.17.0.1", "192.168.1.1", "169.254.1.1", "100.64.0.1", "224.0.0.1", "255.255.255.255", "::1"} {
			assert.False(t, isGlobalIPv4(netip.MustParseAddr(s)), s)
		}
	})
}

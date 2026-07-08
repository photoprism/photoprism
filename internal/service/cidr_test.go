package service

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseCIDRs(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		cidrs, err := ParseCIDRs("")

		require.NoError(t, err)
		assert.Empty(t, cidrs)
	})
	t.Run("CIDRAndIP", func(t *testing.T) {
		cidrs, err := ParseCIDRs("172.18.0.0/16, 127.0.0.1")

		require.NoError(t, err)
		require.Len(t, cidrs, 2)
		assert.Equal(t, "172.18.0.0/16", cidrs[0].String())
		assert.Equal(t, "127.0.0.1/32", cidrs[1].String())
	})
	t.Run("Invalid", func(t *testing.T) {
		_, err := ParseCIDRs("not-a-cidr")
		require.Error(t, err)
	})
}

func TestIPAllowed(t *testing.T) {
	t.Run("NoCIDRConfigured", func(t *testing.T) {
		assert.True(t, IPAllowed(net.ParseIP("203.0.113.10"), nil))
		assert.True(t, IPAllowed(net.ParseIP("127.0.0.1"), nil))
		assert.True(t, IPAllowed(net.ParseIP("169.254.169.254"), nil))
	})
	t.Run("CIDRConfigured", func(t *testing.T) {
		cidrs, err := ParseCIDRs("172.18.0.0/16")
		require.NoError(t, err)

		assert.True(t, IPAllowed(net.ParseIP("172.18.0.10"), cidrs))
		assert.False(t, IPAllowed(net.ParseIP("192.168.1.10"), cidrs))
		assert.False(t, IPAllowed(net.ParseIP("127.0.0.1"), cidrs))
		assert.False(t, IPAllowed(net.ParseIP("169.254.169.254"), cidrs))
	})
	t.Run("LoopbackExplicitlyAllowed", func(t *testing.T) {
		cidrs, err := ParseCIDRs("127.0.0.0/8")
		require.NoError(t, err)
		assert.True(t, IPAllowed(net.ParseIP("127.0.0.1"), cidrs))
	})
	t.Run("LinkLocalAlwaysDeniedWhenCIDRConfigured", func(t *testing.T) {
		cidrs, err := ParseCIDRs("169.254.0.0/16")
		require.NoError(t, err)
		assert.False(t, IPAllowed(net.ParseIP("169.254.169.254"), cidrs))
	})
}

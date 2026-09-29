package cluster

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseCIDRs(t *testing.T) {
	t.Run("Single", func(t *testing.T) {
		prefixes, err := ParseCIDRs("10.0.0.0/8")
		require.NoError(t, err)
		require.Len(t, prefixes, 1)
		assert.Equal(t, "10.0.0.0/8", prefixes[0].String())
	})
	t.Run("List", func(t *testing.T) {
		prefixes, err := ParseCIDRs(" 10.0.0.0/8 ,fd00:10::/64")
		require.NoError(t, err)
		require.Len(t, prefixes, 2)
		assert.Equal(t, "fd00:10::/64", prefixes[1].String())
	})
	t.Run("Masked", func(t *testing.T) {
		prefixes, err := ParseCIDRs("10.1.2.3/8")
		require.NoError(t, err)
		assert.Equal(t, "10.0.0.0/8", prefixes[0].String())
	})
	t.Run("Mapped", func(t *testing.T) {
		prefixes, err := ParseCIDRs("::ffff:10.0.0.0/104, 0:0:0:0:0:ffff:0.0.0.0/96")
		require.NoError(t, err)
		assert.Equal(t, "10.0.0.0/8", prefixes[0].String())
		assert.Equal(t, "0.0.0.0/0", prefixes[1].String())
	})
	t.Run("LeadingZeros", func(t *testing.T) {
		prefixes, err := ParseCIDRs("10.0.0.0/08, fd00::/064, 192.0.2.0/0")
		require.NoError(t, err)
		assert.Equal(t, "10.0.0.0/8", prefixes[0].String())
		assert.Equal(t, "fd00::/64", prefixes[1].String())
		assert.Equal(t, "0.0.0.0/0", prefixes[2].String())
	})
	t.Run("Invalid", func(t *testing.T) {
		for _, value := range []string{"", " ", "10.0.0.0/8,", "10.0.0.0/", "10.0.0.0/0x8", "10.0.0.0/8,,fd00::/64", "10.0.0.0/8,garbage", "10.0.0.1", "10.0.0.0/33", "example.com/8"} {
			prefixes, err := ParseCIDRs(value)
			assert.ErrorIs(t, err, ErrInvalidCIDR, value)
			assert.Nil(t, prefixes, value)
		}
	})
}

func TestCIDRsContain(t *testing.T) {
	t.Run("Single", func(t *testing.T) {
		assert.True(t, CIDRsContain("10.0.0.0/8", "10.1.2.3"))
		assert.False(t, CIDRsContain("10.0.0.0/8", "192.0.2.1"))
	})
	t.Run("List", func(t *testing.T) {
		assert.True(t, CIDRsContain("10.0.0.0/8, fd00:10::/64", "fd00:10::5"))
		assert.True(t, CIDRsContain("10.0.0.0/8, fd00:10::/64", "10.0.0.1"))
		assert.False(t, CIDRsContain("10.0.0.0/8, fd00:10::/64", "fd00:11::5"))
	})
	t.Run("MappedAddress", func(t *testing.T) {
		assert.True(t, CIDRsContain("10.0.0.0/8", "::ffff:10.0.0.1"))
		assert.True(t, CIDRsContain("::ffff:10.0.0.0/104", "10.1.2.3"))
		assert.True(t, CIDRsContain("::ffff:10.0.0.0/104", "::ffff:10.1.2.3"))
		assert.False(t, CIDRsContain("::ffff:10.0.0.0/104", "192.0.2.1"))
	})
	t.Run("OtherFamily", func(t *testing.T) {
		assert.False(t, CIDRsContain("0.0.0.0/0", "fd00::1"))
		assert.False(t, CIDRsContain("::/0", "10.0.0.1"))
		assert.False(t, CIDRsContain("fd00::/8", "10.0.0.1"))
		assert.True(t, CIDRsContain("0.0.0.0/0", "203.0.113.5"))
		assert.True(t, CIDRsContain("::/0", "2001:db8::1"))
	})
	t.Run("InvalidList", func(t *testing.T) {
		assert.False(t, CIDRsContain("10.0.0.0/8,garbage", "10.0.0.1"))
		assert.False(t, CIDRsContain("", "10.0.0.1"))
	})
	t.Run("InvalidAddress", func(t *testing.T) {
		assert.False(t, CIDRsContain("0.0.0.0/0", ""))
		assert.False(t, CIDRsContain("0.0.0.0/0", "unknown"))
		assert.False(t, CIDRsContain("0.0.0.0/8", "0.0.0.0"))
		assert.False(t, CIDRsContain("::/0", "::"))
	})
}

func TestTrimPrefixBits(t *testing.T) {
	assert.Equal(t, "10.0.0.0/8", trimPrefixBits("10.0.0.0/08"))
	assert.Equal(t, "10.0.0.0/0", trimPrefixBits("10.0.0.0/00"))
	assert.Equal(t, "10.0.0.0/8", trimPrefixBits("10.0.0.0/8"))
	assert.Equal(t, "10.0.0.0/", trimPrefixBits("10.0.0.0/"))
	assert.Equal(t, "10.0.0.0", trimPrefixBits("10.0.0.0"))
}

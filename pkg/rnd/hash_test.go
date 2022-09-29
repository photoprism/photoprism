package rnd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsSHA(t *testing.T) {
	t.Run("SHA1", func(t *testing.T) {
		assert.False(t, IsSHA1(""))
		assert.True(t, IsSHA1("de9f2c7fd25e1b3afad3e85a0bd17d9b100db4b3"))
		assert.False(t, IsSHA1("de9f2c7fd25e1b3afad3e85a0bd17d9b100db4b3a"))
	})
	t.Run("SHA224", func(t *testing.T) {
		assert.False(t, IsSHA224(""))
		assert.True(t, IsSHA224("d14a028c2a3a2bc9476102bb288234c415a2b01f828ea62ac5b3e42f"))
		assert.False(t, IsSHA224("de9f2c7fd25e1b3afad3e85a0bd17d9b100db4b3"))
	})
	t.Run("SHA256", func(t *testing.T) {
		assert.False(t, IsSHA256(""))
		assert.True(t, IsSHA256("e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"))
		assert.False(t, IsSHA256("de9f2c7fd25e1b3afad3e85a0bd17d9b100db4b3a"))
	})
	t.Run("SHA384", func(t *testing.T) {
		assert.False(t, IsSHA384(""))
		assert.True(t, IsSHA384("38b060a751ac96384cd9327eb1b1e36a21fdb71114be07434c0cc7bf63f6e1da274edebfe76f65fbd51ad2f14898b95b"))
		assert.False(t, IsSHA384("de9f2c7fd25e1b3afad3e85a0bd17d9b100db4b3a"))
	})
	t.Run("SHA512", func(t *testing.T) {
		assert.False(t, IsSHA512(""))
		assert.True(t, IsSHA512("cf83e1357eefb8bdf1542850d66d8007d620e4050b5715dc83f4a921d36ce9ce47d0d13c5d85f2b0ff8318d2877eec2f63b931bd47417a81a538327af927da3e"))
		assert.False(t, IsSHA512("de9f2c7fd25e1b3afad3e85a0bd17d9b100db4b3a"))
	})
}

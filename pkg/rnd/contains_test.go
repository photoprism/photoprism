package rnd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContainsUID(t *testing.T) {
	assert.False(t, ContainsUID([]string{}, 'l'))
	assert.True(t, ContainsUID([]string{"ls6sg341wowuy3c2", "loj2qo0awowuy0c0"}, 'l'))
	assert.False(t, ContainsUID([]string{"ls6sg341wowuy3c2", "aoj2qo0awowuy0c0"}, 'l'))
	assert.True(t, ContainsUID([]string{"dafbfeb8-a129-4e7c-9cf0-e7996a701cdb"}, PrefixNone))
	assert.False(t, ContainsUID([]string{"dafbfeb8-a129-4e7c-9cf0-e7996a701cdb"}, 'l'))
	assert.False(t, ContainsUID([]string{"_"}, '_'))
	assert.False(t, ContainsUID([]string{""}, '_'))
}

func TestContainsType(t *testing.T) {
	t.Run("None", func(t *testing.T) {
		result, prefix := ContainsType([]string{})
		assert.Equal(t, TypeEmpty, result)
		assert.Equal(t, PrefixNone, prefix)
	})
	t.Run("Unknown", func(t *testing.T) {
		result, prefix := ContainsType([]string{"dgsgseh24t"})
		assert.Equal(t, TypeUnknown, result)
		assert.Equal(t, PrefixNone, prefix)
	})
	t.Run("LabelUID", func(t *testing.T) {
		result, prefix := ContainsType([]string{"ls6sg341wowuy3c2", "loj2qo0awowuy0c0"})
		assert.Equal(t, TypeUID, result)
		assert.Equal(t, byte('l'), prefix)
	})
	t.Run("MixedUID", func(t *testing.T) {
		result, prefix := ContainsType([]string{"ls6sg341wowuy3c2", "aoj2qo0awowuy0c0"})
		assert.Equal(t, TypeUID, result)
		assert.Equal(t, PrefixMixed, prefix)
	})
	t.Run("TypeUUID", func(t *testing.T) {
		result, prefix := ContainsType([]string{"dafbfeb8-a129-4e7c-9cf0-e7996a701cdb"})
		assert.Equal(t, TypeUUID, result)
		assert.Equal(t, PrefixNone, prefix)
	})
	t.Run("TypeSessionID", func(t *testing.T) {
		result, prefix := ContainsType([]string{"69be27ac5ca305b394046a83f6fda18167ca3d3f2dbe7ac0"})
		assert.Equal(t, TypeSessionID, result)
		assert.Equal(t, PrefixNone, prefix)
	})
	t.Run("TypeMD5", func(t *testing.T) {
		result, prefix := ContainsType([]string{"79054025255fb1a26e4bc422aef54eb4"})
		assert.Equal(t, TypeMD5, result)
		assert.Equal(t, PrefixNone, prefix)
	})
	t.Run("TypeSHA1", func(t *testing.T) {
		result, prefix := ContainsType([]string{"de9f2c7fd25e1b3afad3e85a0bd17d9b100db4b3"})
		assert.Equal(t, TypeSHA1, result)
		assert.Equal(t, PrefixNone, prefix)
	})
	t.Run("TypeSHA224", func(t *testing.T) {
		result, prefix := ContainsType([]string{"d14a028c2a3a2bc9476102bb288234c415a2b01f828ea62ac5b3e42f", "d14a028c2a3a2bc9476102bb288234c415a2b01f828ea62ac5b3e42a"})
		assert.Equal(t, TypeSHA224, result)
		assert.Equal(t, PrefixNone, prefix)
	})
	t.Run("TypeSHA256", func(t *testing.T) {
		result, prefix := ContainsType([]string{"a3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855", "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"})
		assert.Equal(t, TypeSHA256, result)
		assert.Equal(t, PrefixNone, prefix)
	})
	t.Run("TypeSHA384", func(t *testing.T) {
		result, prefix := ContainsType([]string{"18b060a751ac96384cd9327eb1b1e36a21fdb71114be07434c0cc7bf63f6e1da274edebfe76f65fbd51ad2f14898b95b", "38b060a751ac96384cd9327eb1b1e36a21fdb71114be07434c0cc7bf63f6e1da274edebfe76f65fbd51ad2f14898b95b"})
		assert.Equal(t, TypeSHA384, result)
		assert.Equal(t, PrefixNone, prefix)
	})
	t.Run("TypeSHA512", func(t *testing.T) {
		result, prefix := ContainsType([]string{"af83e1357eefb8bdf1542850d66d8007d620e4050b5715dc83f4a921d36ce9ce47d0d13c5d85f2b0ff8318d2877eec2f63b931bd47417a81a538327af927da3e", "cf83e1357eefb8bdf1542850d66d8007d620e4050b5715dc83f4a921d36ce9ce47d0d13c5d85f2b0ff8318d2877eec2f63b931bd47417a81a538327af927da3e"})
		assert.Equal(t, TypeSHA512, result)
		assert.Equal(t, PrefixNone, prefix)
	})
}

package rnd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUidType(t *testing.T) {
	t.Run("None", func(t *testing.T) {
		result, prefix := IdType("")
		assert.Equal(t, TypeEmpty, result)
		assert.Equal(t, PrefixNone, prefix)
	})
	t.Run("Unknown", func(t *testing.T) {
		result, prefix := IdType("lt9kr7ur57iru5i7uy3c2")
		assert.Equal(t, TypeUnknown, result)
		assert.Equal(t, PrefixNone, prefix)
	})
	t.Run("LabelUID", func(t *testing.T) {
		result, prefix := IdType("ls6sg1e1wowuy3c2")
		assert.Equal(t, TypeUID, result)
		assert.Equal(t, byte('l'), prefix)
	})
}

func TestType_Equal(t *testing.T) {
	assert.True(t, TypeSHA1.Equal("SHA1"))
	assert.False(t, TypeSHA1.Equal("SHA256"))
}

func TestType_NotEqual(t *testing.T) {
	assert.False(t, TypeSHA1.NotEqual("SHA1"))
	assert.True(t, TypeSHA1.NotEqual("SHA256"))
}

func TestType_EntityID(t *testing.T) {
	assert.True(t, TypeUID.EntityID())
	assert.False(t, TypeSHA384.EntityID())
}

func TestType_SessionID(t *testing.T) {
	assert.True(t, TypeSessionID.SessionID())
	assert.False(t, TypeRefID.SessionID())
}

func TestType_CrcToken(t *testing.T) {
	assert.True(t, TypeCrcToken.CrcToken())
	assert.False(t, TypeMixed.CrcToken())
}

func TestType_Hash(t *testing.T) {
	assert.True(t, TypeMD5.Hash())
	assert.True(t, TypeSHA384.Hash())
	assert.False(t, TypeUUID.Hash())
}

func TestType_SHA(t *testing.T) {
	assert.True(t, TypeSHA1.SHA())
	assert.True(t, TypeSHA384.SHA())
	assert.True(t, TypeSHA224.SHA())
	assert.False(t, TypeUID.SHA())
}

func TestType_Unknown(t *testing.T) {
	assert.True(t, TypeUnknown.Unknown())
	assert.False(t, TypeSHA384.Unknown())
	assert.False(t, TypeRefID.Unknown())

}

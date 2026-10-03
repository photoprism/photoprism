package rnd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsUUID(t *testing.T) {
	assert.True(t, IsUUID("dafbfeb8-a129-4e7c-9cf0-e7996a701cdb"))
	assert.True(t, IsUUID("6ba7b810-9dad-11d1-80b4-00c04fd430c8"))
	assert.False(t, IsUUID("55785BAC-9H4B-4747-B090-EE123FFEE437"))
	assert.True(t, IsUUID("550e8400-e29b-11d4-a716-446655440000"))
	assert.False(t, IsUUID("4B1FEF2D1CF4A5BE38B263E0637EDEAD"))
}

func TestSanitizeUUID(t *testing.T) {
	assert.Equal(t, "dafbfeb8-a129-4e7c-9cf0-e7996a701cdb", SanitizeUUID("  \"dafbfeb8-a129-4e7c-9cf0-e7996a701cdb\"  "))
	assert.Equal(t, "dafbfeb8-a129-4e7c-9cf0-e7996a701cdb", SanitizeUUID("  xmp:dafbfeb8-a129-4e7c-9cf0-e7996a701cdb  "))
	assert.Equal(t, "dafbfeb8-a129-4e7c-9cf0-e7996a701cdb", SanitizeUUID("dafbfeb8-a129-4e7c-9cf0-e7996a701cdb"))
	assert.Equal(t, "6ba7b810-9dad-11d1-80b4-00c04fd430c8", SanitizeUUID("6ba7b810-9dad-11d1-80b4-00c04fd430c8"))
	assert.Equal(t, "", SanitizeUUID("55785BAC-9H4B-4747-B090-EE123FFEE437"))
	assert.Equal(t, "550e8400-e29b-11d4-a716-446655440000", SanitizeUUID("550e8400-e29b-11d4-a716-446655440000"))
	assert.Equal(t, "", SanitizeUUID("4B1FEF2D1CF4A5BE38B263E0637EDEAD"))
	assert.Equal(t, "", SanitizeUUID(""))
}

func TestIsCanonicalUUID(t *testing.T) {
	t.Run("Generated", func(t *testing.T) {
		assert.True(t, IsCanonicalUUID(UUIDv7()))
		assert.True(t, IsCanonicalUUID(UUID()))
	})
	t.Run("NoSeparators", func(t *testing.T) {
		assert.True(t, IsUUID("111111111111111111111111111111111111"))
		assert.False(t, IsCanonicalUUID("111111111111111111111111111111111111"))
	})
	t.Run("SeparatorsOnly", func(t *testing.T) {
		assert.False(t, IsCanonicalUUID("------------------------------------"))
	})
	t.Run("MisplacedSeparator", func(t *testing.T) {
		assert.False(t, IsCanonicalUUID("0198-4c21e8773a2b734b8d0ed31ac0c19984"))
	})
	t.Run("Uppercase", func(t *testing.T) {
		assert.False(t, IsCanonicalUUID("019984c2-1E87-73a2-b734-b8d0ed31ac0c"))
	})
	t.Run("TooShort", func(t *testing.T) {
		assert.False(t, IsCanonicalUUID("019984c2-1e87-73a2-b734-b8d0ed31ac0"))
	})
	t.Run("Empty", func(t *testing.T) {
		assert.False(t, IsCanonicalUUID(""))
	})
}

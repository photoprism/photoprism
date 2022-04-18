package rnd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPPID(t *testing.T) {
	for n := 0; n < 5; n++ {
		uid := GenerateUID('x')
		t.Logf("id: %s", uid)
		assert.Equal(t, len(uid), 16)
	}
}

func BenchmarkPPID(b *testing.B) {
	for n := 0; n < b.N; n++ {
		GenerateUID('x')
	}
}

func TestIsPPID(t *testing.T) {
	prefix := byte('x')

	for n := 0; n < 10; n++ {
		id := GenerateUID(prefix)
		assert.True(t, EntityUID(id, prefix))
	}

	assert.True(t, EntityUID("lt9k3pw1wowuy3c2", 'l'))
	assert.False(t, EntityUID("lt9k3pw1wowuy3c2123", 'l'))
	assert.False(t, EntityUID("lt9k3pw1wowuy3c2123", 'l'))
	assert.False(t, EntityUID("lt9k3pw1AAA-owuy3c2123", 'l'))
	assert.False(t, EntityUID("", 'l'))
	assert.False(t, EntityUID("lt9k3pw1w  ?owuy  3c2123", 'l'))
}

func TestIsHex(t *testing.T) {
	assert.True(t, IsHex("dafbfeb8-a129-4e7c-9cf0-e7996a701cdb"))
	assert.True(t, IsHex("6ba7b810-9dad-11d1-80b4"))
	assert.False(t, IsHex("55785BAC-9A4B-4747-B090-GE123FFEE437"))
	assert.False(t, IsHex("550e8400-e29b-11d4-a716_446655440000"))
	assert.True(t, IsHex("4B1FEF2D1CF4A5BE38B263E0637EDEAD"))
	assert.False(t, IsHex(""))
}

func TestUniqueID(t *testing.T) {
	assert.True(t, ValidID("lt9k3pw1wowuy3c2", 'l'))
	assert.True(t, ValidID("dafbfeb8-a129-4e7c-9cf0-e7996a701cdb", 'l'))
	assert.True(t, ValidID("6ba7b810-9dad-11d1-80b4-00c04fd430c8", 'l'))
	assert.True(t, ValidID("55785BAC-9A4B-4747-B090-EE123FFEE437", 'l'))
	assert.True(t, ValidID("550e8400-e29b-11d4-a716-446655440000", 'l'))
	assert.False(t, ValidID("4B1FEF2D1CF4A5BE38B263E0637EDEAD", 'l'))
	assert.False(t, ValidID("123", '1'))
	assert.False(t, ValidID("_", '_'))
	assert.False(t, ValidID("", '_'))
}

func TestUniqueIDs(t *testing.T) {
	assert.True(t, ValidIDs([]string{"lt9k3pw1wowuy3c2", "ltxk3pwawowuy0c0"}, 'l'))
	assert.True(t, ValidIDs([]string{"dafbfeb8-a129-4e7c-9cf0-e7996a701cdb"}, 'l'))
	assert.False(t, ValidIDs([]string{"_"}, '_'))
	assert.False(t, ValidIDs([]string{""}, '_'))
}

func TestIsLowerAlnum(t *testing.T) {
	assert.False(t, IsAlnum("dafbfeb8-a129-4e7c-9cf0-e7996a701cdb"))
	assert.True(t, IsAlnum("dafbe7996a701cdb"))
	assert.False(t, IsAlnum(""))
}

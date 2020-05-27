package rnd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPPID(t *testing.T) {
	for n := 0; n < 5; n++ {
		uid := PPID('x')
		t.Logf("id: %s", uid)
		assert.Equal(t, len(uid), 16)
	}
}

func BenchmarkPPID(b *testing.B) {
	for n := 0; n < b.N; n++ {
		PPID('x')
	}
}

func TestIsPPID(t *testing.T) {
	prefix := byte('x')

	for n := 0; n < 10; n++ {
		id := PPID(prefix)
		assert.True(t, IsPPID(id, prefix))
	}

	assert.True(t, IsPPID("lt9k3pw1wowuy3c2", 'l'))
}

func TestIsUID(t *testing.T) {
	assert.True(t, IsUID("lt9k3pw1wowuy3c2", 'l'))
	// xmp.iid:dafbfeb8-a129-4e7c-9cf0-e7996a701cdb
	assert.True(t, IsUID("dafbfeb8-a129-4e7c-9cf0-e7996a701cdb", 'l'))
	assert.True(t, IsUID("6ba7b810-9dad-11d1-80b4-00c04fd430c8", 'l'))
	assert.True(t, IsUID("55785BAC-9A4B-4747-B090-EE123FFEE437", 'l'))
	assert.True(t, IsUID("550e8400-e29b-11d4-a716-446655440000", 'l'))
	assert.False(t, IsUID("4B1FEF2D1CF4A5BE38B263E0637EDEAD", 'l'))
	assert.False(t, IsUID("123", '1'))
	assert.False(t, IsUID("_", '_'))
	assert.False(t, IsUID("", '_'))
}

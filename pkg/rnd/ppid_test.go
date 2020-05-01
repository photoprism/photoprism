package rnd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPPID(t *testing.T) {
	for n := 0; n < 5; n++ {
		uuid := PPID('x')
		t.Logf("id: %s", uuid)
		assert.Equal(t, len(uuid), 16)
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

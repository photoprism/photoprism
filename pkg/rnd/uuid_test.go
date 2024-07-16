package rnd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUUID(t *testing.T) {
	for n := 0; n < 5; n++ {
		s := UUID()
		t.Logf("UUID %d: %s", n, s)
		assert.Equal(t, 36, len(s))
	}
}

func BenchmarkUUID(b *testing.B) {
	for n := 0; n < b.N; n++ {
		UUID()
	}
}

func TestState(t *testing.T) {
	for n := 0; n < 5; n++ {
		s := State()
		t.Logf("UUID %d: %s", n, s)
		assert.Equal(t, 36, len(s))
	}
}

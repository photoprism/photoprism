package rnd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUUID(t *testing.T) {
	for n := range 5 {
		s := UUID()
		t.Logf("UUID %d: %s", n, s)
		assert.Equal(t, 36, len(s))
	}
}

func BenchmarkUUID(b *testing.B) {
	for b.Loop() {
		UUID()
	}
}

func TestState(t *testing.T) {
	for n := range 5 {
		s := State()
		t.Logf("UUID %d: %s", n, s)
		assert.Equal(t, 36, len(s))
	}
}

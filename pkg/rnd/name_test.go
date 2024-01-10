package rnd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestName(t *testing.T) {
	assert.NotEmpty(t, Name())

	for n := 0; n < 10; n++ {
		s := Name()
		t.Logf("Name %d: %s", n, s)
		assert.NotEmpty(t, Name())
	}
}

func BenchmarkName(b *testing.B) {
	for n := 0; n < b.N; n++ {
		Name()
	}
}

func TestNameN(t *testing.T) {
	assert.NotEmpty(t, NameN(2))

	for n := 0; n < 10; n++ {
		s := NameN(n + 1)
		t.Logf("NameN %d: %s", n, s)
		assert.NotEmpty(t, Name())
	}
}

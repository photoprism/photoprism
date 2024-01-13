package rnd

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestName(t *testing.T) {
	name := Name()
	assert.NotEmpty(t, name)
	assert.Equal(t, 1, strings.Count(name, " "))

	for n := 0; n < 10; n++ {
		s := Name()
		t.Logf("Name %d: %s", n, s)
		assert.NotEmpty(t, s)
		assert.Equal(t, 1, strings.Count(s, " "))
	}
}

func BenchmarkName(b *testing.B) {
	for n := 0; n < b.N; n++ {
		Name()
	}
}

func TestNameN(t *testing.T) {
	name := NameN(2)
	assert.NotEmpty(t, name)
	assert.Equal(t, 1, strings.Count(name, " "))

	for n := 0; n < 10; n++ {
		s := NameN(n + 1)
		t.Logf("NameN %d: %s", n, s)
		assert.NotEmpty(t, s)
		assert.Equal(t, n, strings.Count(s, " "))
	}
}

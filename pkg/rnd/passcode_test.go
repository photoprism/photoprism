package rnd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGeneratePasscode(t *testing.T) {
	for n := 0; n < 10; n++ {
		s := GeneratePasscode()
		t.Logf("Passcode %d: %s", n, s)
		assert.Equal(t, 19, len(s))
	}
}

func BenchmarkGeneratePasscode(b *testing.B) {
	for n := 0; n < b.N; n++ {
		GeneratePasscode()
	}
}

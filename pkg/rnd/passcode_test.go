package rnd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGeneratePasscode(t *testing.T) {
	for n := 0; n < 10; n++ {
		code := GeneratePasscode()
		t.Logf("Passcode %d: %s", n, code)
		assert.Equal(t, 19, len(code))
	}
}

func BenchmarkGeneratePasscode(b *testing.B) {
	for n := 0; n < b.N; n++ {
		GeneratePasscode()
	}
}

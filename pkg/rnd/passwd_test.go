package rnd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGeneratePasswd(t *testing.T) {
	pw := GeneratePasswd()
	t.Logf("password: %s", pw)
	assert.Equal(t, 8, len(pw))
}

func BenchmarkGeneratePasswd(b *testing.B) {
	for n := 0; n < b.N; n++ {
		GeneratePasswd()
	}
}

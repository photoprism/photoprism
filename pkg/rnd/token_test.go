package rnd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRandomToken(t *testing.T) {
	t.Run("Size4", func(t *testing.T) {
		token := GenerateToken(4)
		assert.NotEmpty(t, token)
	})
	t.Run("Size8", func(t *testing.T) {
		token := GenerateToken(9)
		assert.NotEmpty(t, token)
	})
	t.Run("Log", func(t *testing.T) {
		for n := 0; n < 10; n++ {
			token := GenerateToken(8)
			t.Logf("%d: %s", n, token)
			assert.NotEmpty(t, token)
		}
	})
}

func BenchmarkGenerateToken4(b *testing.B) {
	for n := 0; n < b.N; n++ {
		GenerateToken(4)
	}
}

func BenchmarkGenerateToken3(b *testing.B) {
	for n := 0; n < b.N; n++ {
		GenerateToken(3)
	}
}

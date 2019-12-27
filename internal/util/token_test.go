package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRandomToken(t *testing.T) {
	t.Run("size 4", func(t *testing.T) {
		token := RandomToken(4)
		assert.NotEmpty(t, token)
	})
	t.Run("size 8", func(t *testing.T) {
		token := RandomToken(9)
		assert.NotEmpty(t, token)
	})
}

func TestRandomPassword(t *testing.T) {
	pw := RandomPassword()
	t.Logf("password: %s", pw)
	assert.Equal(t, 8, len(pw))
}

func BenchmarkRandomPassword(b *testing.B) {
	for n := 0; n < b.N; n++ {
		RandomPassword()
	}
}

func TestUUID(t *testing.T) {
	for n := 0; n < 5; n++ {
		uuid := UUID()
		t.Logf("token: %s", uuid)
		assert.Equal(t, 36, len(uuid))
	}
}

func BenchmarkUUID(b *testing.B) {
	for n := 0; n < b.N; n++ {
		UUID()
	}
}

func BenchmarkRandomToken4(b *testing.B) {
	for n := 0; n < b.N; n++ {
		RandomToken(4)
	}
}

func BenchmarkRandomToken3(b *testing.B) {
	for n := 0; n < b.N; n++ {
		RandomToken(3)
	}
}

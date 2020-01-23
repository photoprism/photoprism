package rnd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRandomToken(t *testing.T) {
	t.Run("size 4", func(t *testing.T) {
		token := Token(4)
		assert.NotEmpty(t, token)
	})
	t.Run("size 8", func(t *testing.T) {
		token := Token(9)
		assert.NotEmpty(t, token)
	})
}

func TestRandomPassword(t *testing.T) {
	pw := Password()
	t.Logf("password: %s", pw)
	assert.Equal(t, 8, len(pw))
}

func BenchmarkRandomPassword(b *testing.B) {
	for n := 0; n < b.N; n++ {
		Password()
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
		Token(4)
	}
}

func BenchmarkRandomToken3(b *testing.B) {
	for n := 0; n < b.N; n++ {
		Token(3)
	}
}

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

package rnd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSha224(t *testing.T) {
	t.Run("Nil", func(t *testing.T) {
		s := Sha224(nil)
		t.Logf("Sha224(nil): %s", s)
		assert.NotEmpty(t, s)
		assert.True(t, IsHex(s))
		assert.Equal(t, 56, len(s))

	})
	t.Run("HelloWorld", func(t *testing.T) {
		s := Sha224([]byte("hello world\n"))
		t.Logf("Sha224(HelloWorld): %s", s)
		assert.NotEmpty(t, s)
		assert.True(t, IsHex(s))
		assert.Equal(t, "95041dd60ab08c0bf5636d50be85fe9790300f39eb84602858a9b430", s)
		assert.Equal(t, 56, len(s))

	})
}

func TestSha256(t *testing.T) {
	t.Run("Nil", func(t *testing.T) {
		s := Sha256(nil)
		t.Logf("Sha256(nil): %s", s)
		assert.NotEmpty(t, s)
		assert.True(t, IsHex(s))
		assert.Equal(t, 64, len(s))

	})
	t.Run("HelloWorld", func(t *testing.T) {
		s := Sha256([]byte("hello world\n"))
		t.Logf("Sha256(HelloWorld): %s", s)
		assert.NotEmpty(t, s)
		assert.True(t, IsHex(s))
		assert.Equal(t, "a948904f2f0f479b8f8197694b30184b0d2ed1c1cd2a1ec0fb85d299a192a447", s)
		assert.Equal(t, 64, len(s))

	})
}

func TestSha512(t *testing.T) {
	t.Run("Nil", func(t *testing.T) {
		s := Sha512(nil)
		t.Logf("Sha512(nil): %s", s)
		assert.NotEmpty(t, s)
		assert.True(t, IsHex(s))
		assert.Equal(t, 128, len(s))

	})
	t.Run("HelloWorld", func(t *testing.T) {
		s := Sha512([]byte("hello world\n"))
		t.Logf("Sha512(HelloWorld): %s", s)
		assert.NotEmpty(t, s)
		assert.True(t, IsHex(s))
		assert.Equal(t, "db3974a97f2407b7cae1ae637c0030687a11913274d578492558e39c16c017de84eacdc8c62fe34ee4e12b4b1428817f09b6a2760c3f8a664ceae94d2434a593", s)
		assert.Equal(t, 128, len(s))

	})
}

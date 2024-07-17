package rnd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBase10(t *testing.T) {
	t.Run("10", func(t *testing.T) {
		s := Base10(10)
		t.Logf("Base10 (10 chars): %s", s)
		assert.NotEmpty(t, s)
		assert.True(t, IsRefID(s))
		assert.False(t, InvalidRefID(s))
		assert.Equal(t, 10, len(s))

		for n := 0; n < 10; n++ {
			s = Base10(10)
			t.Logf("Base10 %d: %s", n, s)
			assert.NotEmpty(t, s)
		}
	})
	t.Run("23", func(t *testing.T) {
		s := Base10(23)
		t.Logf("Base10 (23 chars): %s", s)
		assert.NotEmpty(t, s)
		assert.False(t, IsRefID(s))
		assert.True(t, InvalidRefID(s))
		assert.Equal(t, 23, len(s))
	})
}

func TestBase36(t *testing.T) {
	t.Run("10", func(t *testing.T) {
		s := Base36(10)
		t.Logf("Base36 (10 chars): %s", s)
		assert.NotEmpty(t, s)
		assert.True(t, IsRefID(s))
		assert.False(t, InvalidRefID(s))
		assert.Equal(t, 10, len(s))

		for n := 0; n < 10; n++ {
			s = Base36(10)
			t.Logf("Base36 %d: %s", n, s)
			assert.NotEmpty(t, s)
		}
	})
	t.Run("23", func(t *testing.T) {
		s := Base36(23)
		t.Logf("Base36 (23 chars): %s", s)
		assert.NotEmpty(t, s)
		assert.False(t, IsRefID(s))
		assert.True(t, InvalidRefID(s))
		assert.Equal(t, 23, len(s))
	})
}

func TestBase62(t *testing.T) {
	t.Run("10", func(t *testing.T) {
		for n := 0; n < 10; n++ {
			s := Base62(10)
			t.Logf("Base62 %d: %s", n, s)
			assert.NotEmpty(t, s)
		}
	})
	t.Run("23", func(t *testing.T) {
		s := Base62(23)
		t.Logf("Base62 (23 chars): %s", s)
		assert.NotEmpty(t, s)
		assert.False(t, IsRefID(s))
		assert.True(t, InvalidRefID(s))
		assert.Equal(t, 23, len(s))
	})
	t.Run("32", func(t *testing.T) {
		for n := 0; n < 10; n++ {
			s := Base62(32)
			t.Logf("Base62 (32 chars) %d: %s", n, s)
			assert.NotEmpty(t, s)
			assert.False(t, IsRefID(s))
			assert.True(t, InvalidRefID(s))
			assert.Equal(t, 32, len(s))
		}
	})
}

func TestCharset(t *testing.T) {
	t.Run("23", func(t *testing.T) {
		s := Charset(23, CharsetBase62)
		t.Logf("CharsetBase62 (23 chars): %s", s)
		assert.NotEmpty(t, s)
		assert.False(t, IsRefID(s))
		assert.True(t, InvalidRefID(s))
		assert.Equal(t, 23, len(s))
	})
	t.Run("0", func(t *testing.T) {
		s := Charset(0, CharsetBase62)
		t.Logf("CharsetBase62 (23 chars): %s", s)
		assert.Empty(t, s)
	})
	t.Run("5000", func(t *testing.T) {
		s := Charset(5000, CharsetBase62)
		t.Logf("CharsetBase62 (23 chars): %s", s)
		assert.NotEmpty(t, s)
		assert.False(t, IsRefID(s))
		assert.True(t, InvalidRefID(s))
		assert.Equal(t, 4096, len(s))
	})
}

func TestRandomToken(t *testing.T) {
	t.Run("Size4", func(t *testing.T) {
		s := Base36(4)
		assert.NotEmpty(t, s)
	})
	t.Run("Size8", func(t *testing.T) {
		s := Base36(9)
		assert.NotEmpty(t, s)
	})
	t.Run("Log", func(t *testing.T) {
		for n := 0; n < 10; n++ {
			s := Base36(8)
			t.Logf("%d: %s", n, s)
			assert.NotEmpty(t, s)
		}
	})
}

func BenchmarkGenerateToken4(b *testing.B) {
	for n := 0; n < b.N; n++ {
		Base36(4)
	}
}

func BenchmarkGenerateToken3(b *testing.B) {
	for n := 0; n < b.N; n++ {
		Base36(3)
	}
}

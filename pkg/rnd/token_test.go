package rnd

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBase10(t *testing.T) {
	t.Run("Ten", func(t *testing.T) {
		s := Base10(10)
		t.Logf("Base10 (10 chars): %s", s)
		assert.NotEmpty(t, s)
		assert.True(t, IsRefID(s))
		assert.False(t, InvalidRefID(s))
		assert.Equal(t, 10, len(s))

		for n := range 10 {
			s = Base10(10)
			t.Logf("Base10 %d: %s", n, s)
			assert.NotEmpty(t, s)
		}
	})
	t.Run("Num23", func(t *testing.T) {
		s := Base10(23)
		t.Logf("Base10 (23 chars): %s", s)
		assert.NotEmpty(t, s)
		assert.False(t, IsRefID(s))
		assert.True(t, InvalidRefID(s))
		assert.Equal(t, 23, len(s))
	})
}

func TestBase36(t *testing.T) {
	t.Run("Ten", func(t *testing.T) {
		s := Base36(10)
		t.Logf("Base36 (10 chars): %s", s)
		assert.NotEmpty(t, s)
		assert.True(t, IsRefID(s))
		assert.False(t, InvalidRefID(s))
		assert.Equal(t, 10, len(s))

		for n := range 10 {
			s = Base36(10)
			t.Logf("Base36 %d: %s", n, s)
			assert.NotEmpty(t, s)
		}
	})
	t.Run("Num23", func(t *testing.T) {
		s := Base36(23)
		t.Logf("Base36 (23 chars): %s", s)
		assert.NotEmpty(t, s)
		assert.False(t, IsRefID(s))
		assert.True(t, InvalidRefID(s))
		assert.Equal(t, 23, len(s))
	})
}

func TestBase62(t *testing.T) {
	t.Run("Ten", func(t *testing.T) {
		for n := range 10 {
			s := Base62(10)
			t.Logf("Base62 %d: %s", n, s)
			assert.NotEmpty(t, s)
		}
	})
	t.Run("Num23", func(t *testing.T) {
		s := Base62(23)
		t.Logf("Base62 (23 chars): %s", s)
		assert.NotEmpty(t, s)
		assert.False(t, IsRefID(s))
		assert.True(t, InvalidRefID(s))
		assert.Equal(t, 23, len(s))
	})
	t.Run("Num32", func(t *testing.T) {
		for n := range 10 {
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
	t.Run("Num23", func(t *testing.T) {
		s := Charset(23, CharsetBase62)
		t.Logf("CharsetBase62 (23 chars): %s", s)
		assert.NotEmpty(t, s)
		assert.False(t, IsRefID(s))
		assert.True(t, InvalidRefID(s))
		assert.Equal(t, 23, len(s))
	})
	t.Run("Zero", func(t *testing.T) {
		s := Charset(0, CharsetBase62)
		t.Logf("CharsetBase62 (23 chars): %s", s)
		assert.Empty(t, s)
	})
	t.Run("Num5000", func(t *testing.T) {
		s := Charset(5000, CharsetBase62)
		t.Logf("CharsetBase62 (23 chars): %s", s)
		assert.NotEmpty(t, s)
		assert.False(t, IsRefID(s))
		assert.True(t, InvalidRefID(s))
		assert.Equal(t, 4096, len(s))
	})
	t.Run("NoCharset", func(t *testing.T) {
		assert.Empty(t, Charset(8, ""))
	})
	t.Run("CharsetTooLong", func(t *testing.T) {
		assert.Empty(t, Charset(8, strings.Repeat("x", 257)))
	})
	t.Run("OnlyCharsetMembers", func(t *testing.T) {
		s := Charset(512, CharsetBase36)
		assert.Equal(t, 512, len(s))
		for _, c := range []byte(s) {
			assert.Contains(t, CharsetBase36, string(c))
		}
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
		for n := range 10 {
			s := Base36(8)
			t.Logf("%d: %s", n, s)
			assert.NotEmpty(t, s)
		}
	})
}

func BenchmarkGenerateToken4(b *testing.B) {
	for b.Loop() {
		Base36(4)
	}
}

func BenchmarkGenerateToken3(b *testing.B) {
	for b.Loop() {
		Base36(3)
	}
}

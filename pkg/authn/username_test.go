package authn

import (
	"strings"
	"testing"

	"github.com/photoprism/photoprism/pkg/clean"

	"github.com/stretchr/testify/assert"
)

func TestUsername(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		name := ""
		if s, err := Username(name); err != nil {
			assert.ErrorIs(t, err, ErrEmpty)
		} else {
			assert.Equal(t, name, s)
		}
	})
	t.Run("TooLong", func(t *testing.T) {
		name := "teeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeest"
		if s, err := Username(name); err != nil {
			assert.ErrorIs(t, err, ErrTooLong)
		} else {
			assert.Equal(t, name, s)
		}
	})
	t.Run("Admin", func(t *testing.T) {
		name := "admin"
		if s, err := Username(name); err != nil {
			t.Fatal(err)
		} else {
			assert.Equal(t, "admin", s)
		}
	})
	t.Run("Uppercase", func(t *testing.T) {
		name := "ADMIN"
		if s, err := Username(name); err != nil {
			t.Fatal(err)
		} else {
			t.Logf("Uppercase: %s -> %s", name, s)
			assert.Equal(t, "admin", s)
		}
	})
	t.Run("Visitor", func(t *testing.T) {
		s, err := Username("Visitor")
		assert.NoError(t, err)
		assert.Equal(t, "visitor", s)
	})
	t.Run("Asterisk", func(t *testing.T) {
		name := "*"
		if s, err := Username(name); err != nil {
			assert.ErrorIs(t, err, ErrInvalid)
		} else {
			assert.Equal(t, name, s)
		}
	})
	t.Run("ClientUID", func(t *testing.T) {
		name := "cs6sg6beu8nm9e6t"
		if s, err := Username(name); err != nil {
			assert.ErrorIs(t, err, ErrReserved)
		} else {
			assert.Equal(t, name, s)
		}
	})
}

func TestUsernameRejectsFieldSeparator(t *testing.T) {
	// clean.Username removes it, and this gate refuses any name it had to alter. The rule is
	// this function's, not the package's: clean.Handle folds the character instead.
	s, err := Username("ok" + string(clean.FieldSep) + "granted")

	assert.Equal(t, ErrInvalid, err)
	assert.NotContains(t, s, string(clean.FieldSep))
}

func TestIsASCIIUsername(t *testing.T) {
	t.Run("ASCII", func(t *testing.T) {
		assert.True(t, IsASCIIUsername("j.doe_1-x@example.com"))
	})
	t.Run("Empty", func(t *testing.T) {
		assert.True(t, IsASCIIUsername(""))
	})
	t.Run("Accent", func(t *testing.T) {
		assert.False(t, IsASCIIUsername("josé"))
	})
}

func TestNewUsername(t *testing.T) {
	t.Run("Accepted", func(t *testing.T) {
		for _, name := range []string{"jose", "Jose", "j.doe", "j_doe-1", "_jdoe", "name@example.com", "a!b^c[d]e\\f"} {
			s, err := NewUsername(name)
			assert.NoError(t, err, name)
			assert.Equal(t, clean.Username(name), s, name)
		}
	})
	t.Run("NotASCII", func(t *testing.T) {
		for name, sanitized := range map[string]string{
			"josé":         "josé",         // Precomposed accent.
			"jose\u0301":   "jose\u0301",   // Combining accent in NFD form.
			"straße":       "straße",       // Sharp s.
			"јоѕе":         "јоѕе",         // Cyrillic letters.
			"ЈОЅЕ":         "јоѕе",         // Cyrillic letters, uppercase.
			"山田":           "山田",           // CJK.
			"ｊｏｓｅ":         "ｊｏｓｅ",         // Full-width Latin.
			"\u212Aate":    "kate",         // Kelvin sign (U+212A).
			"zoë@test.com": "zoë@test.com", // Email with an accent.
		} {
			s, err := NewUsername(name)
			assert.ErrorIs(t, err, ErrNotASCII, name)
			assert.Equal(t, sanitized, s, name)
		}
	})
	t.Run("NotPortable", func(t *testing.T) {
		for _, name := range []string{"50%off", "a%b@example.com", "__bob", "_.bob", "corp\\__bob", "__bob@example.com", "a..b"} {
			_, err := NewUsername(name)
			assert.ErrorIs(t, err, ErrInvalid, name)
		}
	})
	t.Run("InvalidBeforeNotASCII", func(t *testing.T) {
		_, err := NewUsername("jos\u00e9\u200b")
		assert.ErrorIs(t, err, ErrInvalid)
	})
	t.Run("Empty", func(t *testing.T) {
		_, err := NewUsername("")
		assert.ErrorIs(t, err, ErrEmpty)
	})
	t.Run("Reserved", func(t *testing.T) {
		_, err := NewUsername("cs5cpu17n6gj2qo5")
		assert.ErrorIs(t, err, ErrReserved)
	})
}

func TestFoldUsername(t *testing.T) {
	t.Run("Accent", func(t *testing.T) {
		assert.Equal(t, "jose", FoldUsername("José"))
	})
	t.Run("CombiningAccent", func(t *testing.T) {
		assert.Equal(t, "jose", FoldUsername("jose\u0301"))
	})
	t.Run("SharpS", func(t *testing.T) {
		assert.Equal(t, "strasse", FoldUsername("straße"))
	})
	t.Run("Cyrillic", func(t *testing.T) {
		assert.Equal(t, "ivan", FoldUsername("Иван"))
	})
	t.Run("FullWidth", func(t *testing.T) {
		assert.Equal(t, "jose", FoldUsername("ｊｏｓｅ"))
	})
	t.Run("ZeroWidthSpace", func(t *testing.T) {
		assert.Equal(t, "jose", FoldUsername("jos\u200be"))
	})
	t.Run("ASCII", func(t *testing.T) {
		assert.Equal(t, "j.doe", FoldUsername("J.Doe"))
	})
	t.Run("Empty", func(t *testing.T) {
		assert.Equal(t, "", FoldUsername(""))
	})
	t.Run("TooLong", func(t *testing.T) {
		assert.Equal(t, "", FoldUsername(strings.Repeat("é", 200)))
	})
	t.Run("Reserved", func(t *testing.T) {
		assert.Equal(t, "", FoldUsername("cs5cpu17n6gj2q\u00f65"))
	})
	t.Run("NothingLeft", func(t *testing.T) {
		assert.Equal(t, "", FoldUsername("\u0301"))
	})
}

func TestPortableUsername(t *testing.T) {
	t.Run("Portable", func(t *testing.T) {
		for _, name := range []string{"jdoe", "j_doe", "_jdoe", "j__doe", "jane.doe@example.com", "corp\\jdoe", "a.b.c", "bob."} {
			assert.True(t, PortableUsername(name), name)
		}
	})
	t.Run("NotPortable", func(t *testing.T) {
		for _, name := range []string{"50%off", "a%b@example.com", "__bob", "_.bob", "corp\\_.bob", "__", "a..b", "a .b",
			"jane..doe@example.com", "a. b", "bob@a..example.com", "corp..x\\bob"} {
			assert.False(t, PortableUsername(name), name)
		}
	})
}

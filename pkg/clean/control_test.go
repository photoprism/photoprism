package clean

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStripControl(t *testing.T) {
	t.Run("Unchanged", func(t *testing.T) {
		assert.Equal(t, "", StripControl(""))
		assert.Equal(t, "Europe/Berlin", StripControl("Europe/Berlin"))
		assert.Equal(t, "/Fotos/Älter 2024 幸福", StripControl("/Fotos/Älter 2024 幸福"))
		assert.Zero(t, testing.AllocsPerRun(10, func() { _ = StripControl("Europe/Berlin") }))
	})
	t.Run("Whitespace", func(t *testing.T) {
		assert.Equal(t, "a b c d e f", StripControl("a\tb\nc\vd\fe\rf"))
	})
	t.Run("Removed", func(t *testing.T) {
		assert.Equal(t, "ab[31mcd", StripControl("a\x00b\x1b[31mc\x7fd"))
		assert.Equal(t, "", StripControl("\x00\x01\x1f\x7f"))
		assert.Equal(t, "Ä幸", StripControl("Ä\x00幸"))
	})
	t.Run("InvalidUTF8Kept", func(t *testing.T) {
		assert.Equal(t, "a\xffb", StripControl("a\xff\x00b"))
	})
}

// controlSample holds NUL, other C0 controls, whitespace controls, ESC, DEL and an invalid byte.
const controlSample = "Ab\x00c\x01d\te\nf\x1bg\x7fh\xffi"

func TestSanitizersControlTable(t *testing.T) {
	tests := []struct {
		name     string
		sanitize func(string) string
		expected string
	}{
		{"Name", Name, "Abcdefgh�i"},
		{"NameCapitalized", NameCapitalized, "Abcdefgh�i"},
		{"DlName", DlName, "Abcdefgh�i"},
		{"Type", Type, "Abcd e fghi"},
		{"TypeLower", TypeLower, "abcd e fghi"},
		{"ShortType", ShortType, "Abcd e f"},
		{"TypeUnicode", TypeUnicode, "Abcd e fgh�i"},
		{"Path", Path, "Abcdefgh\xffi"},
		{"UserPath", UserPath, "Abcdefgh\xffi"},
		{"FileName", FileName, "Abcdefgh\xffi"},
		{"Handle", Handle, "abcdefgh�i"},
		{"Username", Username, "abcdefgh�i"},
		{"Email", Email, ""},
		{"Role", Role, "abcdefgh�i"},
		{"Attr", Attr, "Abcd e fghi"},
		{"Auth", Auth, "Abcdefgh�i"},
		{"Passcode", Passcode, "abcdefghi"},
		{"Codec", Codec, "Abcdefghi"},
		{"ContentType", ContentType, "Abcd e fghi"},
		{"Header", Header, "Abcdefghi"},
		{"ID", func(s string) string { return string(ID(s)) }, "Abcdefghi"},
		{"UID", UID, "abcdefghi"},
		{"Token", Token, "Abcdefghi"},
		{"ShareToken", ShareToken, "Abcdefghi"},
		{"Scope", Scope, "abcd e fghi"},
		{"Log", Log, "'Ab[U+0000]c[U+0001]d e f[U+001B]g[U+007F]h[0xFF]i'"},
		{"DNSLabel", DNSLabel, "abc-d-e-f-g-h-i"},
		{"FieldName", FieldName, "Abcdefghi"},
		{"State", func(s string) string { return State(s, "de") }, "Abcdefgh�i"},
		{"StripControl", StripControl, "Abcd e fgh\xffi"},
		{"Unicode", Unicode, "Ab\x00c\x01d\te\nf\x1bg\x7fhi"},
		{"ASCII", ASCII, "Ab\x00c\x01d\te\nf\x1bg\x7fhi"},
		{"SlashPath", SlashPath, controlSample},
		{"SearchString", SearchString, "Abc\x01d\te\nf\x1bg\x7fhi"},
		{"SearchQuery", SearchQuery, "Abc\x01d\te\nf\x1bg\x7fhi"},
		{"SearchTerms", SearchTerms, "Abc\x01d\te\nf\x1bg\x7fhi"},
		{"JSON", JSON, controlSample},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.sanitize(controlSample))
		})
	}
}

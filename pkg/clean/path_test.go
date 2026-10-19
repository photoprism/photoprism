package clean

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPath(t *testing.T) {
	t.Run("ValidPath", func(t *testing.T) {
		assert.Equal(t, ProjectRoot, Path(ProjectRoot))
	})
	t.Run("ColonTildeDollar", func(t *testing.T) {
		assert.Equal(t, "~/a:b c", Path("~/a:b c"))
		assert.Equal(t, "a$b/$(c)", Path("a$b/$(c)"))
	})
	t.Run("ValidFile", func(t *testing.T) {
		assert.Equal(t, "filename.TXT", Path("filename.TXT"))
	})
	t.Run("TheQuickBrownFox", func(t *testing.T) {
		assert.Equal(t, "The quick brown fox.", Path("The quick brown fox."))
	})
	t.Run("FilenameTxt", func(t *testing.T) {
		assert.Equal(t, "filename.txt", Path("filename.txt"))
	})
	t.Run("Empty", func(t *testing.T) {
		assert.Equal(t, "", Path(""))
	})
	t.Run("Dot", func(t *testing.T) {
		assert.Equal(t, ".", Path("."))
	})
	t.Run("DotDot", func(t *testing.T) {
		assert.Equal(t, "", Path(".."))
	})
	t.Run("DotDotDot", func(t *testing.T) {
		assert.Equal(t, "", Path("..."))
	})
	t.Run("Replace", func(t *testing.T) {
		assert.Equal(t, "", Path("${https://<host>:<port>/<path>}"))
	})
	t.Run("ControlCharacter", func(t *testing.T) {
		assert.Equal(t, "filename.", Path("filename."+string(rune(127))))
	})
	t.Run("SpecialChars", func(t *testing.T) {
		assert.Equal(t, "filename.", Path("filename.?**"))
	})
	t.Run("IncludesHash", func(t *testing.T) {
		assert.Equal(t, "1970-05 (Tray #005) to 3-78 Person-Abbrev Place", Path("1970-05 (Tray #005) to 3-78 Person-Abbrev Place"))
	})
	t.Run("StartsWithHash", func(t *testing.T) {
		assert.Equal(t, "#2020", Path("#2020"))
	})
	t.Run("RealNamesKept", func(t *testing.T) {
		for _, s := range []string{
			"2026/Family \U0001F469\u200D\U0001F469\u200D\U0001F467",
			"2026/\U0001F9D1\U0001F3FD\u200D\U0001F680 Launch",
			"2026/\U0001F3F4\U000E0067\U000E0062\U000E0073\U000E0063\U000E0074\U000E007F Trip",
			"2026/\u2764\uFE0F",
			"2026/mi\u200Cxwaham",
			"2026/a\u00A0b",
			"Screenshot 2026-10-14 at 10.00.00\u202FAM.png",
			"2026/\u200Bzero\u2060width\uFEFF\u00ADsoft",
			"2026/\u200Eltr\u200Frtl",
			"2026/\uF022private",
			"2026/Gr\xfc\xdfe",
		} {
			assert.Equal(t, s, Path(s), "%+q", s)
		}
		assert.Zero(t, testing.AllocsPerRun(10, func() { _ = Path("2026/Family \U0001F469\u200D\U0001F469") }))
	})
	t.Run("ControlsRemoved", func(t *testing.T) {
		assert.Equal(t, "abcdefghij", Path("a\x00b\x1bc\x7fd\u0085e\u009bf\u2028g\u2029hi\tj"))
	})
	t.Run("CheckedAfterRemoval", func(t *testing.T) {
		assert.Equal(t, "", Path("a.*.b"))
		assert.Equal(t, "", Path("a.\x00.b"))
		assert.Equal(t, "", Path("a/\u202E/b"))
		assert.Equal(t, "", Path("*"))
	})
	t.Run("ASCIISpacesTrimmed", func(t *testing.T) {
		assert.Equal(t, "a b", Path("  a b \t"))
		assert.Equal(t, "\u00A0trail\u202F", Path(" \u00A0trail\u202F "))
	})
	t.Run("JoinedBytesRemoved", func(t *testing.T) {
		assert.Equal(t, "ab.jpg", Path("a\xe2\x80?\xaeb.jpg"))
		assert.Equal(t, "ab.jpg", UserPath("a\xe2\x80?\xaeb.jpg"))
		assert.Equal(t, "ab.jpg", FileName("a\xe2\x80?\xaeb.jpg"))
	})
	t.Run("BidiControlsRemoved", func(t *testing.T) {
		assert.Equal(t, "photoabc.jpg", Path("photo\u202Eabc.jpg"))
		assert.Equal(t, "abcdefghi", Path("a\u202Ab\u202Bc\u202Cd\u202De\u2066f\u2067g\u2068h\u2069i"))
	})
}

func TestUserPath(t *testing.T) {
	t.Run("ValidPath", func(t *testing.T) {
		assert.Equal(t, "go/src/github.com/photoprism/photoprism", UserPath("/go/src/github.com/photoprism/photoprism"))
	})
	t.Run("ValidFile", func(t *testing.T) {
		assert.Equal(t, "filename.TXT", UserPath("filename.TXT"))
	})
	t.Run("TheQuickBrownFox", func(t *testing.T) {
		assert.Equal(t, "The quick brown fox", UserPath("The quick brown fox."))
	})
	t.Run("FilenameTxt", func(t *testing.T) {
		assert.Equal(t, "filename.txt", UserPath("filename.txt"))
	})
	t.Run("Empty", func(t *testing.T) {
		assert.Equal(t, "", UserPath(""))
	})
	t.Run("Dot", func(t *testing.T) {
		assert.Equal(t, "", UserPath("."))
	})
	t.Run("DotDot", func(t *testing.T) {
		assert.Equal(t, "", UserPath(".."))
	})
	t.Run("DotDotDot", func(t *testing.T) {
		assert.Equal(t, "", UserPath("..."))
	})
	t.Run("Replace", func(t *testing.T) {
		assert.Equal(t, "", UserPath("${https://<host>:<port>/<path>}"))
	})
	t.Run("IncludesHash", func(t *testing.T) {
		assert.Equal(t, "1970-05 (Tray #005) to 3-78 Person-Abbrev Place", UserPath("1970-05 (Tray #005) to 3-78 Person-Abbrev Place"))
	})
	t.Run("StartsWithHash", func(t *testing.T) {
		assert.Equal(t, "#2020", UserPath("#2020"))
	})
	t.Run("RealNamesKept", func(t *testing.T) {
		assert.Equal(t, "2026/Family \U0001F469\u200D\U0001F469\u200D\U0001F467", UserPath("/2026/Family \U0001F469\u200D\U0001F469\u200D\U0001F467/"))
		assert.Equal(t, "Screenshots/10.00\u202FAM", UserPath("Screenshots/10.00\u202FAM"))
		assert.Equal(t, "2026/Gr\xfc\xdfe", UserPath("2026/Gr\xfc\xdfe"))
	})
	t.Run("Unclean", func(t *testing.T) {
		assert.Equal(t, "foo/bar/baz", UserPath("/foo/bar/baz/"))
		assert.Equal(t, "dirty/path", UserPath("/dirty/path/"))
		assert.Equal(t, "dev.txt", UserPath("dev.txt"))
		assert.Equal(t, "", UserPath("../hello/foo/bar/../todo.txt"))
		assert.Equal(t, "hello/foo/bar/todo.txt", UserPath(". ./hello/foo/bar/./todo.txt"))
		assert.Equal(t, "", UserPath("./hello/foo/./bar/. ./todo.txt"))
		assert.Equal(t, "", UserPath(".."))
		assert.Equal(t, "", UserPath("."))
		assert.Equal(t, "", UserPath("/"))
		assert.Equal(t, "", UserPath(""))
	})
}

func TestPathString(t *testing.T) {
	t.Run("Unchanged", func(t *testing.T) {
		assert.Equal(t, "", pathString("", "*"))
		assert.Equal(t, "a\u200Db\xff", pathString("a\u200Db\xff", "*"))
	})
	t.Run("Removed", func(t *testing.T) {
		assert.Equal(t, "ab", pathString("*a\x00b*", "*"))
		assert.Equal(t, "", pathString("\x00\u202E*", "*"))
		assert.Equal(t, "a\xffb", pathString("a\xff\x01b", ""))
	})
	t.Run("JoinedBytesRemoved", func(t *testing.T) {
		assert.Equal(t, "xy", pathString("x\xe2\x80?\xaey", "?"))
		assert.Equal(t, "", pathString("\xe2\x81{\xa6", "{"))
		assert.Equal(t, "", pathString("\xe2\x80|\xa8", "|"))
		assert.Equal(t, "", pathString("\xc2*\x85", "*"))
		assert.Equal(t, "", pathString("\xc2\x01\x9b", ""))
		assert.Equal(t, "", pathString("\xe2\xe2\x80\x01\xae\x80\x01\xae", ""))
	})
}

func TestUnsafePathRune(t *testing.T) {
	t.Run("Unsafe", func(t *testing.T) {
		for _, r := range []rune{0x00, 0x09, 0x0A, 0x1B, 0x1F, 0x7F, 0x80, 0x85, 0x9B, 0x9F,
			0x2028, 0x2029, 0x202A, 0x202E, 0x2066, 0x2069} {
			assert.True(t, unsafePathRune(r), "U+%04X", r)
		}
	})
	t.Run("Safe", func(t *testing.T) {
		for _, r := range []rune{' ', 'a', '*', 0xA0, 0xAD, 0x061C, 0x200B, 0x200C, 0x200D, 0x200E,
			0x200F, 0x2060, 0x202F, 0x206A, 0xFE0F, 0xFEFF, 0xF022, 0xFFFD, 0x1F3FD, 0xE0067, 0xE007F} {
			assert.False(t, unsafePathRune(r), "U+%04X", r)
		}
	})
}

func TestSlashPath(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		assert.Equal(t, "", SlashPath(""))
	})
	t.Run("TrimWhitespace", func(t *testing.T) {
		assert.Equal(t, "foo/bar", SlashPath("  foo/bar  "))
	})
	t.Run("TrimSlashes", func(t *testing.T) {
		assert.Equal(t, "foo/bar", SlashPath("/foo/bar/"))
	})
	t.Run("NormalizeBackslashes", func(t *testing.T) {
		assert.Equal(t, "foo/bar", SlashPath(`\foo\bar\`))
	})
	t.Run("RootOnly", func(t *testing.T) {
		assert.Equal(t, "", SlashPath("/"))
	})
}

func TestPortableName(t *testing.T) {
	t.Run("Portable", func(t *testing.T) {
		for _, name := range []string{"03 - Março", "#1 & Friends", "[2020] (copy)", "Ünïcödé 日本 😀", "Tom's Party",
			"a+b,c;d=e@f", "IMG_1234.JPG", "2020", "file.tar.gz", "\U0001F469\u200d\U0001F469\u200d\U0001F467", "a b", "", ".hidden", "@eaDir",
			"a!b^c`d", "a\u00a0b", "\u65e5\u672c\u3000\u65c5\u884c", "a\u200bb", "\ufeffa", "Screenshot 2026-10-14 at 10.00.00\u202fAM.png",
			"a\u200eb", "\ufffd"} {
			assert.True(t, PortableName(name), "%q", name)
		}
	})
	t.Run("ReservedSet", func(t *testing.T) {
		assert.Equal(t, `\|"?*%<>{}:~$`, portableReserved)
	})
	t.Run("Reserved", func(t *testing.T) {
		for _, c := range portableReserved {
			for _, name := range []string{string(c), string(c) + "name", "na" + string(c) + "me", "name" + string(c)} {
				assert.False(t, PortableName(name), "%q", name)
			}
		}
		for _, name := range []string{`a\b`, "a|b", `a"b`, "a?b", "a*b", "a%b", "a<b", "a>b", "a{b", "a}b", "x:y.jpg", "~home", "a~b", "$HOME", "a$(b)", "a${b}", "price$"} {
			assert.False(t, PortableName(name), "%q", name)
		}
	})
	t.Run("Unsafe", func(t *testing.T) {
		for _, name := range []string{"a\x00b", "a\tb", "a\nb", "a\x7fb", "a\u0085b", "a\u2028b", "a\u2029b", "a\u202eb", "a\u2066b", "\x1b[31m",
			"name\x00", "name\u202e", "name\u2069"} {
			assert.False(t, PortableName(name), "%q", name)
		}
	})
	t.Run("InvalidUTF8", func(t *testing.T) {
		for _, name := range []string{"\xff", "a\xc4\xd6b", "caf\xe9", "a\xed\xa0\x80b"} {
			assert.False(t, PortableName(name), "%q", name)
		}
	})
}

func TestUnportablePath(t *testing.T) {
	t.Run("Portable", func(t *testing.T) {
		for _, rel := range []string{"", "a", "2020/03 - Março/#1 & Friends.jpg", "a/b/c", "@eaDir/x.jpg", "__MACOSX/_.x.jpg"} {
			assert.Equal(t, "", UnportablePath(rel), rel)
		}
	})
	t.Run("NotPortable", func(t *testing.T) {
		assert.Equal(t, "a|b", UnportablePath("a|b"))
		assert.Equal(t, "2020/dir*", UnportablePath("2020/dir*/inner.jpg"))
		assert.Equal(t, "2020/x:y.jpg", UnportablePath("2020/x:y.jpg"))
		assert.Equal(t, "~home", UnportablePath("~home/a/b"))
		assert.Equal(t, "a/caf\xe9", UnportablePath("a/caf\xe9/b"))
	})
}

func TestUnsupportedPath(t *testing.T) {
	t.Run("Supported", func(t *testing.T) {
		for _, rel := range []string{"", "2019/10", "users/jane.doe", "a_b/c"} {
			assert.Equal(t, "", UnsupportedPath(rel), rel)
		}
	})
	t.Run("NotSupported", func(t *testing.T) {
		assert.Equal(t, "2019/a:b", UnsupportedPath("2019/a:b/c"))
		assert.Equal(t, "users/__bob", UnsupportedPath("users/__bob/x"))
		assert.Equal(t, "@eaDir", UnsupportedPath("@eaDir"))
		assert.Equal(t, "a/$x", UnsupportedPath("a/$x/b"))
		assert.Equal(t, "@x", UnsupportedPath("@x/a|b"))
	})
}

func TestRefusedPath(t *testing.T) {
	// refused rejects names that start with an x.
	refused := func(name string) bool { return strings.HasPrefix(name, "x") }

	assert.Equal(t, "", refusedPath("", refused))
	assert.Equal(t, "", refusedPath("a/b/c", refused))
	assert.Equal(t, "x", refusedPath("x/a", refused))
	assert.Equal(t, "a/xb", refusedPath("a/xb/c/xd", refused))
	assert.Equal(t, "a//xc", refusedPath("a//xc", refused))
	assert.Equal(t, "/x", refusedPath("/x/", refused))
	assert.Equal(t, "", refusedPath("/a/", refused))
	assert.Equal(t, "a/", refusedPath("a/", func(name string) bool { return name == "" }))
}

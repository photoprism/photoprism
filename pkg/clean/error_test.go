package clean

import (
	"errors"
	"fmt"
	iofs "io/fs"
	"net"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestErrorFull(t *testing.T) {
	t.Run("Nil", func(t *testing.T) {
		assert.Equal(t, "no error", ErrorFull(nil))
	})
	t.Run("Empty", func(t *testing.T) {
		assert.Equal(t, "unknown error", ErrorFull(errors.New("")))
	})
	t.Run("Simple", func(t *testing.T) {
		assert.Equal(t, "simple", ErrorFull(errors.New("simple")))
	})
	t.Run("Spaces", func(t *testing.T) {
		assert.Equal(t, "the quick brown fox", ErrorFull(errors.New("the quick brown fox")))
	})
	t.Run("Invalid", func(t *testing.T) {
		assert.Equal(t, "??https://?host?:?port?/?path??", ErrorFull(errors.New("${https://<host>:<port>/<path>}")))
	})
	t.Run("Quotes", func(t *testing.T) {
		assert.Equal(t, "the quick 'brown fox", ErrorFull(errors.New("the quick `brown fox")))
	})
	t.Run("Truncated", func(t *testing.T) {
		assert.Len(t, ErrorFull(errors.New(strings.Repeat("e", LengthLimit+100))), LengthLimit)
	})
	t.Run("PathIsKept", func(t *testing.T) {
		err := &iofs.PathError{Op: "open", Path: "/photoprism/storage/config/options.yml", Err: syscall.EACCES}
		assert.Equal(t, "open /photoprism/storage/config/options.yml: permission denied", ErrorFull(err))
	})
}

func TestErrorText(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		assert.Equal(t, "unknown error", errorText(""))
	})
	t.Run("Whitespace", func(t *testing.T) {
		assert.Equal(t, "unknown error", errorText("  \n\t "))
	})
	t.Run("Trimmed", func(t *testing.T) {
		assert.Equal(t, "failed", errorText("  failed  "))
	})
	t.Run("FieldSep", func(t *testing.T) {
		// A message is one field, so the separator a value interpolated into it carries is folded.
		assert.Equal(t, "album a?b not found", errorText("album a"+string(FieldSep)+"b not found"))
		assert.NotContains(t, Error(fmt.Errorf("album %s not found", "a"+string(FieldSep)+"b")), string(FieldSep))
		assert.NotContains(t, ErrorFull(fmt.Errorf("album %s not found", "a"+string(FieldSep)+"b")), string(FieldSep))
	})
	t.Run("Truncated", func(t *testing.T) {
		assert.Len(t, errorText(strings.Repeat("e", LengthLimit*2)), LengthLimit)
	})
	t.Run("TrimmedBeforeTruncated", func(t *testing.T) {
		// Trimming after truncation would return a message two characters short.
		assert.Len(t, errorText("  "+strings.Repeat("e", LengthLimit)), LengthLimit)
	})
	t.Run("Mapped", func(t *testing.T) {
		assert.Equal(t, "?a?b?c?d?e?f?'g'h", errorText("%a\\b$c<d>e{f}`g\"h"))
	})
	t.Run("NonPrintable", func(t *testing.T) {
		assert.Equal(t, "ab", errorText("a\x00\x7fb"))
	})
}

func TestErrorLocation(t *testing.T) {
	t.Run("AbsolutePath", func(t *testing.T) {
		assert.True(t, errorLocation("/photoprism/storage/config"))
	})
	t.Run("RelativePath", func(t *testing.T) {
		assert.True(t, errorLocation("config/options.yml"))
	})
	t.Run("WindowsPath", func(t *testing.T) {
		assert.True(t, errorLocation(`C:\photoprism\storage`))
	})
	t.Run("BareName", func(t *testing.T) {
		assert.False(t, errorLocation("options.yml"))
	})
	t.Run("ShortWord", func(t *testing.T) {
		assert.False(t, errorLocation("db"))
	})
	t.Run("SeparatorsOnly", func(t *testing.T) {
		assert.False(t, errorLocation("/"))
		assert.False(t, errorLocation("//"))
	})
	t.Run("DotsAndSeparatorsOnly", func(t *testing.T) {
		// Replacing these would match a path fragment of every message.
		assert.False(t, errorLocation("./"))
		assert.False(t, errorLocation("/."))
		assert.False(t, errorLocation("../"))
	})
}

func TestErrorPaths(t *testing.T) {
	const dir = "/photoprism/storage/config"

	t.Run("Nil", func(t *testing.T) {
		paths, complete := errorPaths(nil)
		assert.Empty(t, paths)
		assert.True(t, complete)
	})
	t.Run("LongestFirst", func(t *testing.T) {
		err := errors.Join(
			&iofs.PathError{Op: "mkdir", Path: dir, Err: syscall.EACCES},
			&iofs.PathError{Op: "open", Path: dir + "/options.yml", Err: syscall.EACCES},
		)
		paths, complete := errorPaths(err)
		assert.Equal(t, []string{dir + "/options.yml", dir}, paths)
		assert.True(t, complete)
	})
	t.Run("Deduplicated", func(t *testing.T) {
		err := errors.Join(
			&iofs.PathError{Op: "mkdir", Path: dir, Err: syscall.EACCES},
			&iofs.PathError{Op: "open", Path: dir, Err: syscall.EACCES},
		)
		paths, complete := errorPaths(err)
		assert.Equal(t, []string{dir}, paths)
		assert.True(t, complete)
	})
	t.Run("BoundedFanOut", func(t *testing.T) {
		// A shared subtree is visited once per reference, so the node budget ends the walk
		// and the result is reported incomplete.
		err := error(&iofs.PathError{Op: "open", Path: dir, Err: syscall.EACCES})
		for range 24 {
			err = errors.Join(err, err)
		}
		_, complete := errorPaths(err)
		assert.False(t, complete)
	})
	t.Run("TypedNilInChain", func(t *testing.T) {
		paths, complete := errorPaths(&wrappedError{msg: "outer", cause: (*iofs.PathError)(nil)})
		assert.Empty(t, paths)
		assert.True(t, complete)
	})
	t.Run("TypedNilOfAnotherType", func(t *testing.T) {
		// Any pointer type whose Unwrap dereferences its receiver reaches the same branch.
		for _, cause := range []error{(*os.SyscallError)(nil), (*wrappedError)(nil)} {
			paths, complete := errorPaths(&wrappedError{msg: "outer", cause: cause})
			assert.Empty(t, paths)
			assert.True(t, complete)
		}
	})
}

func TestError(t *testing.T) {
	const keyPath = "/photoprism/storage/config/keys/private-kid42.json"

	t.Run("Nil", func(t *testing.T) {
		assert.Equal(t, "no error", Error(nil))
	})
	t.Run("Empty", func(t *testing.T) {
		assert.Equal(t, "unknown error", Error(errors.New("")))
	})
	t.Run("NoPath", func(t *testing.T) {
		assert.Equal(t, "permission denied", Error(errors.New("permission denied")))
	})
	t.Run("PathError", func(t *testing.T) {
		err := &iofs.PathError{Op: "open", Path: keyPath, Err: syscall.EACCES}
		assert.Equal(t, "open ***: permission denied", Error(err))
		assert.NotContains(t, Error(err), "storage")
	})
	t.Run("WrappedPathError", func(t *testing.T) {
		err := fmt.Errorf("jwt: read signing key: %w", &iofs.PathError{Op: "open", Path: keyPath, Err: syscall.ENOENT})
		out := Error(err)
		assert.Equal(t, "jwt: read signing key: open ***: no such file or directory", out)
		assert.NotContains(t, out, "photoprism/storage")
	})
	t.Run("LinkError", func(t *testing.T) {
		out := Error(newLinkError())
		assert.Equal(t, "rename *** ***: invalid cross-device link", out)
		assert.NotContains(t, out, "options.yml")
	})
	t.Run("WrappedLinkError", func(t *testing.T) {
		out := Error(fmt.Errorf("config: %w", newLinkError()))
		assert.Equal(t, "config: rename *** ***: invalid cross-device link", out)
		assert.NotContains(t, out, "options.yml")
	})
	t.Run("JoinedErrors", func(t *testing.T) {
		err := errors.Join(
			&iofs.PathError{Op: "open", Path: keyPath, Err: syscall.ENOENT},
			&iofs.PathError{Op: "stat", Path: "/photoprism/storage/config/options.yml", Err: syscall.ENOENT},
		)
		out := Error(err)
		assert.NotContains(t, out, "storage")
		assert.Contains(t, out, "open ***")
		assert.Contains(t, out, "stat ***")
	})
	t.Run("NestedPathsShareAPrefix", func(t *testing.T) {
		// The longer path is replaced first, so the shorter one cannot leave a remainder.
		dir := "/photoprism/storage/config"
		err := errors.Join(
			&iofs.PathError{Op: "mkdir", Path: dir, Err: syscall.EACCES},
			&iofs.PathError{Op: "open", Path: dir + "/options.yml", Err: syscall.EACCES},
		)
		out := Error(err)
		assert.NotContains(t, out, "photoprism")
		assert.NotContains(t, out, "options.yml")
	})
	t.Run("BareNameIsKept", func(t *testing.T) {
		// A value without a separator names no location, and replacing it by substring
		// would also match the words around it.
		err := fmt.Errorf("unable to open the database db while reading dbconfig: %w",
			&iofs.PathError{Op: "open", Path: "db", Err: syscall.EACCES})
		assert.Equal(t, "unable to open the database db while reading dbconfig: open db: permission denied", Error(err))
	})
	t.Run("LongPathIsRemovedBeforeTruncation", func(t *testing.T) {
		// Replacement runs on the raw message, so a path longer than the length limit cannot
		// survive as a truncated prefix.
		long := "/photoprism/storage/" + strings.Repeat("d/", LengthLimit) + "keys"
		err := &iofs.PathError{Op: "open", Path: long, Err: syscall.EACCES}
		out := Error(err)
		assert.Equal(t, "open ***: permission denied", out)
		assert.NotContains(t, out, "photoprism")
	})
	t.Run("MessageIsStillSanitized", func(t *testing.T) {
		err := fmt.Errorf("%s: %w", strings.Repeat("e", LengthLimit),
			&iofs.PathError{Op: "open", Path: "/photoprism/<storage>", Err: syscall.EACCES})
		out := Error(err)
		assert.Len(t, out, LengthLimit)
		assert.NotContains(t, out, "photoprism")
	})
	t.Run("SelfWrappingError", func(t *testing.T) {
		// The cycle exhausts the budget, so the chain cannot be inspected completely.
		assert.Equal(t, errorOmitted, Error(&recursiveError{}))
		assert.Equal(t, "recursive", ErrorFull(&recursiveError{}))
	})
	t.Run("DeepChainIsDropped", func(t *testing.T) {
		// Every path must be found for replacement to mean anything, so a chain that cannot
		// be walked completely is not rendered at all.
		err := error(&iofs.PathError{Op: "open", Path: keyPath, Err: syscall.EACCES})
		for range errorUnwrapNodes + 1 {
			err = fmt.Errorf("layer: %w", err)
		}
		assert.Equal(t, errorOmitted, Error(err))
		assert.Contains(t, ErrorFull(err), keyPath)
	})
	t.Run("RepeatedPath", func(t *testing.T) {
		// A read and a close failure on the same file join to two errors naming one path,
		// which is collected once and has to be replaced at every occurrence.
		p := "/photoprism/storage/originals/2026/09/IMG_1234.jpg"
		err := errors.Join(
			&iofs.PathError{Op: "read", Path: p, Err: syscall.EIO},
			&iofs.PathError{Op: "close", Path: p, Err: os.ErrClosed},
		)
		out := Error(err)
		assert.NotContains(t, out, "photoprism")
		assert.NotContains(t, out, "IMG_1234")
		assert.Equal(t, 2, strings.Count(out, "***"))
	})
	t.Run("TypedNilInChain", func(t *testing.T) {
		assert.Equal(t, "outer", Error(&wrappedError{msg: "outer", cause: (*iofs.PathError)(nil)}))
	})
	t.Run("TypedNilOfAnotherType", func(t *testing.T) {
		assert.Equal(t, "outer", Error(&wrappedError{msg: "outer", cause: (*os.SyscallError)(nil)}))
	})
	t.Run("TypedNilWrappedByFmt", func(t *testing.T) {
		// fmt recovers the nil receiver's panic when it renders, so the chain walk is the
		// only place that meets the nil pointer.
		assert.Equal(t, "outer: ?nil?", Error(fmt.Errorf("outer: %w", error((*os.SyscallError)(nil)))))
	})
	t.Run("ShortPathIsKept", func(t *testing.T) {
		// A single-character path names no location, and removing it would match everywhere.
		err := &iofs.PathError{Op: "open", Path: ".", Err: syscall.ENOENT}
		assert.Equal(t, "open .: no such file or directory", Error(err))
	})
}

func TestErrorMatchesErrorFull(t *testing.T) {
	// Without a path to remove, both entry points must render identically. The cases that
	// carry a path-typed error exercise the branch where the walk runs but replaces nothing.
	cases := map[string]error{
		"Nil":               nil,
		"Empty":             errors.New(""),
		"Quotes":            errors.New("the quick `brown fox"),
		"Mapped":            errors.New("${https://<host>:<port>/<path>}"),
		"Padded":            errors.New("  failed  "),
		"Truncated":         errors.New(strings.Repeat("e", LengthLimit+100)),
		"MultiByte":         errors.New(strings.Repeat("ä", LengthLimit)),
		"WrappedNoPath":     fmt.Errorf("wrapped: %w", errors.New("cause")),
		"BareNamePath":      &iofs.PathError{Op: "open", Path: "options.yml", Err: syscall.ENOENT},
		"TrivialPath":       &iofs.PathError{Op: "stat", Path: ".", Err: syscall.ENOENT},
		"PathNotInMessage":  &wrappedError{msg: "outer", cause: &iofs.PathError{Op: "open", Path: "/a/b", Err: syscall.ENOENT}},
		"TypedNilInMessage": &wrappedError{msg: "outer", cause: (*iofs.PathError)(nil)},
	}

	for name, err := range cases {
		t.Run(name, func(t *testing.T) {
			require.Equal(t, ErrorFull(err), Error(err))
		})
	}
}

// newLinkError returns a rename failure naming two paths.
func newLinkError() *os.LinkError {
	return &os.LinkError{
		Op:  "rename",
		Old: "/photoprism/storage/config/options.yml.tmp",
		New: "/photoprism/storage/config/options.yml",
		Err: syscall.EXDEV,
	}
}

// recursiveError wraps itself, so that the unwrap bound is exercised.
type recursiveError struct{}

// Error returns the error message.
func (e *recursiveError) Error() string { return "recursive" }

// Unwrap returns the error itself.
func (e *recursiveError) Unwrap() error { return e }

// wrappedError reports a cause that its own message does not render.
type wrappedError struct {
	msg   string
	cause error
}

// Error returns the error message.
func (e *wrappedError) Error() string { return e.msg }

// Unwrap returns the cause.
func (e *wrappedError) Unwrap() error { return e.cause }

func TestErrorLocators(t *testing.T) {
	// Each of these renders a location through the error it carries, so Error collects it the way
	// it already collects a file path. The separator requirement in errorLocation is what keeps a
	// bare host or a host:port out, which the last two cases pin.
	t.Run("Url", func(t *testing.T) {
		err := &url.Error{Op: "Get", URL: "https://cdn.example.com/a/b.tar.gz", Err: errors.New("i/o timeout")}
		out := Error(err)
		assert.NotContains(t, out, "cdn.example.com")
		assert.Contains(t, out, errorPathPlaceholder)
		assert.Contains(t, ErrorFull(err), "cdn.example.com")
	})
	t.Run("Command", func(t *testing.T) {
		err := &exec.Error{Name: "/usr/local/bin/darktable-cli", Err: errors.New("no such file")}
		out := Error(err)
		assert.NotContains(t, out, "/usr/local/bin")
		assert.Contains(t, out, errorPathPlaceholder)
	})
	t.Run("EscapedSpelling", func(t *testing.T) {
		// Both render with %q, so a value holding a character Quote escapes reaches the message
		// in its escaped form and only matches when that form is collected too.
		u := &url.Error{Op: "Get", URL: `https://cdn.example.com/a?x=\y`, Err: errors.New("timeout")}
		assert.NotContains(t, Error(u), "cdn.example.com")
		e := &exec.Error{Name: `C:\Program Files\pp\ffmpeg.exe`, Err: errors.New("not found")}
		assert.NotContains(t, Error(e), "Program Files")
	})
	t.Run("UnixSocket", func(t *testing.T) {
		err := &net.OpError{Op: "dial", Net: "unix", Addr: &net.UnixAddr{Name: "/var/run/photoprism/vision.sock", Net: "unix"}, Err: errors.New("connection refused")}
		out := Error(err)
		assert.NotContains(t, out, "/var/run/photoprism")
		assert.Contains(t, out, errorPathPlaceholder)
	})
	t.Run("NetworkAddressIsNotALocation", func(t *testing.T) {
		// A host and port carry no separator, so the existing guard leaves them alone - they name
		// the peer, which the caller is already reporting.
		err := &net.OpError{Op: "write", Net: "tcp", Addr: &net.TCPAddr{IP: net.ParseIP("192.0.2.9"), Port: 443}, Err: errors.New("broken pipe")}
		assert.Contains(t, Error(err), "192.0.2.9:443")
	})
	t.Run("NilAddr", func(t *testing.T) {
		err := &net.OpError{Op: "read", Net: "tcp", Err: errors.New("broken pipe")}
		assert.NotEmpty(t, Error(err))
	})
	t.Run("Wrapped", func(t *testing.T) {
		// The locator is collected from anywhere in the chain, as a file path already is.
		inner := &url.Error{Op: "Get", URL: "https://cdn.example.com/a/b.tar.gz", Err: errors.New("refused")}
		assert.NotContains(t, Error(fmt.Errorf("download failed (%w)", inner)), "cdn.example.com")
	})
}

func TestAddrString(t *testing.T) {
	t.Run("Nil", func(t *testing.T) {
		assert.Empty(t, addrString(nil))
	})
	t.Run("TypedNil", func(t *testing.T) {
		// An interface holding a nil pointer is not nil, and String would dereference it.
		var a *net.UnixAddr
		assert.NotPanics(t, func() { assert.Empty(t, addrString(a)) })
	})
	t.Run("Set", func(t *testing.T) {
		assert.Equal(t, "/run/photoprism.sock", addrString(&net.UnixAddr{Name: "/run/photoprism.sock", Net: "unix"}))
	})
}

func TestQuotedInner(t *testing.T) {
	// An error rendering a value with %q carries the escaped spelling, which is what has to be
	// matched for the replacement to land.
	assert.Equal(t, `a/b`, quotedInner(`a/b`))
	assert.Equal(t, `C:\\x\\y`, quotedInner(`C:\x\y`))
	assert.Equal(t, `a\"b`, quotedInner(`a"b`))
}

func TestErrorFlattenedUrl(t *testing.T) {
	// A message can carry the same URL in two spellings: the one a producer rendered into its own
	// text and the one the wrapper holds, which net/http writes with the password replaced. Both
	// are covered, and because they reduce to the same text the location goes with them.
	u, parseErr := url.Parse("https://user:notreal@cdn.example.com/a/b.tar.gz")
	require.NoError(t, parseErr)

	wrapped := &url.Error{Op: "Get", URL: "https://user:***@cdn.example.com/a/b.tar.gz", Err: errors.New("i/o timeout")}
	err := fmt.Errorf("failed to download %s (%w)", u.String(), wrapped)

	t.Run("Error", func(t *testing.T) {
		out := Error(err)
		assert.NotContains(t, out, "notreal")
		assert.NotContains(t, out, "cdn.example.com")
		assert.Equal(t, 2, strings.Count(out, errorPathPlaceholder))
	})
	t.Run("ErrorFull", func(t *testing.T) {
		// The location is the detail an operator acts on; the credential is not.
		out := ErrorFull(err)
		assert.NotContains(t, out, "notreal")
		assert.Contains(t, out, "cdn.example.com")
		assert.Contains(t, out, UriRedactedValue)
	})
}

func TestErrorNil(t *testing.T) {
	// A non-nil interface may hold a nil pointer, and both spellings of absent render the same.
	var typed *iofs.PathError

	t.Run("TypedNil", func(t *testing.T) {
		assert.Equal(t, "no error", Error(typed))
		assert.Equal(t, "no error", ErrorFull(typed))
	})
	t.Run("Nil", func(t *testing.T) {
		assert.Equal(t, "no error", Error(nil))
		assert.Equal(t, "no error", ErrorFull(nil))
	})
	t.Run("Present", func(t *testing.T) {
		assert.NotEqual(t, "no error", Error(&iofs.PathError{Op: "open", Path: "/tmp/x", Err: os.ErrNotExist}))
	})
}

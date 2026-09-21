package clean

import (
	"os/exec"
	"strings"

	"github.com/photoprism/photoprism/pkg/txt"
)

// SecretMinLength is the shortest value Secrets replaces.
const SecretMinLength = 4

// Cmd returns a command and its arguments as a string for logging, with the specified values
// masked where they are a whole argument or the value a flag attaches, and with the credentials of
// every URI an argument holds removed. The result is not shell-quoted.
func Cmd(cmd *exec.Cmd, secrets ...string) string {
	if cmd == nil {
		return Empty
	}

	b := new(strings.Builder)
	b.WriteString(maskedArg(cmd.Path, secrets))

	if len(cmd.Args) > 1 {
		for _, arg := range cmd.Args[1:] {
			b.WriteString(Space)
			b.WriteString(maskedArg(arg, secrets))
		}
	}

	return b.String()
}

// Secrets returns text, such as the output of a command, with every occurrence of the specified
// values masked. One shorter than SecretMinLength is skipped, as replacing it by substring would
// leave the text unreadable. Cmd has no such limit, because an argument list is structured.
func Secrets(s string, secrets ...string) string {
	for _, secret := range secrets {
		if len(secret) < SecretMinLength {
			continue
		}

		s = strings.ReplaceAll(s, secret, txt.Masked)
	}

	return s
}

// maskedArg returns a command-line argument with a value it carries replaced: the whole argument,
// or what a flag attaches, replaced as a whole so a repetition leaves no copy. A value that ends a
// flag name is part of that name, so the name stays readable. An argument naming none of the
// values is still rendered through UriRedactedText, which reaches a credential a URI carries.
func maskedArg(arg string, secrets []string) string {
	value := flagValueIndex(arg)

	for _, secret := range secrets {
		switch {
		case secret == Empty:
			continue
		case arg == secret:
			return txt.Masked
		case value > 0 && strings.Contains(arg[value:], secret):
			return arg[:value] + txt.Masked
		}
	}

	return UriRedactedText(arg)
}

// flagValueIndex returns the index at which a flag argument attaches its value, either after an
// equals sign or after a short option name, and zero when the argument attaches none.
func flagValueIndex(arg string) int {
	if len(arg) < 3 || arg[0] != '-' {
		return 0
	}

	if i := strings.IndexByte(arg, '='); i > 0 {
		return i + 1
	} else if arg[1] != '-' {
		return 2
	}

	return 0
}

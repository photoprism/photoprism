package clean

import (
	"os/exec"
	"slices"
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
	attached := flagValueIndexes(arg)

	for _, secret := range secrets {
		if secret == Empty {
			continue
		} else if arg == secret {
			return txt.Masked
		}

		for _, value := range attached {
			if strings.Contains(arg[value:], secret) {
				return arg[:value] + txt.Masked
			}
		}
	}

	return UriRedactedText(arg)
}

// flagValueIndexes returns every index at which a flag argument may attach its value: after an
// equals sign, and after a short option name. Both are reported because "-pname=value" reads as
// either, and only the value being looked for says which it was. The equals sign comes first, so
// a long option keeps its name readable when both would match.
func flagValueIndexes(arg string) []int {
	var at []int

	if i := flagValueIndex(arg); i > 0 {
		at = append(at, i)
	}

	if len(arg) > 2 && arg[0] == '-' && arg[1] != '-' && arg[1] != '=' {
		if i := 2; !slices.Contains(at, i) {
			at = append(at, i)
		}
	}

	return at
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

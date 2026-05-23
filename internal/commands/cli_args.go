package commands

import (
	"fmt"
	"strings"

	"github.com/urfave/cli/v2"
)

// TrailingFlagToken returns the first arg after the leading positional that
// looks like a CLI flag (e.g. "--name" or "-r").
//
// urfave/cli v2 delegates flag parsing to the stdlib flag package, which
// stops at the first non-flag token; flags placed after a positional are
// silently ignored. Callers use this helper to reject such invocations
// instead of acting on a context with the supplied flag values dropped.
func TrailingFlagToken(ctx *cli.Context) string {
	for _, a := range ctx.Args().Tail() {
		if a == "--" {
			return ""
		}
		if strings.HasPrefix(a, "-") && len(a) > 1 {
			return a
		}
	}
	return ""
}

// RejectTrailingFlags returns a usage error when flag-like tokens appear after
// the leading positional argument, naming the first offending token so users
// can fix the invocation rather than seeing a silent no-op.
func RejectTrailingFlags(ctx *cli.Context) error {
	if tok := TrailingFlagToken(ctx); tok != "" {
		return fmt.Errorf("flag %q must appear before positional arguments", tok)
	}
	return nil
}

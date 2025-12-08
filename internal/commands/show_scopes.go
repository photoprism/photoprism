package commands

import (
	"fmt"

	"github.com/urfave/cli/v2"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/pkg/txt/report"
)

// ShowScopesCommand configures the CLI command that lists supported authorization scopes.
var ShowScopesCommand = &cli.Command{
	Name:   "scopes",
	Usage:  "Displays supported authorization scopes",
	Flags:  report.CliFlags,
	Action: showScopesAction,
}

// showScopesAction renders the list of supported authorization scopes.
func showScopesAction(ctx *cli.Context) error {
	rows, cols := scopesReport()
	format, formatErr := report.CliFormatStrict(ctx)
	if formatErr != nil {
		return formatErr
	}

	result, err := report.RenderFormat(rows, cols, format)
	fmt.Println(result)

	return err
}

// scopesReport returns the table rows and columns for the scope listing.
func scopesReport() (rows [][]string, cols []string) {
	cols = []string{"scope", "description"}

	rows = make([][]string, 0, len(acl.ScopeDescriptions))

	for _, scope := range []string{"*", acl.ScopeRead.String(), acl.ScopeWrite.String()} {
		if desc, ok := acl.ScopeDescriptions[scope]; ok {
			rows = append(rows, []string{scope, desc})
		}
	}

	for _, resource := range acl.ResourceNames {
		name := resource.String()
		desc, ok := acl.ScopeDescriptions[name]
		if !ok {
			// Should never happen, but keep the command resilient.
			desc = "Resource available for use in authorization scopes."
		}
		rows = append(rows, []string{name, desc})
	}

	return rows, cols
}

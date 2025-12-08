package report

import (
	"github.com/urfave/cli/v2"
)

// CliFormat resolves the report format from CLI flags, defaulting to the fallback format.
func CliFormat(ctx *cli.Context) Format {
	switch {
	case ctx.Bool("json"):
		return JSON
	case ctx.Bool("md"), ctx.Bool("markdown"):
		return Markdown
	case ctx.Bool("tsv"):
		return TSV
	case ctx.Bool("csv"):
		return CSV
	default:
		return Default
	}
}

// CliFormatStrict selects a single output format from flags and returns
// a usage error (exit code 2) if multiple format flags are provided.
func CliFormatStrict(ctx *cli.Context) (Format, error) {
	count := 0
	if ctx.Bool("json") {
		count++
	}
	if ctx.Bool("md") || ctx.Bool("markdown") {
		count++
	}
	if ctx.Bool("tsv") {
		count++
	}
	if ctx.Bool("csv") {
		count++
	}
	if count > 1 {
		return Default, cli.Exit("choose exactly one output format: --json | --md | --csv | --tsv", 2)
	}
	return CliFormat(ctx), nil
}

// CliFlags registers common format selection flags.
var CliFlags = []cli.Flag{
	&cli.BoolFlag{
		Name:    "json",
		Aliases: []string{"j"},
		Usage:   "print machine-readable JSON",
	},
	&cli.BoolFlag{
		Name:    "md",
		Aliases: []string{"m"},
		Usage:   "print machine-readable Markdown",
	},
	&cli.BoolFlag{
		Name:    "csv",
		Aliases: []string{"c"},
		Usage:   "print semicolon separated values",
	},
	&cli.BoolFlag{
		Name:    "tsv",
		Aliases: []string{"t"},
		Usage:   "print tab separated values",
	},
}

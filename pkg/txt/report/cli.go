package report

import "github.com/urfave/cli/v2"

func CliFormat(ctx *cli.Context) Format {
	switch {
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

var CliFlags = []cli.Flag{
	&cli.BoolFlag{
		Name:    "md",
		Aliases: []string{"m"},
		Usage:   "format as machine-readable Markdown",
	},
	&cli.BoolFlag{
		Name:    "csv",
		Aliases: []string{"c"},
		Usage:   "export as semicolon separated values",
	},
	&cli.BoolFlag{
		Name:    "tsv",
		Aliases: []string{"t"},
		Usage:   "export as tab separated values",
	},
}

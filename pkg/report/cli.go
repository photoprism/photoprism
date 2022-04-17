package report

import "github.com/urfave/cli"

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
	cli.BoolFlag{
		Name:  "md, m",
		Usage: "format as machine-readable Markdown",
	},
	cli.BoolFlag{
		Name:  "csv, c",
		Usage: "export as semicolon separated values",
	},
	cli.BoolFlag{
		Name:  "tsv, t",
		Usage: "export as tab separated values",
	},
}

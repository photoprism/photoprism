package test

import (
	"flag"

	"github.com/urfave/cli"
)

// Returns example cli context for testing
func CliContext() *cli.Context {
	globalSet := flag.NewFlagSet("test", 0)
	globalSet.Bool("debug", false, "doc")
	globalSet.String("config-file", ConfigFile, "doc")
	globalSet.String("assets-path", AssetsPath, "doc")
	globalSet.String("originals-path", OriginalsPath, "doc")
	globalSet.String("darktable-cli", DarktableCli, "doc")

	app := cli.NewApp()

	c := cli.NewContext(app, globalSet, nil)

	c.Set("config-file", ConfigFile)
	c.Set("assets-path", AssetsPath)
	c.Set("originals-path", OriginalsPath)
	c.Set("darktable-cli", DarktableCli)

	return c
}

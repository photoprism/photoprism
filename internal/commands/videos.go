package commands

import "github.com/urfave/cli/v2"

// VideosCommands configures the CLI subcommands for working with indexed videos.
var VideosCommands = &cli.Command{
	Name:    "videos",
	Aliases: []string{"video"},
	Usage:   "Video troubleshooting and editing subcommands",
	Subcommands: []*cli.Command{
		VideoListCommand,
		VideoTrimCommand,
		VideoRemuxCommand,
		VideoTranscodeCommand,
		VideoInfoCommand,
	},
}

// videoCountFlag limits the number of results returned by video commands.
var videoCountFlag = &cli.IntFlag{
	Name:    "count",
	Aliases: []string{"n"},
	Usage:   "maximum `NUMBER` of results",
	Value:   10000,
}

// videoForceFlag allows overwriting existing output files for remux/transcode.
var videoForceFlag = &cli.BoolFlag{
	Name:    "force",
	Aliases: []string{"f"},
	Usage:   "replace existing output files",
}

// videoVerboseFlag adds raw metadata to video info output.
var videoVerboseFlag = &cli.BoolFlag{
	Name:  "verbose",
	Usage: "include raw metadata output",
}

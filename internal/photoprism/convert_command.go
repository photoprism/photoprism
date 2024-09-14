package photoprism

import (
	"os/exec"

	"github.com/photoprism/photoprism/pkg/media"
)

// ConvertCommand represents a command to be executed for converting a MediaFile.
// including any options to be used for this.
type ConvertCommand struct {
	Cmd         *exec.Cmd
	Orientation media.Orientation
}

// String returns the conversion command as string e.g. for logging.
func (c *ConvertCommand) String() string {
	if c.Cmd == nil {
		return ""
	}

	return c.Cmd.String()
}

// WithOrientation sets the media Orientation after successful conversion.
func (c *ConvertCommand) WithOrientation(o media.Orientation) *ConvertCommand {
	c.Orientation = media.ParseOrientation(o, c.Orientation)
	return c
}

// ResetOrientation resets the media Orientation after successful conversion.
func (c *ConvertCommand) ResetOrientation() *ConvertCommand {
	return c.WithOrientation(media.ResetOrientation)
}

// NewConvertCommand returns a new file converter command with default options.
func NewConvertCommand(cmd *exec.Cmd) *ConvertCommand {
	if cmd == nil {
		return nil
	}

	return &ConvertCommand{
		Cmd:         cmd,                   // File conversion command.
		Orientation: media.KeepOrientation, // Keep the orientation by default.
	}
}

// ConvertCommands represents a list of possible ConvertCommand commands for converting a MediaFile,
// sorted by priority.
type ConvertCommands []*ConvertCommand

// NewConvertCommands returns a new, empty list of ConvertCommand commands.
func NewConvertCommands() ConvertCommands {
	return make(ConvertCommands, 0, 8)
}

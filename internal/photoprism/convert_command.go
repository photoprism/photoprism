package photoprism

import (
	"os/exec"

	"github.com/photoprism/photoprism/pkg/media"
)

// ConvertCmd represents a command to be executed for converting a MediaFile.
// including any options to be used for this.
type ConvertCmd struct {
	Cmd         *exec.Cmd
	Orientation media.Orientation
}

// String returns the conversion command as string e.g. for logging.
func (c *ConvertCmd) String() string {
	if c.Cmd == nil {
		return ""
	}

	return c.Cmd.String()
}

// WithOrientation sets the media Orientation after successful conversion.
func (c *ConvertCmd) WithOrientation(o media.Orientation) *ConvertCmd {
	c.Orientation = media.ParseOrientation(o, c.Orientation)
	return c
}

// ResetOrientation resets the media Orientation after successful conversion.
func (c *ConvertCmd) ResetOrientation() *ConvertCmd {
	return c.WithOrientation(media.ResetOrientation)
}

// NewConvertCmd returns a new file converter command with default options.
func NewConvertCmd(cmd *exec.Cmd) *ConvertCmd {
	if cmd == nil {
		return nil
	}

	return &ConvertCmd{
		Cmd:         cmd,                   // File conversion command.
		Orientation: media.KeepOrientation, // Keep the orientation by default.
	}
}

// ConvertCmds represents a list of possible ConvertCommand commands for converting a MediaFile, sorted by priority.
type ConvertCmds []*ConvertCmd

// NewConvertCmds returns a new, empty list of ConvertCommand commands.
func NewConvertCmds() ConvertCmds {
	return make(ConvertCmds, 0, 8)
}

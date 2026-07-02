package photoprism

import (
	"os/exec"
	"strings"

	"github.com/photoprism/photoprism/pkg/media"
)

// ConvertCmd represents a command to be executed for converting a MediaFile.
// including any options to be used for this.
type ConvertCmd struct {
	Cmd          *exec.Cmd
	Orientation  media.Orientation
	VerifyImage  bool
	RejectStderr []string
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

// WithImageVerification marks the command's output for a decode check before acceptance,
// so corrupt embedded previews are rejected and the loop tries the next command.
func (c *ConvertCmd) WithImageVerification() *ConvertCmd {
	c.VerifyImage = true
	return c
}

// WithStderrRejection rejects the command's output when its stderr contains any of the given
// substrings, even on a zero exit code, so the loop tries the next converter.
func (c *ConvertCmd) WithStderrRejection(patterns ...string) *ConvertCmd {
	c.RejectStderr = append(c.RejectStderr, patterns...)
	return c
}

// StderrRejected reports whether the given stderr output matches a rejection pattern.
func (c *ConvertCmd) StderrRejected(stderr string) bool {
	if stderr == "" {
		return false
	}

	for _, pattern := range c.RejectStderr {
		if pattern != "" && strings.Contains(stderr, pattern) {
			return true
		}
	}

	return false
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

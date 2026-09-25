package photoprism

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/photoprism/photoprism/internal/ffmpeg/encode"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/proc"
)

// ConvertBudget is the time available for converting one file, shared by the commands tried
// for it. Each run is charged against the remainder, so a chain of candidates cannot occupy
// the converter for the configured time once per candidate.
type ConvertBudget struct {
	remaining time.Duration
	limited   bool
}

// NewConvertBudget returns a budget for the given total, which may be 0 for no limit.
func NewConvertBudget(total time.Duration) *ConvertBudget {
	return &ConvertBudget{remaining: total, limited: total > 0}
}

// Remaining returns the time left in the budget, or 0 when it is not limited.
func (b *ConvertBudget) Remaining() time.Duration {
	if b == nil || !b.limited {
		return 0
	}

	if b.remaining < 0 {
		return 0
	}

	return b.remaining
}

// Exhausted reports whether a limited budget has no time left.
func (b *ConvertBudget) Exhausted() bool {
	return b != nil && b.limited && b.remaining <= 0
}

// Run executes the given command within the remaining budget and charges what it used.
//
// The stderr accessor is called only when the command fails, so that a tool's own diagnostic
// replaces a bare exit status without discarding a deadline, which the caller must still be
// able to recognize.
func (b *ConvertBudget) Run(cmd *exec.Cmd, stderr func() string) error {
	if b == nil {
		return proc.Run(cmd, 0)
	}

	if b.Exhausted() {
		return fmt.Errorf("%w before %s ran", proc.ErrTimeout, filepath.Base(cmd.Path))
	}

	start := time.Now()
	err := proc.Run(cmd, b.Remaining())

	if b.limited {
		b.remaining -= time.Since(start)
	}

	if err == nil {
		return nil
	}

	// Wrapped rather than replaced: a stopped command has usually written to stderr too, and
	// substituting that text would hide why it stopped.
	if stderr != nil {
		if s := strings.TrimSpace(stderr()); s != "" {
			return fmt.Errorf("%w: %s", err, s)
		}
	}

	return err
}

// LogConvertError reports a failed conversion command, naming the deadline when one was hit so
// that a stopped command is distinguishable from input the tool could not read.
func LogConvertError(err error, cmd *exec.Cmd, logName string) {
	if err == nil {
		return
	}

	name := "converter"

	if cmd != nil {
		name = filepath.Base(cmd.Path)
	}

	if errors.Is(err, proc.ErrTimeout) {
		log.Warnf("convert: %s did not finish %s within the time allowed", name, logName)
		return
	}

	log.Debugf("convert: %s (%s)", clean.Error(err), name)
}

// RemoveConvertOutput deletes a conversion result that was not accepted, so that an incomplete
// file is not mistaken for a finished rendition by the next command or the next indexing pass.
func RemoveConvertOutput(fileName string, cmd *exec.Cmd) {
	if fileName == "" || !fs.FileExists(fileName) {
		return
	}

	if err := os.Remove(fileName); err != nil && !os.IsNotExist(err) {
		name := "converter"

		if cmd != nil {
			name = filepath.Base(cmd.Path)
		}

		log.Tracef("convert: %s (%s)", clean.Error(err), name)
	}
}

// RemuxOptions returns FFmpeg remux options carrying the configured transcoding budget, so a
// container rewrite is bounded the same way a transcode is.
func (w *Convert) RemuxOptions(container fs.Type, force bool) encode.Options {
	opt := encode.NewRemuxOptions(w.conf.FFmpegBin(), container, force)
	opt.Timeout = w.conf.TranscodeTimeout()

	return opt
}

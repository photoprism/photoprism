package photoprism

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/proc"
)

// exifToolConditionFailed is the ExifTool exit status when every file failed the -if condition.
const exifToolConditionFailed = 2

// exifToolTmpSuffix is appended to a file name for the temporary file ExifTool writes in its place.
const exifToolTmpSuffix = "_exiftool_tmp"

// writeMissingOrientation writes the EXIF orientation to an image that has no Orientation tag yet,
// and reports whether it did. An existing tag is kept, and values outside 2..8 are never written.
func (w *Convert) writeMissingOrientation(fileName string, orientation int) (written bool, err error) {
	if orientation < 2 || orientation > 8 {
		return false, nil
	}

	// #nosec G204 -- arguments are the configured ExifTool binary, a bounded number, and a file path.
	cmd := exec.Command(w.conf.ExifToolBin(), "-overwrite_original", "-n", "-if", "not defined $Orientation", "-Orientation="+strconv.Itoa(orientation), fileName)

	var stderr bytes.Buffer
	cmd.Stdout = &bytes.Buffer{}
	cmd.Stderr = &stderr
	cmd.Env = append(cmd.Env, fmt.Sprintf("HOME=%s", w.conf.CmdCachePath()))

	log.Trace(clean.Cmd(cmd))

	if err = proc.Run(cmd, w.conf.ConvertTimeout()); err == nil {
		return true, nil
	}

	if exitErr := (*exec.ExitError)(nil); errors.As(err, &exitErr) && exitErr.ExitCode() == exifToolConditionFailed {
		return false, nil
	} else if s := strings.TrimSpace(stderr.String()); s != "" {
		// The error is logged at warning level, so it names the file without its directory.
		return false, fmt.Errorf("%w: %s", err, strings.ReplaceAll(s, fileName, filepath.Base(fileName)))
	}

	return false, err
}

package photoprism

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/proc"
)

// UploadPreviewTimeout is the longest time upload screening spends converting one file to a
// temporary preview; conversion is stopped when it is reached.
const UploadPreviewTimeout = time.Minute

// TempPreview creates a request-scoped JPEG for media that the native image decoder cannot read.
func (w *Convert) TempPreview(f *MediaFile) (fileName string, cleanup func(), err error) {
	if w == nil || w.conf == nil {
		return "", nil, fmt.Errorf("convert: no configuration provided")
	}

	return w.tempPreview(f, NewConvertBudget(w.previewBudget()))
}

// uploadPreviewLimit is the screening bound applied by previewBudget, overridden only in tests.
var uploadPreviewLimit = UploadPreviewTimeout

// previewBudget returns the time available for creating a temporary upload preview.
func (w *Convert) previewBudget() time.Duration {
	return uploadPreviewBudget(w.conf.ConvertTimeout(), uploadPreviewLimit)
}

// uploadPreviewBudget returns the smaller of the conversion timeout and the limit, or the limit
// when the conversion timeout is disabled, so that screening converters always have a deadline.
func uploadPreviewBudget(convertTimeout, limit time.Duration) time.Duration {
	if convertTimeout > 0 && convertTimeout < limit {
		return convertTimeout
	}

	return limit
}

// tempPreview creates a temporary JPEG preview, running all converters within the given budget,
// or within a new preview budget if it is nil.
func (w *Convert) tempPreview(f *MediaFile, budget *ConvertBudget) (fileName string, cleanup func(), err error) {
	if w == nil || w.conf == nil {
		return "", nil, fmt.Errorf("convert: no configuration provided")
	}

	if budget == nil {
		budget = NewConvertBudget(w.previewBudget())
	}

	if f == nil || !f.Exists() || f.Empty() {
		return "", nil, fmt.Errorf("convert: invalid media file")
	}
	if _, _, decodeErr := fs.DecodeImageFile(f.FileName()); decodeErr == nil {
		return f.FileName(), func() {}, nil
	}

	logName := clean.Log(filepath.Base(f.FileName()))

	tempDir, err := os.MkdirTemp(w.conf.TempPath(), "nsfw-preview-")
	if err != nil {
		return "", nil, fmt.Errorf("convert: failed to create preview directory for %s", logName)
	}

	// The directory is removed on every way out, including a panic, unless the caller owns it.
	owned := false
	removeDir := func() {
		if removeErr := os.RemoveAll(tempDir); removeErr != nil {
			log.Debugf("convert: failed to remove temporary preview of %s", logName)
		}
	}
	defer func() {
		if !owned {
			removeDir()
		}
	}()

	previewName := filepath.Join(tempDir, "preview.jpg")
	cmds, useMutex, err := w.JpegConvertCmds(f, previewName, "")
	if err != nil {
		return "", nil, err
	}
	if useMutex {
		w.cmdMutex.Lock()
		defer w.cmdMutex.Unlock()
	}

	if previewFromCmds(cmds, previewName, logName, budget, w.cmdEnv()) {
		owned = true
		return previewName, removeDir, nil
	}

	if budget.Exhausted() {
		return "", nil, fmt.Errorf("convert: %w creating a preview of %s", proc.ErrTimeout, logName)
	}

	return "", nil, fmt.Errorf("convert: no usable preview for %s", logName)
}

// cmdEnv returns the environment variables for running conversion commands.
func (w *Convert) cmdEnv() []string {
	return []string{
		fmt.Sprintf("HOME=%s", w.conf.CmdCachePath()),
		fmt.Sprintf("LD_LIBRARY_PATH=%s", w.conf.CmdLibPath()),
	}
}

// previewFromCmds runs the candidates within the budget until one writes a decodable JPEG to
// previewName, either directly or to stdout, and reports whether one did. A nil env gives the
// commands an empty environment rather than the process environment.
func previewFromCmds(cmds ConvertCmds, previewName, logName string, budget *ConvertBudget, env []string) bool {
	for _, candidate := range cmds {
		if budget.Exhausted() {
			return false
		}

		// Each candidate is judged on its own output, never on a file a previous one left behind.
		_ = os.Remove(previewName)
		var stdout, stderr bytes.Buffer
		candidate.Cmd.Stdout = &stdout
		candidate.Cmd.Stderr = &stderr
		candidate.Cmd.Env = append([]string{}, env...)
		cmdName := filepath.Base(candidate.Cmd.Path)
		if runErr := budget.Run(candidate.Cmd, nil); runErr != nil {
			if errors.Is(runErr, proc.ErrTimeout) {
				log.Warnf("convert: %s did not finish a preview of %s within the time allowed", cmdName, logName)
			} else {
				log.Debugf("convert: %s failed to create a preview of %s", cmdName, logName)
			}
			RemoveConvertOutput(previewName, candidate.Cmd)
			continue
		}
		if candidate.StderrRejected(stderr.String()) {
			continue
		}
		if !fs.FileExistsNotEmpty(previewName) {
			data := stdout.Bytes()
			if _, format, decodeErr := fs.DecodeImageData(data); decodeErr != nil || format != "jpeg" {
				continue
			}
			if writeErr := os.WriteFile(previewName, data, fs.ModeFile); writeErr != nil {
				continue
			}
		}
		if _, _, decodeErr := fs.DecodeImageFile(previewName); decodeErr == nil {
			return true
		}
	}

	return false
}

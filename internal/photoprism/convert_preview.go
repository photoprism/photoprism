package photoprism

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
)

// TempPreview creates a request-scoped JPEG for media that the native image decoder cannot read.
func (w *Convert) TempPreview(f *MediaFile) (fileName string, cleanup func(), err error) {
	if w == nil || w.conf == nil {
		return "", nil, fmt.Errorf("convert: no configuration provided")
	}
	if f == nil || !f.Exists() || f.Empty() {
		return "", nil, fmt.Errorf("convert: invalid media file")
	}
	if _, _, decodeErr := fs.DecodeImageFile(f.FileName()); decodeErr == nil {
		return f.FileName(), func() {}, nil
	}

	tempDir, err := os.MkdirTemp(w.conf.TempPath(), "nsfw-preview-")
	if err != nil {
		return "", nil, fmt.Errorf("convert: create preview directory: %w", err)
	}
	cleanup = func() {
		if removeErr := os.RemoveAll(tempDir); removeErr != nil {
			log.Debugf("convert: %s (remove temporary preview)", clean.Error(removeErr))
		}
	}
	previewName := filepath.Join(tempDir, "preview.jpg")
	cmds, useMutex, err := w.JpegConvertCmds(f, previewName, "")
	if err != nil {
		cleanup()
		return "", nil, err
	}
	if useMutex {
		w.cmdMutex.Lock()
		defer w.cmdMutex.Unlock()
	}

	for _, candidate := range cmds {
		_ = os.Remove(previewName)
		var stdout, stderr bytes.Buffer
		candidate.Cmd.Stdout = &stdout
		candidate.Cmd.Stderr = &stderr
		candidate.Cmd.Env = append(os.Environ(),
			fmt.Sprintf("HOME=%s", w.conf.CmdCachePath()),
			fmt.Sprintf("LD_LIBRARY_PATH=%s", w.conf.CmdLibPath()))
		if runErr := candidate.Cmd.Run(); runErr != nil {
			if message := strings.TrimSpace(stderr.String()); message != "" {
				runErr = errors.New(message)
			}
			log.Debugf("convert: %s (%s preview)", clean.Error(runErr), filepath.Base(candidate.Cmd.Path))
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
			return previewName, cleanup, nil
		}
	}

	cleanup()
	return "", nil, fmt.Errorf("convert: no usable preview for %s", clean.Log(f.RootRelName()))
}

package photoprism

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"unicode"

	"github.com/photoprism/photoprism/internal/meta"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/proc"
)

// ToJson uses exiftool to export metadata to a json file, or returns the name of a valid cached export.
func (w *Convert) ToJson(f *MediaFile, force bool) (jsonName string, err error) {
	if f == nil {
		return "", fmt.Errorf("exiftool: no media file provided for processing - you may have found a bug")
	}

	jsonName, err = f.ExifToolJsonName()

	if err != nil {
		return "", nil
	}

	if exifToolCacheValid(jsonName) {
		return jsonName, nil
	}

	log.Debugf("exiftool: extracting metadata from %s", clean.Log(f.RootRelName()))

	// ExifTool command arguments.
	var args []string

	// Use the "-ee" flag to extract embedded metadata from MPEG-2 Transport Stream and AVCHD video files,
	// see https://exiftool.org/exiftool_pod.html#ee-NUM--extractEmbedded for details.
	if f.IsVideo() {
		args = []string{"-n", "-ee", "-m", "-api", "LargeFileSupport", "-j", f.FileName()}
	} else {
		args = []string{"-n", "-m", "-api", "LargeFileSupport", "-j", f.FileName()}
	}

	// Create ExifTool command with arguments.
	// #nosec G204 -- arguments are built from validated config and file paths.
	cmd := exec.Command(w.conf.ExifToolBin(), args...)

	// Command environment, output and errors.
	out := jsonOutputBuffer{limit: int(meta.JSONFileLimit()), failOnLimit: true}
	stderr := jsonOutputBuffer{limit: 64 << 10}
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	cmd.Env = append(cmd.Env, []string{
		fmt.Sprintf("HOME=%s", w.conf.CmdCachePath()),
	}...)

	// Log exact command for debugging in trace mode.
	log.Trace(clean.Cmd(cmd))

	// Run convert command.
	err = proc.Run(cmd, w.conf.ConvertTimeout())

	if out.exceeded {
		err = meta.ErrJSONFileTooLarge
	}

	if err != nil {
		if s := string(stderr.data); s != "" {
			if stderr.exceeded {
				s += " [truncated]"
			}
			err = fmt.Errorf("%w: %s", err, s)
		}

		LogConvertError(err, cmd, clean.Log(f.RootRelName()))

		return "", err
	}

	if !jsonArrayBounds(out.data, out.data) {
		return "", fmt.Errorf("exiftool: unexpected output for %s", clean.Log(f.RootRelName()))
	} else if err = writeExifToolCache(jsonName, out.data); err != nil {
		return "", err
	}

	log.Debugf("cache: created %s", filepath.Base(jsonName))

	return jsonName, err
}

// exifToolCacheValid reports whether a cached ExifTool export can be read and holds a JSON array, as
// written with "-j". It only checks the first and last non-space characters within 64 bytes of each end,
// so an empty or truncated file is replaced, without parsing the whole file each time it is checked.
func exifToolCacheValid(fileName string) bool {
	// Opening a named pipe would block, so the type is checked first.
	if info, err := os.Stat(fileName); err != nil || !info.Mode().IsRegular() {
		return false
	}

	f, err := os.Open(fileName) //nolint:gosec // the name is derived from a file hash in the cache folder.

	if err != nil {
		return false
	}

	defer f.Close()

	info, err := f.Stat()

	if err != nil || !info.Mode().IsRegular() || info.Size() < 2 {
		return false
	}

	head := make([]byte, min(info.Size(), 64))
	tail := make([]byte, len(head))

	if _, err = f.ReadAt(head, 0); err != nil {
		return false
	} else if _, err = f.ReadAt(tail, info.Size()-int64(len(tail))); err != nil {
		return false
	}

	return jsonArrayBounds(head, tail)
}

// jsonArrayBounds reports whether data that starts with head and ends with tail is enclosed in square
// brackets, ignoring surrounding white space.
func jsonArrayBounds(head, tail []byte) bool {
	head, tail = bytes.TrimLeftFunc(head, unicode.IsSpace), bytes.TrimRightFunc(tail, unicode.IsSpace)

	return len(head) > 0 && head[0] == '[' && len(tail) > 0 && tail[len(tail)-1] == ']'
}

// writeExifToolCache writes an ExifTool export through a staged sibling and publishes it by rename,
// replacing an existing cache file, so readers never see a partially written file.
func writeExifToolCache(fileName string, data []byte) (err error) {
	if err = fs.MkdirAll(filepath.Dir(fileName)); err != nil {
		return err
	}

	f, err := fs.OpenStageFile(fileName)

	// The cache folder may have been removed concurrently, e.g. by a cleanup, so it is created again once.
	if os.IsNotExist(err) {
		if err = fs.MkdirAll(filepath.Dir(fileName)); err != nil {
			return err
		}

		f, err = fs.OpenStageFile(fileName)
	}

	if err != nil {
		return err
	}

	staged, published := f.Name(), false

	defer func() {
		if !published {
			_ = f.Close()
			_ = os.Remove(staged)
		}
	}()

	if _, err = f.Write(data); err != nil {
		return err
	} else if err = f.Close(); err != nil {
		return err
	} else if err = fs.PublishFile(staged, fileName, true); err != nil {
		return err
	}

	published = true

	return nil
}

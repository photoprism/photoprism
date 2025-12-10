package photoprism

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
)

// ToJson uses exiftool to export metadata to a json file.
func (w *Convert) ToJson(f *MediaFile, force bool) (jsonName string, err error) {
	if f == nil {
		return "", fmt.Errorf("exiftool: no media file provided for processing - you may have found a bug")
	}

	jsonName, err = f.ExifToolJsonName()

	if err != nil {
		return "", nil
	}

	if fs.FileExists(jsonName) {
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
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	cmd.Env = append(cmd.Env, []string{
		fmt.Sprintf("HOME=%s", w.conf.CmdCachePath()),
	}...)

	// Log exact command for debugging in trace mode.
	log.Trace(cmd.String())

	// Run convert command.
	if err = cmd.Run(); err != nil {
		if stderr.String() != "" {
			return "", errors.New(stderr.String())
		} else {
			return "", err
		}
	}

	// Write output to file (make parent dir robustly in case a parallel test cleaned the cache).
	if err = os.WriteFile(jsonName, out.Bytes(), fs.ModeFile); err != nil {
		// If the parent directory vanished due to concurrent cleanup, recreate and retry once.
		if !os.IsNotExist(err) {
			return "", err
		} else if err = fs.MkdirAll(filepath.Dir(jsonName)); err != nil {
			return "", err
		} else if err = os.WriteFile(jsonName, out.Bytes(), fs.ModeFile); err != nil {
			return "", err
		}
	}

	// Check if file exists.
	if fs.FileExists(jsonName) {
		log.Debugf("cache: created %s", filepath.Base(jsonName))
	} else {
		return "", fmt.Errorf("exiftool: failed to create %s", filepath.Base(jsonName))
	}

	return jsonName, err
}

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
func (c *Convert) ToJson(f *MediaFile, force bool) (jsonName string, err error) {
	if f == nil {
		return "", fmt.Errorf("exiftool: file is nil - you may have found a bug")
	}

	jsonName, err = f.ExifToolJsonName()

	if err != nil {
		return "", nil
	}

	if fs.FileExists(jsonName) {
		return jsonName, nil
	}

	log.Debugf("exiftool: extracting metadata from %s", clean.Log(f.RootRelName()))

	cmd := exec.Command(c.conf.ExifToolBin(), "-n", "-m", "-api", "LargeFileSupport", "-j", f.FileName())

	// Fetch command output.
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	cmd.Env = []string{fmt.Sprintf("HOME=%s", c.conf.CmdCachePath())}

	// Log exact command for debugging in trace mode.
	log.Trace(cmd.String())

	// Run convert command.
	if err := cmd.Run(); err != nil {
		if stderr.String() != "" {
			return "", errors.New(stderr.String())
		} else {
			return "", err
		}
	}

	// Write output to file.
	if err := os.WriteFile(jsonName, []byte(out.String()), fs.ModeFile); err != nil {
		return "", err
	}

	// Check if file exists.
	if fs.FileExists(jsonName) {
		log.Debugf("cache: created %s", filepath.Base(jsonName))
	} else {
		return "", fmt.Errorf("exiftool: failed to create %s", filepath.Base(jsonName))
	}

	return jsonName, err
}

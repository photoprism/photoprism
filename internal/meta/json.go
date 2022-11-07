package meta

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"

	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
)

// JSON parses a json sidecar file (as used by Exiftool) and returns a Data struct.
func JSON(jsonName, originalName string) (data Data, err error) {
	err = data.JSON(jsonName, originalName)

	return data, err
}

// JSON parses a json sidecar file (as used by Exiftool) and returns a Data struct.
func (data *Data) JSON(jsonName, originalName string) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("metadata: %s in %s (json panic)\nstack: %s", e, clean.Log(filepath.Base(jsonName)), debug.Stack())
		}
	}()

	quotedName := clean.Log(filepath.Base(jsonName))

	// Resolve JSON file name e.g. in case it's a symlink.
	if jsonName, err = fs.Resolve(jsonName); err != nil {
		return fmt.Errorf("metadata: %s not found (%s)", quotedName, err)
	} else if !fs.FileExists(jsonName) {
		return fmt.Errorf("metadata: %s not found", quotedName)
	}

	jsonData, err := os.ReadFile(jsonName)

	if err != nil {
		return fmt.Errorf("cannot read json file %s", quotedName)
	}

	if bytes.Contains(jsonData, []byte("ExifToolVersion")) {
		return data.Exiftool(jsonData, originalName)
	} else if bytes.Contains(jsonData, []byte("albumData")) {
		return data.GMeta(jsonData)
	} else if bytes.Contains(jsonData, []byte("photoTakenTime")) {
		return data.GPhoto(jsonData)
	}

	log.Warnf("metadata: unknown json in %s", quotedName)

	return fmt.Errorf("unknown json in %s", quotedName)
}

package photoprism

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/h2non/filetype"
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
	if !fs.FileExists(jsonName) {
		return "", fmt.Errorf(
			"exiftool: failed creating %s", filepath.Base(jsonName),
		)
	}

	// attempt to extract the video file if it exists
	var payload []interface{}
	err = json.Unmarshal([]byte(out.String()), &payload)
	if err != nil {
		log.Debugf("Error parsing exiftool JSON payload: ", err)
		return "", err
	}
	if len(payload) > 0 {
		if payloadMap, ok := payload[0].(map[string]interface{}); ok {
			var extractCmd *exec.Cmd
			path := filepath.Join(
				Config().SidecarPath(), f.RootRelPath(), "%f",
			)
			if _, exists := payloadMap["EmbeddedVideoFile"]; exists {
				extractCmd = exec.Command(
					c.conf.ExifToolBin(), "-EmbeddedVideoFile", "-b", "-w",
					path, f.FileName(),
				)
			} else if _, exists := payloadMap["MotionPhotoVideo"]; exists {
				extractCmd = exec.Command(
					c.conf.ExifToolBin(), "-MotionPhotoVideo", "-b", "-w",
					path, f.FileName(),
				)
			}

			if extractCmd != nil {
				extractCmd.Stdout = &out
				extractCmd.Stderr = &stderr
				if err := extractCmd.Run(); err != nil {
					log.Debugf("Error running exiftool on video file: ", err)
					if stderr.String() != "" {
						return "", errors.New(stderr.String())
					} else {
						return "", err
					}
				}

				// find the extracted file
				savedPath := filepath.Join(
					Config().SidecarPath(),
					f.RootRelPath(),
					f.BasePrefix(false),
				)

				// find the video file type
				buf, err := ioutil.ReadFile(savedPath)
				if err != nil {
					log.Debugf("Error reading sidecar file at %s", savedPath)
					return "", err
				}
				kind, err := filetype.Match(buf)
				if err != nil {
					log.Debugf(
						"Error finding the type of sidecar file at %s",
						savedPath,
					)
					return "", err
				}

				// rename the file with the correct extension
				dir, file := filepath.Split(savedPath)
				newFileName := fmt.Sprintf("%s.%s", file, kind.Extension)
				newSavedPath := filepath.Join(dir, newFileName)
				err = os.Rename(savedPath, newSavedPath)
				if err != nil {
					log.Debugf("Error renaming the file at %s", savedPath)
					return "", err
				}
			}
		}
	}

	return jsonName, err
}

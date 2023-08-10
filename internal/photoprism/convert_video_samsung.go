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
)

func (c *Convert) ToSamsungVideo(
	f *MediaFile,
	jsonName string, force bool,
) (
	*MediaFile,
	error,
) {
	if f == nil {
		return nil, fmt.Errorf("convert: file is nil - you may have found a bug")
	}

	// Abort if the source media file does not exist.
	if !f.Exists() {
		return nil, fmt.Errorf(
			"convert: %s not found", clean.Log(f.RootRelName()),
		)
	} else if f.Empty() {
		return nil, fmt.Errorf(
			"convert: %s is empty", clean.Log(f.RootRelName()),
		)
	}

	// given the filename, read the json sidecar
	jsonBytes, err := os.ReadFile(jsonName)
	if err != nil {
		log.Debugf("Error reading the JSON file: ", err)
		return nil, err
	}
	var jsonPayload []interface{}
	err = json.Unmarshal(jsonBytes, &jsonPayload)
	if err != nil {
		log.Debugf("Error parsing exiftool JSON payload: ", err)
		return nil, err
	}

	if len(jsonPayload) > 0 {
		if payloadMap, ok := jsonPayload[0].(map[string]interface{}); ok {
			var cmd *exec.Cmd
			path := filepath.Join(
				Config().SidecarPath(), f.RootRelPath(), "%f",
			)
			if _, exists := payloadMap["EmbeddedVideoFile"]; exists {
				cmd = exec.Command(
					c.conf.ExifToolBin(), "-EmbeddedVideoFile", "-b", "-w",
					path, f.FileName(),
				)
			} else if _, exists := payloadMap["MotionPhotoVideo"]; exists {
				cmd = exec.Command(
					c.conf.ExifToolBin(), "-MotionPhotoVideo", "-b", "-w",
					path, f.FileName(),
				)
			}

			var out bytes.Buffer
			var stderr bytes.Buffer
			cmd.Stdout = &out
			cmd.Stderr = &stderr
			cmd.Env = []string{
				fmt.Sprintf("HOME=%s", c.conf.CmdCachePath()),
			}

			if cmd != nil {
				if err := cmd.Run(); err != nil {
					log.Debugf("Error running exiftool on video file: ", err)
					if stderr.String() != "" {
						return nil, errors.New(stderr.String())
					} else {
						return nil, err
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
					return nil, err
				}
				kind, err := filetype.Match(buf)
				if err != nil {
					log.Debugf(
						"Error finding the type of sidecar file at %s",
						savedPath,
					)
					return nil, err
				}

				// rename the file with the correct extension
				dir, file := filepath.Split(savedPath)
				newFileName := fmt.Sprintf("%s.%s", file, kind.Extension)
				newSavedPath := filepath.Join(dir, newFileName)
				err = os.Rename(savedPath, newSavedPath)
				if err != nil {
					log.Debugf("Error renaming the file at %s", savedPath)
					return nil, err
				}

				return NewMediaFile(newSavedPath)
			}
		}
	}

	return nil, nil

}

package photoprism

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
)

func execExifTool(conf *config.Config, filename string, arguments ...string) (output bytes.Buffer, err error) {
	cliArgs := append(arguments, filename)
	cmd := exec.Command(conf.ExifToolBin(), cliArgs...)

	// Fetch command output.
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	cmd.Env = []string{fmt.Sprintf("HOME=%s", conf.CmdCachePath())}

	// Log exact command for debugging in trace mode.
	log.Trace(cmd.String())

	// Run convert command.
	if err := cmd.Run(); err != nil {
		if stderr.String() != "" {
			return out, errors.New(stderr.String())
		} else {
			return out, err
		}
	}

	return out, err
}

func injectMotionVideoOffset(conf *config.Config, f *MediaFile, exifJsonOutput string) (output string, err error) {
	rawOutput, err := execExifTool(conf, f.FileName(), "-a", "-DirectoryItemLength", "-m", "-n")
	if err != nil {
		return exifJsonOutput, err
	}
	scanner := bufio.NewScanner(&rawOutput)

	maxLength := 0
	iterations := 0
	for scanner.Scan() {
		line := scanner.Text()
		pos := strings.Index(line, ":")
		if pos > 0 {
			itemLength, err := strconv.Atoi(strings.TrimSpace(line[pos+1 : len(line)]))

			if err == nil {
				if itemLength > maxLength {
					maxLength = itemLength
				}
				iterations++
			}
		}
	}
	if maxLength < 1 || iterations < 2 {
		return exifJsonOutput, errors.New("exiftool: did not find valid DirectoryItemLength data for video offset for google motion photo")
	}

	var mapSlice []map[string]interface{}
	if err := json.Unmarshal([]byte(exifJsonOutput), &mapSlice); err != nil {
		return exifJsonOutput, err
	}

	mapSlice[0]["EmbeddedVideoOffset"] = f.FileSize() - int64(maxLength)
	json, err := json.MarshalIndent(mapSlice, "", " ")
	return string(json), err
}

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

	out, err := execExifTool(c.conf, f.FileName(), "-n", "-m", "-api", "LargeFileSupport", "-j")
	if err != nil {
		return "", err
	}

	outputString := out.String()

	if f.IsGoogleMotionPhoto() {
		injectedJson, err := injectMotionVideoOffset(c.conf, f, outputString)
		if err != nil {
			log.Infof("exiftool: Failed to extract video offset for %s ignoring. %s", clean.Log(f.RootRelName()), err.Error())
		} else {
			outputString = injectedJson
		}
	}

	// Write output to file.
	if err := os.WriteFile(jsonName, []byte(outputString), fs.ModeFile); err != nil {
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

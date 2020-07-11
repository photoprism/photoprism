package meta

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/photoprism/photoprism/pkg/txt"
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
			err = fmt.Errorf("metadata: %s (json panic)", e)
		}
	}()

	if data.All == nil {
		data.All = make(map[string]string)
	}

	quotedName := txt.Quote(filepath.Base(jsonName))
	jsonData, err := ioutil.ReadFile(jsonName)

	if err != nil {
		log.Warnf("metadata: %s (json)", err.Error())
		return fmt.Errorf("can't read json file %s", quotedName)
	}

	if bytes.Contains(jsonData, []byte("ExifToolVersion")) {
		return data.Exiftool(jsonData, originalName)
	} else if bytes.Contains(jsonData, []byte("geoData")) {
		return data.GPhotos(jsonData)
	}

	log.Warnf("metadata: unknown format in %s (json)", quotedName)
	return fmt.Errorf("unknown json format in %s", quotedName)
}

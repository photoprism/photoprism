package face

import (
	"errors"
	"path/filepath"

	pigo "github.com/esimov/pigo/core"
)

// FlpCascade holds the binary representation of the facial landmark points cascade files
type FlpCascade struct {
	*pigo.PuplocCascade
	error
}

// ReadCascadeDir reads the facial landmark points cascade files from the provided directory.
func ReadCascadeDir(plc *pigo.PuplocCascade, path string) (result map[string][]*FlpCascade, err error) {
	result = make(map[string][]*FlpCascade)
	cascades, err := efs.ReadDir(path)

	if len(cascades) == 0 {
		return nil, errors.New("the cascade directory is empty")
	}

	if err != nil {
		return nil, err
	}

	for _, cascade := range cascades {
		cf, err := filepath.Abs(path + "/" + cascade.Name())
		if err != nil {
			return nil, err
		}
		flpc, err := plc.UnpackFlp(cf)
		result[cascade.Name()] = append(result[cascade.Name()], &FlpCascade{flpc, err})
	}

	return result, err
}

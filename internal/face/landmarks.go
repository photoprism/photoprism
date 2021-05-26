package face

import (
	"embed"
	"errors"
	"path/filepath"

	pigo "github.com/esimov/pigo/core"
)

//go:embed cascade/lps
var efs embed.FS

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
		return result, errors.New("the cascade directory is empty")
	}

	if err != nil {
		return nil, err
	}

	for _, cascade := range cascades {
		cf := filepath.Join(path, cascade.Name())

		f, err := efs.ReadFile(cf)

		if err != nil {
			return result, err
		}

		flpc, err := plc.UnpackCascade(f)

		if err != nil {
			return result, err
		}

		result[cascade.Name()] = append(result[cascade.Name()], &FlpCascade{flpc, err})
	}

	return result, err
}

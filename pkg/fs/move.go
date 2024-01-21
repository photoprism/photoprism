package fs

import (
	"fmt"
	"os"
	"path/filepath"
)

// Move moves an existing file to a new destination and returns an error if it fails.
func Move(src, dest string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("move: %s (panic)", r)
		}
	}()

	if err = MkdirAll(filepath.Dir(dest)); err != nil {
		return err
	}

	if err = os.Rename(src, dest); err == nil {
		return nil
	}

	if err = Copy(src, dest); err != nil {
		return err
	}

	if err = os.Remove(src); err != nil {
		return err
	}

	return nil
}

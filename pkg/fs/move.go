package fs

import (
	"fmt"
	"os"
	"path/filepath"
)

// Moves a file to a new destination.
func Move(src, dest string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("move: %s (panic)", r)
		}
	}()

	if err := os.MkdirAll(filepath.Dir(dest), os.ModePerm); err != nil {
		return err
	}

	if err := os.Rename(src, dest); err == nil {
		return nil
	}

	if err := Copy(src, dest); err != nil {
		return err
	}

	if err := os.Remove(src); err != nil {
		return err
	}

	return nil
}

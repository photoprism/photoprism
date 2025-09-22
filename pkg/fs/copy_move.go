package fs

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// Copy copies a file to a destination.
func Copy(src, dest string, force bool) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%s (panic)", r)
		}
	}()

	// Check for obviously empty or invalid source and destination file paths.
	if src == "" || src == "." || src == ".." {
		return errors.New("invalid copy source file path")
	} else if dest == "" || dest == "." || dest == ".." {
		return errors.New("invalid copy destination file path")
	}

	// Check whether a destination file and directory name are specified.
	if filepath.Base(dest) == "" {
		return errors.New("invalid copy destination name")
	} else if filepath.Dir(dest) == "" {
		return errors.New("invalid copy destination path")
	}

	// Resolve absolute destination file path and return an error if unsuccessful.
	if dest, err = filepath.Abs(dest); err != nil {
		return err
	}

	destName := filepath.Base(dest)
	destDir := filepath.Dir(dest)

	// Error if source and destination file path are the same.
	if dest == src {
		return fmt.Errorf("cannot copy file %s onto itself", destName)
	}

	// Error if destination exists (and is not empty) without the force flag being used.
	if Exists(dest) {
		if !force && !FileExistsIsEmpty(dest) {
			return fmt.Errorf("copy destination %s already exists", destName)
		}
	}

	// Make sure the target directory exists.
	if err = MkdirAll(destDir); err != nil {
		return err
	}

	thisFile, err := os.Open(src)

	if err != nil {
		return err
	}

	defer thisFile.Close()

	// Open destination for write; create or truncate to avoid trailing bytes
	destFile, err := os.OpenFile(dest, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, ModeFile)

	if err != nil {
		return err
	}

	defer destFile.Close()

	_, err = io.Copy(destFile, thisFile)

	if err != nil {
		return err
	}

	return nil
}

// Move moves an existing file to a new destination and returns an error if it fails.
func Move(src, dest string, force bool) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%s (panic)", r)
		}
	}()

	// Check for obviously empty or invalid source and destination file paths.
	if src == "" || src == "." || src == ".." {
		return errors.New("invalid move source file path")
	} else if dest == "" || dest == "." || dest == ".." {
		return errors.New("invalid move destination file path")
	}

	// Check whether a destination file and directory name are specified.
	if filepath.Base(dest) == "" {
		return errors.New("invalid move destination name")
	} else if filepath.Dir(dest) == "" {
		return errors.New("invalid move destination path")
	}

	// Resolve absolute destination file path and return an error if unsuccessful.
	if dest, err = filepath.Abs(dest); err != nil {
		return err
	}

	destName := filepath.Base(dest)
	destDir := filepath.Dir(dest)

	// Error if source and destination file path are the same.
	if dest == src {
		return fmt.Errorf("cannot move file %s onto itself", destName)
	}

	// Error if destination exists (and is not empty) without the force flag being used.
	if Exists(dest) {
		if !force && !FileExistsIsEmpty(dest) {
			return fmt.Errorf("move destination %s already exists", destName)
		}
	}

	// Make sure the target directory exists.
	if err = MkdirAll(destDir); err != nil {
		return err
	}

	if err = os.Rename(src, dest); err == nil {
		return nil
	}

	if err = Copy(src, dest, true); err != nil {
		return err
	}

	if err = os.Remove(src); err != nil {
		return err
	}

	return nil
}

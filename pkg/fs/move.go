package fs

import (
	"io"
	"os"
	"path/filepath"
)

// Moves a file to a new destination.
func Move(src, dest string) error {
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

// Copies a file to a destination.
func Copy(src, dest string) error {
	if err := os.MkdirAll(filepath.Dir(dest), os.ModePerm); err != nil {
		return err
	}

	thisFile, err := os.Open(src)

	if err != nil {
		return err
	}

	defer thisFile.Close()

	destFile, err := os.OpenFile(dest, os.O_RDWR|os.O_CREATE, os.ModePerm)

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
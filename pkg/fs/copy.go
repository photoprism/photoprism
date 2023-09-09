package fs

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// Copies a file to a destination.
func Copy(src, dest string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("copy: %s (panic)", r)
		}
	}()

	if err := os.MkdirAll(filepath.Dir(dest), ModeDir); err != nil {
		return err
	}

	thisFile, err := os.Open(src)

	if err != nil {
		return err
	}

	defer thisFile.Close()

	destFile, err := os.OpenFile(dest, os.O_RDWR|os.O_CREATE, ModeFile)

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

// Copy by skipping {n} {blockSize}-sized blocks and reading {blockSize}-sized blocks at a time
func CopyWithOffset(src string, dest string, blockSize int, n int) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	destFile, err := os.OpenFile(dest, os.O_RDWR|os.O_CREATE, ModeFile)
	if err != nil {
		return err
	}
	defer destFile.Close()

	// Seek to the position to skip {offset} bytes from the start
	offset := n * blockSize
	_, err = srcFile.Seek(int64(offset), io.SeekStart)
	if err != nil {
		return err
	}

	// Create a buffer to read {n} bytes at a time
	buffer := make([]byte, blockSize)

	// Loop to read {n} bytes at a time and write to the dest file
	for {
		bytesRead, err := srcFile.Read(buffer)
		if err != nil {
			if err == io.EOF {
				break // Reached the end of the file
			}
			return err
		}

		_, err = destFile.Write(buffer[:bytesRead])
		if err != nil {
			return err
		}
	}

	return nil
}

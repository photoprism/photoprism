package fs

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

// WriteFile overwrites a file with the specified bytes as content.
// If the path does not exist or the file cannot be written, an error is returned.
func WriteFile(fileName string, data []byte, perm os.FileMode) error {
	// Return error if no filename was provided.
	if fileName == "" {
		return errors.New("missing filename")
	}

	// Default to regular file permissions.
	if perm == 0 {
		perm = ModeFile
	}

	// Create storage directory if it does not exist yet.
	if dir := filepath.Dir(fileName); dir != "" && dir != "/" && dir != "." && dir != ".." && !PathExists(dir) {
		if err := MkdirAll(dir); err != nil {
			return err
		}
	}

	file, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_TRUNC, perm) //nolint:gosec // caller-controlled path; intended write

	if err != nil {
		return err
	}

	_, err = file.Write(data)

	if closeErr := file.Close(); closeErr != nil && err == nil {
		err = closeErr
	}

	return err
}

// WriteString overwrites a file with the specified string as content.
// If the path does not exist or the file cannot be written, an error is returned.
func WriteString(fileName string, s string) error {
	return WriteFile(fileName, []byte(s), ModeFile)
}

// WriteUnixTime overwrites a file with the current Unix timestamp as content.
// If the path does not exist or the file cannot be written, an error is returned.
func WriteUnixTime(fileName string) (unixTime int64, err error) {
	unixTime = time.Now().Unix()
	return unixTime, WriteString(fileName, strconv.FormatInt(unixTime, 10))
}

// WriteFileFromReader writes data from an io.Reader to a newly created file with the specified name.
// If the path does not exist or the file cannot be written, an error is returned.
func WriteFileFromReader(fileName string, reader io.Reader) (err error) {
	if fileName == "" {
		return errors.New("filename missing")
	} else if reader == nil {
		return errors.New("reader missing")
	}

	var file *os.File

	if file, err = os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_TRUNC, ModeFile); err != nil { //nolint:gosec // caller-controlled path; intended write
		return err
	}

	buf := getCopyBuffer()
	defer putCopyBuffer(buf)

	_, err = io.CopyBuffer(file, reader, buf)

	if closeErr := file.Close(); closeErr != nil && err == nil {
		err = closeErr
	}

	return err
}

// CacheFileFromReader writes data from an io.Reader to a file with the specified name if it does not exist.
// If the path does not exist or the file cannot be written, an error is returned.
// No error is returned if the file already exists.
func CacheFileFromReader(fileName string, reader io.Reader) (string, error) {
	if FileExistsNotEmpty(fileName) {
		return fileName, nil
	}

	if err := WriteFileFromReader(fileName, reader); err != nil {
		return "", err
	}

	return fileName, nil
}

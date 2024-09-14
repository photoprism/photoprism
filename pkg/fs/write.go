package fs

import (
	"errors"
	"io"
	"os"
	"strconv"
	"time"
)

// WriteFile overwrites a file with the specified bytes as content.
// If the path does not exist or the file cannot be written, an error is returned.
func WriteFile(fileName string, data []byte) error {
	file, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_TRUNC, ModeFile)

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
	return WriteFile(fileName, []byte(s))
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

	if file, err = os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_TRUNC, ModeFile); err != nil {
		return err
	}

	_, err = io.Copy(file, reader)

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

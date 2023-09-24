package fs

import (
	"errors"
	"io"
	"os"
)

// WriteFile writes data from a reader to a new file.
func WriteFile(fileName string, reader io.Reader) (err error) {
	if fileName == "" {
		return errors.New("filename missing")
	} else if reader == nil {
		return errors.New("reader missing")
	}

	var f *os.File

	if f, err = os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE, ModeFile); err != nil {
		return err
	}

	defer f.Close()

	if _, err = io.Copy(f, reader); err != nil {
		return err
	}

	return nil
}

// CacheFile writes data from a reader to a new file if it does not already exist,
// and returns the name of the file or an error otherwise.
func CacheFile(fileName string, reader io.Reader) (string, error) {
	if FileExistsNotEmpty(fileName) {
		return fileName, nil
	}

	if err := WriteFile(fileName, reader); err != nil {
		return "", err
	}

	return fileName, nil
}

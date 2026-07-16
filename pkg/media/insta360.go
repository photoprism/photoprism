package media

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
)

const (
	insta360MetadataScanLimit = 64 * 1024
	insta360OneRSModel        = "Insta360 OneRS"
	insta360OneRSImageField   = "\x00Arashi Vision\x00Insta360 OneRS\x00"
)

var insta360OneRSVideoField = append([]byte{0x12, byte(len(insta360OneRSModel))}, []byte(insta360OneRSModel)...)

// Insta360CameraModelFile returns a recognized camera model from an Insta360 original.
func Insta360CameraModelFile(fileName string) (model string, err error) {
	if fileName == "" {
		return "", errors.New("filename missing")
	}

	// #nosec G304 -- the media filename is validated and supplied by the caller.
	file, err := os.Open(fileName)

	if err != nil {
		return "", err
	}

	defer func() {
		err = errors.Join(err, file.Close())
	}()

	fileSize, err := file.Seek(0, io.SeekEnd)

	if err != nil {
		return "", err
	} else if fileSize <= 0 {
		return "", nil
	}

	headSize := min(fileSize, int64(insta360MetadataScanLimit))
	head := make([]byte, headSize)

	if _, err = file.Seek(0, io.SeekStart); err != nil {
		return "", err
	} else if _, err = io.ReadFull(file, head); err != nil {
		return "", fmt.Errorf("read Insta360 metadata header: %w", err)
	} else if bytes.Contains(head, []byte(insta360OneRSImageField)) {
		return insta360OneRSModel, nil
	}

	tailSize := min(fileSize, int64(insta360MetadataScanLimit))
	tail := make([]byte, tailSize)

	if _, err = file.Seek(-tailSize, io.SeekEnd); err != nil {
		return "", err
	} else if _, err = io.ReadFull(file, tail); err != nil {
		return "", fmt.Errorf("read Insta360 metadata trailer: %w", err)
	} else if bytes.Contains(tail, insta360OneRSVideoField) {
		return insta360OneRSModel, nil
	}

	return "", nil
}

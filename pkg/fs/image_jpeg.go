package fs

import (
	"bufio"
	"encoding/binary"
	"errors"
	"io"
	"os"
)

// MaxJpegScans is the largest number of scans accepted in a JPEG image. Set to 0 to disable the check.
var MaxJpegScans = 256

// ErrImageTooComplex is returned when the structure of an image exceeds a supported limit.
var ErrImageTooComplex = errors.New("image structure exceeds the supported limit")

// CheckJpegScans reads the markers of a JPEG stream and returns ErrImageTooComplex when it contains
// more than MaxJpegScans scans. Streams that are not JPEG are ignored.
func CheckJpegScans(reader io.Reader) error {
	if MaxJpegScans <= 0 || reader == nil {
		return nil
	}

	r := bufio.NewReaderSize(reader, 32*1024)

	// The stream starts with an SOI marker; anything else is left to the decoder.
	if soi, err := r.Peek(2); err != nil || soi[0] != 0xFF || soi[1] != 0xD8 {
		return nil
	} else if _, err = r.Discard(2); err != nil {
		return nil
	}

	scans := 0

	for {
		b, err := r.ReadByte()

		if err != nil {
			return nil
		} else if b != 0xFF {
			continue
		}

		// Skip fill bytes before the marker code.
		marker := byte(0xFF)
		for marker == 0xFF {
			if marker, err = r.ReadByte(); err != nil {
				return nil
			}
		}

		switch {
		case marker == 0x00 || marker == 0x01 || marker >= 0xD0 && marker <= 0xD7:
			// Stuffed byte, TEM or restart marker without a length field.
			continue
		case marker == 0xD9:
			return nil
		case marker == 0xDA:
			if scans++; scans > MaxJpegScans {
				return ErrImageTooComplex
			}
		}

		// Every other marker carries a length field that includes its own two bytes. A shorter
		// length skips nothing, and reading continues after the field.
		var size [2]byte
		if _, err = io.ReadFull(r, size[:]); err != nil {
			return nil
		} else if n := int(binary.BigEndian.Uint16(size[:])); n <= 2 {
			continue
		} else if _, err = r.Discard(n - 2); err != nil {
			return nil
		}
	}
}

// CheckJpegScansFile runs CheckJpegScans on the specified file.
func CheckJpegScansFile(fileName string) (err error) {
	file, err := os.Open(fileName) //nolint:gosec // fileName is supplied by the caller and may point to user media
	if err != nil {
		return err
	}

	defer func() {
		if closeErr := file.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}()

	return CheckJpegScans(file)
}

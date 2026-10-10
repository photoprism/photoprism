package fs

import (
	"bufio"
	"encoding/binary"
	"errors"
	"image"
	"image/color"
	"io"
	"os"
)

// MaxJpegScans is the largest number of scans accepted in a JPEG image. Set to 0 to disable the check.
var MaxJpegScans = 256

// ErrImageTooComplex is returned when the structure of an image exceeds a supported limit.
var ErrImageTooComplex = errors.New("image structure exceeds the supported limit")

// CheckJpegScans reads the markers of a JPEG stream and returns ErrImageTooComplex when its first image
// contains more than MaxJpegScans scans. Streams that are not JPEG, or that end early, are ignored.
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
		case marker == 0xD8 || marker == 0xD9:
			// The image ends at EOI, and decoders stop at a second SOI.
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

// jpegFrameConfig returns the dimensions and color model declared by the first frame header of a JPEG
// stream, so that its size is known even when the Go decoder does not support its encoding.
func jpegFrameConfig(reader io.Reader) (image.Config, error) {
	r := bufio.NewReaderSize(reader, 4*1024)

	if soi, err := r.Peek(2); err != nil || soi[0] != 0xFF || soi[1] != 0xD8 {
		return image.Config{}, errUnsupportedImageFormat
	} else if _, err = r.Discard(2); err != nil {
		return image.Config{}, err
	}

	for {
		// Bytes between segments are skipped up to the next marker, as libjpeg does.
		b, err := r.ReadByte()

		if err != nil {
			return image.Config{}, err
		} else if b != 0xFF {
			continue
		}

		marker := byte(0xFF)
		for marker == 0xFF {
			if marker, err = r.ReadByte(); err != nil {
				return image.Config{}, err
			}
		}

		switch {
		case marker == 0x00 || marker == 0x01 || marker >= 0xD0 && marker <= 0xD7:
			continue
		case marker == 0xD8 || marker == 0xD9 || marker == 0xDA:
			return image.Config{}, errUnsupportedImageFormat
		}

		var size [2]byte
		if _, err = io.ReadFull(r, size[:]); err != nil {
			return image.Config{}, err
		}

		n := int(binary.BigEndian.Uint16(size[:]))

		// SOF0 to SOF15, except DHT (C4), JPG (C8) and DAC (CC), declare the frame size.
		if marker >= 0xC0 && marker <= 0xCF && marker != 0xC4 && marker != 0xC8 && marker != 0xCC {
			var frame [6]byte
			if n < 8 {
				return image.Config{}, errUnsupportedImageFormat
			} else if _, err = io.ReadFull(r, frame[:]); err != nil {
				return image.Config{}, err
			} else if components := int(frame[5]); components < 1 || components > 4 || n != 8+3*components {
				return image.Config{}, errUnsupportedImageFormat
			}

			height := int(binary.BigEndian.Uint16(frame[1:3]))
			width := int(binary.BigEndian.Uint16(frame[3:5]))

			// A height of 0 is defined later in the stream, which is not supported here.
			if width == 0 || height == 0 {
				return image.Config{}, errUnsupportedImageFormat
			}

			cfg := image.Config{Width: width, Height: height, ColorModel: color.YCbCrModel}

			switch frame[5] {
			case 1:
				cfg.ColorModel = color.GrayModel
			case 4:
				cfg.ColorModel = color.CMYKModel
			}

			return cfg, nil
		} else if n < 2 {
			continue
		} else if _, err = r.Discard(n - 2); err != nil {
			return image.Config{}, err
		}
	}
}

package thumb

import (
	"bytes"
	"fmt"
	"image"
	"io"
	"os"
	"path/filepath"
	"syscall"

	"github.com/disintegration/imaging"
	"github.com/mandykoh/prism/meta"
	"github.com/mandykoh/prism/meta/autometa"
	"golang.org/x/sys/unix"

	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/colors"
)

var (
	EOI = []byte{0xff, 0xd9}
)

func decode(reader io.Reader, logName string) (md *meta.Data, img image.Image, err error) {
	// Read color metadata.
	md, imgStream, err := autometa.Load(reader)

	if err != nil {
		log.Warnf("thumb: %s in %s (read color metadata)", err, logName)
		img, err = imaging.Decode(reader)
	} else {
		img, err = imaging.Decode(imgStream)
	}
	return md, img, err
}

func attemptRepair(fileReader *os.File) (io.Reader, error) {
	fi, err := fileReader.Stat()
	if err != nil {
		return nil, fmt.Errorf("%s trying to stat() file", err)
	}
	size := int(fi.Size())
	b, err := unix.Mmap(int(fileReader.Fd()), 0, size, syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_PRIVATE)
	if err != nil {
		return nil, fmt.Errorf("%s while mmap()ing file", err)
	}

	// Check for missing EOI.
	if !bytes.Equal(b[size-len(EOI):size], EOI) {
		b = append(b, EOI...)
	}

	return bytes.NewReader(b), nil
}

// OpenJpeg loads a JPEG image from disk, rotates it, and converts the color profile if necessary.
func OpenJpeg(fileName string, orientation int) (image.Image, error) {
	if fileName == "" {
		return nil, fmt.Errorf("filename missing")
	}

	logName := clean.Log(filepath.Base(fileName))

	// Open file.
	fileReader, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer fileReader.Close()

	// Reset file offset.
	// see https://github.com/golang/go/issues/45902#issuecomment-1007953723
	if _, err := fileReader.Seek(0, 0); err != nil {
		return nil, fmt.Errorf("%s on seek", err)
	}

	md, img, err := decode(fileReader, logName)
	if err != nil {
		log.Warnf("%s during initial decoding attempt", err)
		repaired, err := attemptRepair(fileReader)
		if err != nil {
			return nil, fmt.Errorf("%s while trying to recover image", err)
		}
		if md, img, err = decode(repaired, logName); err != nil {
			return nil, fmt.Errorf("%s while decoding after recovery attempt", err)
		}
	}

	// Read ICC profile and convert colors if possible.
	if md != nil {
		if iccProfile, err := md.ICCProfile(); err != nil || iccProfile == nil {
			// Do nothing.
			log.Tracef("thumb: %s has no color profile", logName)
		} else if profile, err := iccProfile.Description(); err == nil && profile != "" {
			log.Tracef("thumb: %s has color profile %s", logName, clean.Log(profile))
			switch {
			case colors.ProfileDisplayP3.Equal(profile):
				img = colors.ToSRGB(img, colors.ProfileDisplayP3)
			}
		}
	}

	// Adjust orientation.
	if orientation > 1 {
		img = Rotate(img, orientation)
	}

	return img, nil
}

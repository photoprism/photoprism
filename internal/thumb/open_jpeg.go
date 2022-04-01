package thumb

import (
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"

	"github.com/disintegration/imaging"
	"github.com/mandykoh/prism/meta/autometa"

	"github.com/photoprism/photoprism/pkg/colors"
	"github.com/photoprism/photoprism/pkg/sanitize"
)

// OpenJpeg loads a JPEG image from disk, rotates it, and converts the color profile if necessary.
func OpenJpeg(fileName string, orientation int) (result image.Image, err error) {
	if fileName == "" {
		return result, fmt.Errorf("filename missing")
	}

	logName := sanitize.Log(filepath.Base(fileName))

	// Open file.
	fileReader, err := os.Open(fileName)

	if err != nil {
		return result, err
	}

	defer fileReader.Close()

	// Read color metadata.
	md, imgStream, err := autometa.Load(fileReader)

	var img image.Image

	if err != nil {
		log.Warnf("resample: %s in %s (read color metadata)", err, logName)
		img, err = imaging.Decode(fileReader)
	} else {
		img, err = imaging.Decode(imgStream)
	}

	if err != nil {
		return result, err
	}

	// Read ICC profile and convert colors if possible.
	if md != nil {
		if iccProfile, err := md.ICCProfile(); err != nil || iccProfile == nil {
			// Do nothing.
			log.Tracef("resample: %s has no color profile", logName)
		} else if profile, err := iccProfile.Description(); err == nil && profile != "" {
			log.Tracef("resample: %s has color profile %s", logName, sanitize.Log(profile))
			switch {
			case colors.ProfileDisplayP3.Equal(profile):
				img = colors.ToSRGB(img, colors.ProfileDisplayP3)
			}
		}
	}

	// Rotate?
	if orientation > 1 {
		img = Rotate(img, orientation)
	}

	return img, nil
}

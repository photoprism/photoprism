package photoprism

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/disintegration/imaging"
	"github.com/dustin/go-humanize/english"

	"github.com/photoprism/photoprism/internal/thumb"
	"github.com/photoprism/photoprism/pkg/capture"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
)

// Bounds returns the media dimensions as image.Rectangle.
func (m *MediaFile) Bounds() image.Rectangle {
	return image.Rectangle{Min: image.Point{}, Max: image.Point{X: m.Width(), Y: m.Height()}}
}

// Thumbnail returns a thumbnail filename.
func (m *MediaFile) Thumbnail(path string, sizeName thumb.Name) (filename string, err error) {
	size, ok := thumb.Sizes[sizeName]

	if !ok {
		log.Errorf("media: invalid type %s", sizeName)
		return "", fmt.Errorf("media: invalid type %s", sizeName)
	}

	// Choose the smallest fitting size if the original image is smaller.
	if size.Fit && m.Bounds().In(size.Bounds()) {
		size = thumb.FitBounds(m.Bounds())
		log.Tracef("media: smallest fitting size for %s is %s (width %d, height %d)", clean.Log(m.RootRelName()), size.Name, size.Width, size.Height)
	}

	thumbName, err := size.FromFile(m.FileName(), m.Hash(), path, m.Orientation())

	if err != nil {
		err = fmt.Errorf("media: failed to create thumbnail for %s (%s)", clean.Log(m.BaseName()), err)
		log.Debug(err)
		return "", err
	}

	return thumbName, nil
}

// Resample returns a resampled image of the file.
func (m *MediaFile) Resample(path string, sizeName thumb.Name) (img image.Image, err error) {
	thumbName, err := m.Thumbnail(path, sizeName)

	if err != nil {
		return nil, err
	}

	return imaging.Open(thumbName)
}

// SkipThumbnailSize tests if the thumbnail size can be skipped,
// e.g. because it is larger than the original.
func (m *MediaFile) SkipThumbnailSize(size thumb.Size) bool {
	return thumb.Skip(size, m.Bounds())
}

// GenerateThumbnails generates thumbnails in the specified storage path,
// existing images are only replaced if the force flag is set to true.
func (m *MediaFile) GenerateThumbnails(thumbPath string, force bool) (err error) {
	if !m.IsPreviewImage() {
		// Skip.
		return
	}

	count := 0
	start := time.Now()

	defer func() {
		switch count {
		case 0:
			log.Debug(capture.Time(start, fmt.Sprintf("media: generated no new thumbnails for %s", clean.Log(m.RootRelName()))))
		default:
			log.Info(capture.Time(start, fmt.Sprintf("media: generated %s for %s", english.Plural(count, "thumbnail", "thumbnails"), clean.Log(m.RootRelName()))))
		}
	}()

	hash := m.Hash()

	var original image.Image
	var srcImage image.Image
	var srcBuffer []byte
	var srcName thumb.Name

	for _, name := range thumb.Names {
		var size thumb.Size
		var fileName string

		if size = thumb.Sizes[name]; size.Uncached() {
			// Exceeds the maximum size of thumbnails to be generated while indexing (--thumb-size).
			continue
		} else if fileName, err = size.FileName(hash, thumbPath); err != nil {
			log.Errorf("media: failed to create %s (%s)", clean.Log(string(name)), err)
			return err
		} else if force || !fs.FileExists(fileName) {
			// Use libvips to generate thumbnails?
			if thumb.Library == thumb.LibVips {
				// Only create a thumbnail if its size does not exceed the size of the original image.
				if m.SkipThumbnailSize(size) {
					continue
				} else if size.Source != "" {
					// Original image filename.
					srcFile := m.FileName()

					// If possible, use existing thumbnail file to create smaller sizes.
					if thumbFile, srcErr := thumb.Sizes[size.Source].FileName(hash, thumbPath); srcErr == nil && fs.FileExistsNotEmpty(thumbFile) {
						srcFile = thumbFile
					}

					// Generate thumbnail with libvips.
					if size.Source == srcName && srcName != "" && srcBuffer != nil {
						_, _, err = thumb.Vips(srcFile, srcBuffer, hash, thumbPath, size.Width, size.Height, size.Options...)
					} else {
						_, _, err = thumb.Vips(srcFile, nil, hash, thumbPath, size.Width, size.Height, size.Options...)
					}

					// Log error, if any.
					if err != nil {
						log.Debugf("vips: %s in %s (generate thumbnail from %s)", err, clean.Log(m.RootRelName()), size.Source)
					}
				} else if _, srcBuffer, err = thumb.Vips(m.FileName(), nil, hash, thumbPath, size.Width, size.Height, size.Options...); err != nil || srcBuffer == nil {
					log.Debugf("vips: failed to generate thumbnail from %s", clean.Log(m.RootRelName()))
					// Clear thumbnail image cache.
					srcBuffer = nil
					srcName = ""
				} else {
					// Remember cached thumbnail size.
					srcName = name
				}
			} else {
				// Open original if needed.
				if original == nil {
					img, imgErr := thumb.Open(m.FileName(), m.Orientation())

					// Failed to open the JPEG file?
					if imgErr != nil {
						msg := imgErr.Error()

						// Non-repairable file error?
						if !(strings.Contains(msg, "EOF") ||
							strings.HasPrefix(msg, "invalid JPEG")) {
							log.Debugf("media: %s in %s", msg, clean.Log(m.RootRelName()))
							return imgErr
						}

						// Try to repair the file by creating a properly encoded copy with ImageMagick.
						if fixed, fixErr := NewConvert(conf).FixJpeg(m, false); fixErr != nil {
							return fixErr
						} else if fixedImg, openErr := thumb.Open(fixed.FileName(), m.Orientation()); openErr != nil {
							return openErr
						} else {
							img = fixedImg
						}
					}

					original = img

					log.Debugf("media: opened %s [%s]", clean.Log(m.RootRelName()), thumb.MemSize(original).String())
				}

				// Thumb size too large
				// for the original image?
				if size.Skip(original) {
					continue
				}

				// Reuse existing thumbnails to improve rendering performance.
				if size.Source != "" {
					if size.Source == srcName && srcImage != nil {
						_, err = size.Create(srcImage, fileName)
					} else {
						_, err = size.Create(original, fileName)
					}
				} else {
					srcImage, err = size.Create(original, fileName)
					srcName = name
				}
			}

			// Failed?
			if err != nil {
				log.Errorf("media: failed to create %s (%s)", name.String(), err)
				return err
			}

			count++
		}
	}

	return nil
}

// ChangeOrientation changes the file orientation.
func (m *MediaFile) ChangeOrientation(val int) (err error) {
	if !m.IsPreviewImage() {
		// Skip.
		return fmt.Errorf("orientation can currently only be changed for jpeg and png files")
	}

	cnf := Config()
	cmd := exec.Command(cnf.ExifToolBin(), "-overwrite_original", "-n", "-Orientation="+strconv.Itoa(val), m.FileName())

	// Fetch command output.
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	cmd.Env = []string{fmt.Sprintf("HOME=%s", cnf.CmdCachePath())}

	// Log exact command for debugging in trace mode.
	log.Trace(cmd.String())

	// Run exiftool command.
	if err = cmd.Run(); err != nil {
		if stderr.String() != "" {
			return errors.New(stderr.String())
		} else {
			return err
		}
	}

	return nil
}

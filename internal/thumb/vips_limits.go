package thumb

import (
	"fmt"
	"image/jpeg"
	"io"

	"github.com/davidbyttow/govips/v2/vips"

	"github.com/photoprism/photoprism/pkg/fs"
)

// vipsLoadedPages returns how many pages of an image libvips actually loaded.
//
// Pages() reports what the file declares, which is not what was read: the import parameters
// ask for a single page, so a multi-page scan or an animated image is loaded one frame at a
// time. Counting the declared pages would measure work that is never done.
func vipsLoadedPages(img *vips.ImageRef) int {
	height, pageHeight := img.Height(), img.PageHeight()

	if pageHeight <= 0 || height <= pageHeight {
		return 1
	}

	return height / pageHeight
}

// vipsCheckPixels reports an error when a loaded image exceeds the supported pixel budget.
// libvips reads the header first and defers the pixel work, so calling this directly after
// loading rejects an oversized geometry before anything is materialized from it.
func vipsCheckPixels(img *vips.ImageRef, logName string) error {
	if img == nil {
		return fmt.Errorf("vips: image not loaded")
	}

	if err := fs.ExceedsPixelBudget(img.Width(), img.PageHeight(), vipsLoadedPages(img)); err != nil {
		return fmt.Errorf("%w in %s", err, logName)
	}

	return nil
}

// checkJpegPixels reads the geometry from a JPEG header and reports an error when it exceeds
// the supported pixel budget, leaving the reader positioned at the start either way.
func checkJpegPixels(reader io.ReadSeeker, logName string) error {
	if _, err := reader.Seek(0, io.SeekStart); err != nil {
		return err
	}

	cfg, cfgErr := jpeg.DecodeConfig(reader)

	if _, err := reader.Seek(0, io.SeekStart); err != nil {
		return err
	}

	// A header that cannot be read is left to the decoder, which reports the real reason.
	if cfgErr != nil {
		return nil
	}

	if err := fs.ExceedsPixelBudget(cfg.Width, cfg.Height, 1); err != nil {
		return fmt.Errorf("%w in %s", err, logName)
	}

	return nil
}

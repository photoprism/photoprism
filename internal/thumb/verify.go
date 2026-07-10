package thumb

import (
	"fmt"

	"github.com/davidbyttow/govips/v2/vips"
)

// VerifySize is the target edge length used to force a thumbnail decode during Verify;
// it matches the standard tile size whose center-crop decode rejects corrupt previews.
const VerifySize = 500

// Verify reports an error if the image file cannot be decoded by the active rendering library.
// Used to reject corrupt converter output (e.g. a bogus-Huffman embedded RAW preview that passes a
// MIME sniff but later fails libvips) so the conversion loop can try the next converter.
func Verify(fileName string) error {
	if fileName == "" {
		return fmt.Errorf("verify: empty filename")
	}

	// Run the same shrink-on-load path the thumbnailer uses so the check matches GenerateThumbnails.
	if Library == LibVips {
		VipsInit()

		img, err := vips.LoadImageFromFile(fileName, VipsImportParams())
		if err != nil {
			return err
		}

		defer img.Close()

		// libvips loads JPEGs lazily, so force the decode now (it would otherwise fail later in
		// GenerateThumbnails and mark the photo IndexFailed). Mirror the center-crop tile path,
		// which rejects bogus-Huffman previews a plain fit resize can skip.
		return img.ThumbnailWithSize(VerifySize, VerifySize, vips.InterestingCentre, vips.SizeBoth)
	}

	_, err := Open(fileName, 1)

	return err
}

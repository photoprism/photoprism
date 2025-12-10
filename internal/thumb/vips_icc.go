package thumb

import (
	"fmt"

	"github.com/davidbyttow/govips/v2/vips"
)

// InteroperabilityIndex EXIF codes used by some cameras to hint at the color space.
// Sources: EXIF TagNames (R03=Adobe RGB, R98=sRGB, THM=thumbnail)
// https://unpkg.com/exiftool-vendored.pl@10.50.0/bin/html/TagNames/EXIF.html
// Additional context: https://regex.info/blog/photo-tech/color-spaces-page7
// Note: exiftool refers to this tag as "InteropIndex"; in libvips/govips the
// corresponding field name is "exif-ifd4-InteroperabilityIndex".
// Exiftool example: exiftool -InteropIndex -InteropVersion -icc_profile:all -G -s file.jpg
const (
	// InteropIndexAdobeRGB is the EXIF code for Adobe RGB (1998) ("R03").
	InteropIndexAdobeRGB = "R03"
	// InteropIndexSRGB is the EXIF code for sRGB ("R98").
	InteropIndexSRGB = "R98"
	// InteropIndexThumb marks a thumbnail image; treated as sRGB ("THM").
	InteropIndexThumb = "THM"
)

// vipsSetIccProfileForInteropIndex embeds an ICC profile when a JPEG declares
// its color space via the EXIF InteroperabilityIndex tag (e.g., "R03"/Adobe RGB)
// but lacks an embedded profile. If an ICC profile is already present, it
// leaves the image untouched.
func vipsSetIccProfileForInteropIndex(img *vips.ImageRef, logName string) (err error) {
	// Some cameras signal color space via EXIF InteroperabilityIndex instead of
	// embedding an ICC profile. Browsers and libvips ignore this tag, so we
	// inject a matching ICC profile to produce correct thumbnails.
	iiFull := img.GetString("exif-ifd4-InteroperabilityIndex")

	if iiFull == "" {
		return nil
	}

	// EXIF InteroperabilityIndex is 4 bytes including null; libvips returns
	// a string with a trailing space. Using the first three bytes covers the
	// meaningful code (e.g., "R03", "R98").
	if len(iiFull) < 3 {
		log.Debugf("interopindex: %s has unexpected interop index %q", logName, iiFull)
		return nil
	}

	ii := iiFull[:3]
	log.Tracef("interopindex: %s read exif and got interopindex %s, %s", logName, ii, iiFull)

	if img.HasICCProfile() {
		log.Debugf("interopindex: %s already has an embedded ICC profile; skipping fallback.", logName)
		return nil
	}

	profilePath := ""

	switch ii {
	case InteropIndexAdobeRGB:
		// Use Adobe RGB 1998 compatible profile.
		profilePath, err = GetIccProfile(IccAdobeRGBCompat, IccAdobeRGBCompatV2, IccAdobeRGBCompatV4)

		if err != nil {
			return fmt.Errorf("interopindex %s: %w", ii, err)
		}
	case InteropIndexSRGB:
		// sRGB: browsers and libvips assume sRGB by default, so no embed needed.
	case InteropIndexThumb:
		// Thumbnail file; specification unclear—treat as sRGB and do nothing.
	default:
		log.Debugf("interopindex: %s has unknown interop index %s", logName, ii)
	}

	if profilePath == "" {
		return nil
	}

	// Embed ICC profile. govips expects both an input (fallback) and output
	// profile; using the same path injects the chosen profile when none is
	// embedded and keeps colors consistent otherwise.
	return img.TransformICCProfileWithFallback(profilePath, profilePath)
}

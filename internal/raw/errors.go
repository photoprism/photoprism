package raw

import "strings"

// WhiteBalanceError is the RawTherapee/dcraw stderr message indicating it could not read the camera
// white balance, so its colors are untrustworthy and the embedded camera preview is preferred.
const WhiteBalanceError = "Cannot use camera white balance"

// DecoderErrors lists RawTherapee/dcraw stderr substrings that signal an untrustworthy render —
// wrong colors or a corrupt decode — so callers can fall back to the embedded camera preview.
// Sourced from dcraw.c; verbose progress lines are deliberately excluded.
var DecoderErrors = []string{
	WhiteBalanceError,           // no camera white balance coefficients: defaults applied, wrong colors
	"matrix not found",          // color matrix missing: generic colors, wrong rendering
	"Invalid white balance",     // unknown white balance preset: wrong colors
	"Corrupt data near",         // corrupt raw stream: unreliable pixels
	"Unexpected end of file",    // truncated raw: unreliable pixels
	"decoder table overflow",    // Huffman decoder overflow: unreliable pixels
	"incorrect JPEG dimensions", // lossless-JPEG raw mismatch: unreliable pixels
}

// Untrustworthy reports whether RAW-decoder stderr output indicates the render cannot be trusted.
func Untrustworthy(stderr string) bool {
	if stderr == "" {
		return false
	}

	for _, pattern := range DecoderErrors {
		if strings.Contains(stderr, pattern) {
			return true
		}
	}

	return false
}

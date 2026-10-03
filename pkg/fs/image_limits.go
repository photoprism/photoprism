package fs

import (
	"errors"
	"fmt"
	"math"
)

// ErrImageTooLarge is returned when an image geometry exceeds MaxImagePixels.
var ErrImageTooLarge = errors.New("image resolution exceeds the supported limit")

// MaxImagePixels limits the number of pixels decoders are asked to materialize, counted over
// all frames that are read. The geometry is validated before the pixel data, so an image is
// only decoded at a resolution the configured limit allows. Set to 0 to disable the check.
var MaxImagePixels = 150000000

// ExceedsPixelBudget returns ErrImageTooLarge when the given geometry exceeds MaxImagePixels.
// Frames below one count as a single frame, so that a format reporting no frame count is
// bounded by its own geometry rather than skipped.
func ExceedsPixelBudget(width, height, frames int) error {
	if MaxImagePixels <= 0 {
		return nil
	}

	if width <= 0 || height <= 0 {
		return nil
	}

	if frames < 1 {
		frames = 1
	}

	budget := int64(MaxImagePixels)

	// Counted in 64-bit and one factor at a time, since three plausible dimensions multiply to
	// more than an int64 holds and a wrapped product would read as a small positive number.
	pixels := int64(width) * int64(height)

	if pixels > budget || pixels > math.MaxInt64/int64(frames) {
		return fmt.Errorf("%w (%dx%d, %d frames)", ErrImageTooLarge, width, height, frames)
	}

	if pixels*int64(frames) > budget {
		return fmt.Errorf("%w (%dx%d, %d frames)", ErrImageTooLarge, width, height, frames)
	}

	return nil
}

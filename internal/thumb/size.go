package thumb

import (
	"image"
)

// Size represents a standard media resolution.
type Size struct {
	Name    Name             `json:"name"`
	Source  Name             `json:"-"`
	Usage   string           `json:"usage"`
	Width   int              `json:"w"`
	Height  int              `json:"h"`
	Public  bool             `json:"-"`
	Fit     bool             `json:"-"`
	Options []ResampleOption `json:"-"`
}

// Bounds returns the thumb size as image.Rectangle.
func (s Size) Bounds() image.Rectangle {
	return image.Rectangle{Min: image.Point{}, Max: image.Point{X: s.Width, Y: s.Height}}
}

// Uncached tests if thumbnail type exceeds the cached thumbnails size limit.
func (s Size) Uncached() bool {
	return s.Width > SizePrecached || s.Height > SizePrecached
}

// ExceedsLimit tests if thumbnail type is too large, and can not be rendered at all.
func (s Size) ExceedsLimit() bool {
	return s.Width > MaxSize() || s.Height > MaxSize()
}

// FromCache returns the filename if a thumbnail image with the matching size is in the cache.
func (s Size) FromCache(fileName, fileHash, cachePath string) (string, error) {
	return FromCache(fileName, fileHash, cachePath, s.Width, s.Height, s.Options...)
}

// FromFile creates a new thumbnail with the matching size if it was not found in the cache, and returns the filename.
func (s Size) FromFile(fileName, fileHash, cachePath string, fileOrientation int) (string, error) {
	return FromFile(fileName, fileHash, cachePath, s.Width, s.Height, fileOrientation, s.Options...)
}

// Create creates a thumbnail with the matching size and returns it as image.Image.
func (s Size) Create(img image.Image, fileName string) (image.Image, error) {
	return Create(img, fileName, s.Width, s.Height, s.Options...)
}

// FileName returns the file name of the thumbnail for the matching size.
func (s Size) FileName(hash, thumbPath string) (string, error) {
	return FileName(hash, thumbPath, s.Width, s.Height, s.Options...)
}

// ResolvedName returns the file name of the thumbnail for the matching size with all symlinks resolved.
func (s Size) ResolvedName(hash, thumbPath string) (string, error) {
	return ResolvedName(hash, thumbPath, s.Width, s.Height, s.Options...)
}

// Skip checks if the thumbnail size is too large for the image and can be skipped.
func (s Size) Skip(img image.Image) bool {
	if !s.Fit || !img.Bounds().In(s.Bounds()) {
		return false
	} else if newSize := FitBounds(img.Bounds()); newSize.Width < s.Width {
		return true
	}

	return false
}

package thumb

import (
	"fmt"
	"image"
	"image/draw"
	"path/filepath"

	xdraw "golang.org/x/image/draw"

	"github.com/davidbyttow/govips/v2/vips"

	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
)

// InputSource holds an oriented image until its selected region is resampled.
type InputSource struct {
	pixels image.Image
	native *vips.ImageRef
}

// NewInputSource wraps decoded pixels for input resampling.
func NewInputSource(pixels image.Image) *InputSource { return &InputSource{pixels: pixels} }

// Bounds returns the oriented source dimensions.
func (s *InputSource) Bounds() image.Rectangle {
	if s.native != nil {
		return image.Rect(0, 0, s.native.Width(), s.native.Height())
	}
	if s.pixels != nil {
		return s.pixels.Bounds()
	}
	return image.Rectangle{}
}

// Close releases the native source when it is no longer needed.
func (s *InputSource) Close() {
	if s.native != nil {
		s.native.Close()
		s.native = nil
	}
	s.pixels = nil
}

// Resample crops and resizes a source once; the caller must then close it.
func (s *InputSource) Resample(region image.Rectangle, width, height int) (image.Image, error) {
	if region.Empty() || !region.In(s.Bounds()) || width <= 0 || height <= 0 || InvalidSize(width) || InvalidSize(height) {
		return nil, fmt.Errorf("invalid input region or dimensions")
	}
	if s.native != nil {
		if err := s.native.ExtractArea(region.Min.X, region.Min.Y, region.Dx(), region.Dy()); err != nil {
			return nil, vipsErr(err)
		}
		if region.Dx() != width || region.Dy() != height {
			if err := s.native.ResizeWithVScale(float64(width)/float64(region.Dx()), float64(height)/float64(region.Dy()), vips.KernelCubic); err != nil {
				return nil, vipsErr(err)
			}
		}
		return s.native.ToGoImage()
	}
	pixels := image.NewNRGBA(image.Rect(0, 0, width, height))
	xdraw.CatmullRom.Scale(pixels, pixels.Bounds(), s.pixels, region, draw.Src, nil)
	return pixels, nil
}

// OpenInputSource opens an oriented source without materializing native pixels.
func OpenInputSource(fileName string, orientation int) (_ *InputSource, err error) {
	if Library != LibVips {
		img, openErr := Open(fileName, orientation)
		return NewInputSource(img), openErr
	}
	defer func() { err = vipsErr(err) }()
	fileName, err = fs.Resolve(fileName)
	if err != nil {
		return nil, err
	}
	VipsInit()
	img, err := vips.LoadImageFromFile(fileName, vipsConvertImportParams())
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			img.Close()
		}
	}()
	logName := clean.Log(filepath.Base(fileName))
	if err = vipsCheckPixels(img, logName); err != nil {
		return nil, err
	}
	if orientation > OrientationNormal && !vipsLoadedViaHeif(img) {
		if err = VipsRotate(img, orientation); err != nil {
			return nil, err
		}
	}
	if Color != ColorNone {
		if profileErr := vipsSetIccProfileForInteropIndex(img, logName); profileErr != nil {
			log.Debugf("vips: %s (input color profile)", vipsErr(profileErr))
		}
		if img.HasICCProfile() {
			if profileErr := img.TransformICCProfile("srgb"); profileErr != nil {
				log.Debugf("vips: %s (input color conversion)", vipsErr(profileErr))
			}
		}
	}
	return &InputSource{native: img}, nil
}

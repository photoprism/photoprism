package colors

import (
	"image"
	_ "image/jpeg"
	"runtime"

	"github.com/mandykoh/prism"
	"github.com/mandykoh/prism/displayp3"
	"github.com/mandykoh/prism/srgb"
)

// ToSRGB converts an image to sRGB colors.
func ToSRGB(img image.Image, profile Profile) image.Image {
	switch profile {
	case ProfileDisplayP3:
		in := prism.ConvertImageToNRGBA(img, runtime.NumCPU())
		out := image.NewNRGBA(in.Rect)

		for i := in.Rect.Min.Y; i < in.Rect.Max.Y; i++ {
			for j := in.Rect.Min.X; j < in.Rect.Max.X; j++ {
				inCol, alpha := displayp3.ColorFromNRGBA(in.NRGBAAt(j, i))
				outCol := srgb.ColorFromXYZ(inCol.ToXYZ())
				out.SetNRGBA(j, i, outCol.ToNRGBA(alpha))
			}
		}
		return out
	default:
		return img
	}
}

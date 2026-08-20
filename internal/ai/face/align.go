package face

import (
	"fmt"
	"image"
	"image/color"
	"math"
)

// NumLandmarks is the number of facial landmark points the detector predicts per face.
const NumLandmarks = 5

// ArcFaceTemplateSize is the crop size, in pixels, that ArcFaceTemplate refers to.
const ArcFaceTemplateSize = 112

// maxAlignResidual is the largest RMS landmark fit error, in template pixels, that still
// yields an aligned crop. Above it the caller falls back to a plain bounding box crop.
const maxAlignResidual = 14.0

// landmarkDisplayScale divides the face size to derive the display size stored with each landmark.
const landmarkDisplayScale = 10

// LandmarkNames lists the landmark names in the order the detector returns them.
// Left and right refer to the image rather than the depicted person, so the names
// line up with ArcFaceTemplate.
var LandmarkNames = [NumLandmarks]string{"eye_l", "eye_r", "nose", "mouth_l", "mouth_r"}

// ArcFaceTemplate holds the reference landmark positions, in x/y order, of the
// standard 112x112 alignment that InsightFace and OpenCV both use. Warping crops
// onto it is what makes their pretrained models produce comparable embeddings.
var ArcFaceTemplate = [NumLandmarks][2]float64{
	{38.2946, 51.6963},
	{73.5318, 51.5014},
	{56.0252, 71.7366},
	{41.5493, 92.3655},
	{70.7299, 92.2041},
}

// affine2D is a 2x3 affine transform matrix in row-major order.
type affine2D [6]float64

// LandmarkAreas converts detector landmark coordinates into eye and landmark areas.
// Eyes are returned separately because Face.EyesMidpoint uses them as the origin for
// relative landmark coordinates.
func LandmarkAreas(kps [NumLandmarks * 2]float32, size int) (eyes, landmarks Areas) {
	scale := max(size/landmarkDisplayScale, 1)

	eyes = make(Areas, 0, 2)
	landmarks = make(Areas, 0, NumLandmarks-2)

	for i, name := range LandmarkNames {
		col := int(math.Round(float64(kps[i*2])))
		row := int(math.Round(float64(kps[i*2+1])))

		if i < 2 {
			eyes = append(eyes, NewArea(name, row, col, scale))
		} else {
			landmarks = append(landmarks, NewArea(name, row, col, scale))
		}
	}

	return eyes, landmarks
}

// AlignPoints returns the landmark positions required for alignment in template order,
// and reports false unless all of them were detected.
func (f *Face) AlignPoints() (pts [NumLandmarks][2]float64, ok bool) {
	if f == nil {
		return pts, false
	}

	var found [NumLandmarks]bool

	for _, areas := range []Areas{f.Eyes, f.Landmarks} {
		for _, a := range areas {
			for i, name := range LandmarkNames {
				if found[i] || a.Name != name {
					continue
				}

				pts[i][0] = float64(a.Col)
				pts[i][1] = float64(a.Row)
				found[i] = true
			}
		}
	}

	for i := range found {
		if !found[i] {
			return pts, false
		}
	}

	return pts, true
}

// ScaledArcFaceTemplate returns the ArcFace landmark template scaled to a crop size.
func ScaledArcFaceTemplate(width, height int) (dst [NumLandmarks][2]float64) {
	sx := float64(width) / ArcFaceTemplateSize
	sy := float64(height) / ArcFaceTemplateSize

	for i, p := range ArcFaceTemplate {
		dst[i][0] = p[0] * sx
		dst[i][1] = p[1] * sy
	}

	return dst
}

// AlignedCrop warps a face onto the ArcFace template and returns a crop of the
// requested size. It fails when the face has no complete landmark set so callers
// can fall back to an unaligned bounding box crop.
//
// The image may be a larger rendition than the one the face was detected in, which is how
// a small face avoids being upscaled from the detection thumbnail. Landmarks are absolute
// coordinates, so they are scaled by the width ratio the face records.
func AlignedCrop(img image.Image, f *Face, width, height int) (*image.RGBA, error) {
	if img == nil {
		return nil, fmt.Errorf("faces: missing image")
	} else if width < 1 || height < 1 {
		return nil, fmt.Errorf("faces: invalid crop size %dx%d", width, height)
	}

	src, ok := f.AlignPoints()

	if !ok {
		return nil, fmt.Errorf("faces: incomplete landmarks")
	}

	if scale := f.ImageScale(img.Bounds().Dx()); scale != 1 {
		for i := range src {
			src[i][0] *= scale
			src[i][1] *= scale
		}
	}

	forward, err := similarityTransform(src, ScaledArcFaceTemplate(width, height))

	if err != nil {
		return nil, err
	}

	// Sampling maps every destination pixel back into the source image.
	inverse, err := forward.invert()

	if err != nil {
		return nil, err
	}

	bounds := img.Bounds()
	out := image.NewRGBA(image.Rect(0, 0, width, height))

	for y := range height {
		for x := range width {
			sx, sy := inverse.apply(float64(x), float64(y))
			out.SetRGBA(x, y, sampleBilinear(img, bounds, sx, sy))
		}
	}

	return out, nil
}

// similarityTransform returns the transform that maps src onto dst with the smallest
// squared error, using rotation, uniform scale, and translation only.
//
// This is the Umeyama estimate that InsightFace and OpenCV apply, restricted to proper
// rotations. A reflection cannot be represented, so a mirrored set produces a poor fit
// rather than a mirrored crop; both that and coincident landmarks return an error.
func similarityTransform(src, dst [NumLandmarks][2]float64) (affine2D, error) {
	var srcMean, dstMean [2]float64

	for i := range src {
		srcMean[0] += src[i][0]
		srcMean[1] += src[i][1]
		dstMean[0] += dst[i][0]
		dstMean[1] += dst[i][1]
	}

	n := float64(NumLandmarks)

	for i := range srcMean {
		srcMean[i] /= n
		dstMean[i] /= n
	}

	var a, b, variance float64

	for i := range src {
		sx := src[i][0] - srcMean[0]
		sy := src[i][1] - srcMean[1]
		dx := dst[i][0] - dstMean[0]
		dy := dst[i][1] - dstMean[1]

		a += sx*dx + sy*dy
		b += sx*dy - sy*dx
		variance += sx*sx + sy*sy
	}

	if variance <= 0 {
		return affine2D{}, fmt.Errorf("faces: coincident landmarks")
	}

	a /= variance
	b /= variance

	result := affine2D{
		a, -b, dstMean[0] - (a*srcMean[0] - b*srcMean[1]),
		b, a, dstMean[1] - (b*srcMean[0] + a*srcMean[1]),
	}

	// A mirrored set cannot be expressed by a proper rotation, so it comes back as a poor
	// fit rather than an error. Measured on the template: a rotated or scaled copy fits at
	// 0, an extreme yaw at ~9.9, and a horizontal mirror at ~22.6.
	if residual := fitResidual(result, src, dst); residual > maxAlignResidual {
		return affine2D{}, fmt.Errorf("faces: landmarks do not fit the template, residual %.1f", residual)
	}

	return result, nil
}

// fitResidual returns the RMS distance between the transformed source landmarks and the
// template they were fitted to.
func fitResidual(t affine2D, src, dst [NumLandmarks][2]float64) float64 {
	var sum float64

	for i := range src {
		dx := t[0]*src[i][0] + t[1]*src[i][1] + t[2] - dst[i][0]
		dy := t[3]*src[i][0] + t[4]*src[i][1] + t[5] - dst[i][1]
		sum += dx*dx + dy*dy
	}

	return math.Sqrt(sum / float64(NumLandmarks))
}

// apply maps a point through the transform.
func (m affine2D) apply(x, y float64) (float64, float64) {
	return m[0]*x + m[1]*y + m[2], m[3]*x + m[4]*y + m[5]
}

// invert returns the inverse transform, or an error when the matrix is singular.
func (m affine2D) invert() (affine2D, error) {
	det := m[0]*m[4] - m[1]*m[3]

	if det == 0 {
		return affine2D{}, fmt.Errorf("faces: singular transform matrix")
	}

	inv := 1 / det
	i0 := m[4] * inv
	i1 := -m[1] * inv
	i3 := -m[3] * inv
	i4 := m[0] * inv

	return affine2D{
		i0, i1, -(i0*m[2] + i1*m[5]),
		i3, i4, -(i3*m[2] + i4*m[5]),
	}, nil
}

// sampleBilinear interpolates the color at a source position, treating everything
// outside the image as transparent black to match the border that the reference
// implementations fill with warpAffine.
func sampleBilinear(img image.Image, bounds image.Rectangle, x, y float64) color.RGBA {
	x0 := int(math.Floor(x))
	y0 := int(math.Floor(y))
	fx := x - float64(x0)
	fy := y - float64(y0)

	weights := [4]struct {
		dx, dy int
		w      float64
	}{
		{0, 0, (1 - fx) * (1 - fy)},
		{1, 0, fx * (1 - fy)},
		{0, 1, (1 - fx) * fy},
		{1, 1, fx * fy},
	}

	var r, g, b, a float64

	for _, s := range weights {
		if s.w <= 0 {
			continue
		}

		px := x0 + s.dx
		py := y0 + s.dy

		if px < bounds.Min.X || px >= bounds.Max.X || py < bounds.Min.Y || py >= bounds.Max.Y {
			continue
		}

		// Interpolating premultiplied values keeps partially transparent borders correct.
		cr, cg, cb, ca := img.At(px, py).RGBA()

		r += float64(cr>>8) * s.w
		g += float64(cg>>8) * s.w
		b += float64(cb>>8) * s.w
		a += float64(ca>>8) * s.w
	}

	return color.RGBA{R: clampUint8(r), G: clampUint8(g), B: clampUint8(b), A: clampUint8(a)}
}

// clampUint8 rounds a channel value and bounds it to the uint8 range.
func clampUint8(v float64) uint8 {
	switch {
	case v <= 0:
		return 0
	case v >= 255:
		return 255
	default:
		return uint8(math.Round(v))
	}
}

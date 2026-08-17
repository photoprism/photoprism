package face

import (
	"image"
	"image/color"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testFaceWithLandmarks returns a face whose landmarks are set to the given x/y points.
func testFaceWithLandmarks(pts [NumLandmarks][2]float64, size int) *Face {
	var kps [NumLandmarks * 2]float32

	for i, p := range pts {
		kps[i*2] = float32(p[0])
		kps[i*2+1] = float32(p[1])
	}

	eyes, landmarks := LandmarkAreas(kps, size)

	return &Face{Rows: 300, Cols: 300, Area: NewArea("face", 150, 150, size), Eyes: eyes, Landmarks: landmarks}
}

// transformPoints applies a rotation, scale, and translation to the ArcFace template.
func transformPoints(angle, scale, tx, ty float64) (pts [NumLandmarks][2]float64) {
	cos := math.Cos(angle) * scale
	sin := math.Sin(angle) * scale

	for i, p := range ArcFaceTemplate {
		pts[i][0] = cos*p[0] - sin*p[1] + tx
		pts[i][1] = sin*p[0] + cos*p[1] + ty
	}

	return pts
}

func TestLandmarkAreas(t *testing.T) {
	kps := [NumLandmarks * 2]float32{10.4, 20.6, 30, 21, 20, 30, 12, 40, 28, 41}
	eyes, landmarks := LandmarkAreas(kps, 50)

	t.Run("Eyes", func(t *testing.T) {
		require.Len(t, eyes, 2)
		assert.Equal(t, "eye_l", eyes[0].Name)
		assert.Equal(t, 10, eyes[0].Col)
		assert.Equal(t, 21, eyes[0].Row)
		assert.Equal(t, "eye_r", eyes[1].Name)
	})
	t.Run("Landmarks", func(t *testing.T) {
		require.Len(t, landmarks, 3)
		assert.Equal(t, "nose", landmarks[0].Name)
		assert.Equal(t, "mouth_l", landmarks[1].Name)
		assert.Equal(t, "mouth_r", landmarks[2].Name)
	})
	t.Run("DisplayScale", func(t *testing.T) {
		assert.Equal(t, 5, eyes[0].Scale)
	})
	t.Run("MinimumScale", func(t *testing.T) {
		small, _ := LandmarkAreas(kps, 4)
		assert.Equal(t, 1, small[0].Scale)
	})
}

func TestFace_AlignPoints(t *testing.T) {
	t.Run("Complete", func(t *testing.T) {
		expected := transformPoints(0, 1, 0, 0)
		f := testFaceWithLandmarks(expected, 100)
		pts, ok := f.AlignPoints()
		require.True(t, ok)

		for i := range expected {
			assert.InDelta(t, expected[i][0], pts[i][0], 1)
			assert.InDelta(t, expected[i][1], pts[i][1], 1)
		}
	})
	t.Run("MissingLandmarks", func(t *testing.T) {
		f := testFaceWithLandmarks(transformPoints(0, 1, 0, 0), 100)
		f.Landmarks = f.Landmarks[:1]
		_, ok := f.AlignPoints()
		assert.False(t, ok)
	})
	t.Run("MissingEyes", func(t *testing.T) {
		f := testFaceWithLandmarks(transformPoints(0, 1, 0, 0), 100)
		f.Eyes = nil
		_, ok := f.AlignPoints()
		assert.False(t, ok)
	})
	t.Run("NoLandmarks", func(t *testing.T) {
		f := &Face{}
		_, ok := f.AlignPoints()
		assert.False(t, ok)
	})
	t.Run("NilFace", func(t *testing.T) {
		var f *Face
		_, ok := f.AlignPoints()
		assert.False(t, ok)
	})
}

func TestScaledArcFaceTemplate(t *testing.T) {
	t.Run("NativeSize", func(t *testing.T) {
		dst := ScaledArcFaceTemplate(ArcFaceTemplateSize, ArcFaceTemplateSize)
		assert.Equal(t, ArcFaceTemplate, dst)
	})
	t.Run("Doubled", func(t *testing.T) {
		dst := ScaledArcFaceTemplate(224, 224)
		assert.InDelta(t, ArcFaceTemplate[0][0]*2, dst[0][0], 0.0001)
		assert.InDelta(t, ArcFaceTemplate[4][1]*2, dst[4][1], 0.0001)
	})
}

func TestSimilarityTransform(t *testing.T) {
	t.Run("Identity", func(t *testing.T) {
		m, err := similarityTransform(ArcFaceTemplate, ArcFaceTemplate)
		require.NoError(t, err)
		assert.InDelta(t, 1, m[0], 0.0001)
		assert.InDelta(t, 0, m[1], 0.0001)
		assert.InDelta(t, 0, m[2], 0.0001)
		assert.InDelta(t, 0, m[3], 0.0001)
		assert.InDelta(t, 1, m[4], 0.0001)
		assert.InDelta(t, 0, m[5], 0.0001)
	})
	t.Run("RecoversKnownTransform", func(t *testing.T) {
		// A transform fitted to warped points must map them back onto the template.
		src := transformPoints(math.Pi/6, 1.5, 40, 25)
		m, err := similarityTransform(src, ArcFaceTemplate)
		require.NoError(t, err)

		for i := range src {
			x, y := m.apply(src[i][0], src[i][1])
			assert.InDelta(t, ArcFaceTemplate[i][0], x, 0.0001)
			assert.InDelta(t, ArcFaceTemplate[i][1], y, 0.0001)
		}
	})
	t.Run("CoincidentLandmarks", func(t *testing.T) {
		var src [NumLandmarks][2]float64

		for i := range src {
			src[i] = [2]float64{10, 10}
		}

		_, err := similarityTransform(src, ArcFaceTemplate)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "coincident")
	})
}

func TestAffine2D(t *testing.T) {
	m := affine2D{2, 0, 5, 0, 2, -3}

	t.Run("Apply", func(t *testing.T) {
		x, y := m.apply(1, 1)
		assert.InDelta(t, 7, x, 0.0001)
		assert.InDelta(t, -1, y, 0.0001)
	})
	t.Run("InvertRoundTrip", func(t *testing.T) {
		inv, err := m.invert()
		require.NoError(t, err)
		x, y := m.apply(3, 4)
		bx, by := inv.apply(x, y)
		assert.InDelta(t, 3, bx, 0.0001)
		assert.InDelta(t, 4, by, 0.0001)
	})
	t.Run("Singular", func(t *testing.T) {
		_, err := affine2D{0, 0, 0, 0, 0, 0}.invert()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "singular")
	})
}

func TestAlignedCrop(t *testing.T) {
	// Paint a marker at each landmark of a rotated and scaled face, then verify that
	// alignment moves the markers onto the template positions.
	src := transformPoints(math.Pi/9, 1.8, 60, 40)
	img := image.NewRGBA(image.Rect(0, 0, 300, 300))

	for y := range 300 {
		for x := range 300 {
			img.SetRGBA(x, y, color.RGBA{R: 20, G: 20, B: 20, A: 255})
		}
	}

	for _, p := range src {
		cx := int(math.Round(p[0]))
		cy := int(math.Round(p[1]))

		for dy := -2; dy <= 2; dy++ {
			for dx := -2; dx <= 2; dx++ {
				img.SetRGBA(cx+dx, cy+dy, color.RGBA{R: 255, G: 255, B: 255, A: 255})
			}
		}
	}

	f := testFaceWithLandmarks(src, 120)

	t.Run("MarkersMatchTemplate", func(t *testing.T) {
		out, err := AlignedCrop(img, f, ArcFaceTemplateSize, ArcFaceTemplateSize)
		require.NoError(t, err)
		require.Equal(t, image.Rect(0, 0, ArcFaceTemplateSize, ArcFaceTemplateSize), out.Bounds())

		for i, p := range ArcFaceTemplate {
			px := int(math.Round(p[0]))
			py := int(math.Round(p[1]))
			c := out.RGBAAt(px, py)
			assert.Greater(t, int(c.R), 200, "landmark %s at %d,%d", LandmarkNames[i], px, py)
		}
	})
	t.Run("BackgroundPreserved", func(t *testing.T) {
		out, err := AlignedCrop(img, f, ArcFaceTemplateSize, ArcFaceTemplateSize)
		require.NoError(t, err)
		assert.Less(t, int(out.RGBAAt(5, 5).R), 100)
	})
	t.Run("CustomSize", func(t *testing.T) {
		out, err := AlignedCrop(img, f, 160, 160)
		require.NoError(t, err)
		assert.Equal(t, image.Rect(0, 0, 160, 160), out.Bounds())

		scaled := ScaledArcFaceTemplate(160, 160)
		c := out.RGBAAt(int(math.Round(scaled[2][0])), int(math.Round(scaled[2][1])))
		assert.Greater(t, int(c.R), 200)
	})
	t.Run("MissingLandmarks", func(t *testing.T) {
		_, err := AlignedCrop(img, &Face{}, 112, 112)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "landmarks")
	})
	t.Run("MissingImage", func(t *testing.T) {
		_, err := AlignedCrop(nil, f, 112, 112)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "image")
	})
	t.Run("InvalidSize", func(t *testing.T) {
		_, err := AlignedCrop(img, f, 0, 112)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "crop size")
	})
}

func TestSampleBilinear(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	img.SetRGBA(0, 0, color.RGBA{R: 0, A: 255})
	img.SetRGBA(1, 0, color.RGBA{R: 100, A: 255})
	img.SetRGBA(0, 1, color.RGBA{R: 0, A: 255})
	img.SetRGBA(1, 1, color.RGBA{R: 100, A: 255})
	bounds := img.Bounds()

	t.Run("ExactPixel", func(t *testing.T) {
		assert.Equal(t, uint8(100), sampleBilinear(img, bounds, 1, 0).R)
	})
	t.Run("Midpoint", func(t *testing.T) {
		assert.Equal(t, uint8(50), sampleBilinear(img, bounds, 0.5, 0).R)
	})
	t.Run("OutsideBounds", func(t *testing.T) {
		assert.Equal(t, color.RGBA{}, sampleBilinear(img, bounds, -5, -5))
	})
	t.Run("PartiallyOutside", func(t *testing.T) {
		c := sampleBilinear(img, bounds, -0.5, 0)
		assert.Equal(t, uint8(0), c.R)
		assert.Equal(t, uint8(128), c.A)
	})
}

func TestClampUint8(t *testing.T) {
	t.Run("Negative", func(t *testing.T) {
		assert.Equal(t, uint8(0), clampUint8(-1))
	})
	t.Run("Rounded", func(t *testing.T) {
		assert.Equal(t, uint8(13), clampUint8(12.6))
	})
	t.Run("Overflow", func(t *testing.T) {
		assert.Equal(t, uint8(255), clampUint8(300))
	})
}

func TestSimilarityTransformResidual(t *testing.T) {
	tpl := ArcFaceTemplate

	t.Run("RotatedAndScaled", func(t *testing.T) {
		// A proper similarity must fit exactly, whatever the rotation and scale.
		var src [NumLandmarks][2]float64
		r := 25 * math.Pi / 180

		for i := range tpl {
			src[i][0] = 2*(tpl[i][0]*math.Cos(r)-tpl[i][1]*math.Sin(r)) + 30
			src[i][1] = 2*(tpl[i][0]*math.Sin(r)+tpl[i][1]*math.Cos(r)) + 30
		}

		tr, err := similarityTransform(src, tpl)
		require.NoError(t, err)
		assert.InDelta(t, 0, fitResidual(tr, src, tpl), 1e-6)
	})
	t.Run("ExtremeYawAccepted", func(t *testing.T) {
		// A profile face compresses one axis and must still align rather than fall back.
		var src [NumLandmarks][2]float64
		for i := range tpl {
			src[i][0] = tpl[i][0] * 0.3
			src[i][1] = tpl[i][1]
		}

		_, err := similarityTransform(src, tpl)
		assert.NoError(t, err)
	})
	t.Run("MirroredRejected", func(t *testing.T) {
		// A reflection cannot be expressed by a proper rotation, so it must be refused
		// instead of warping the crop to a 22 px residual.
		var src [NumLandmarks][2]float64
		for i := range tpl {
			src[i][0] = float64(ArcFaceTemplateSize) - tpl[i][0]
			src[i][1] = tpl[i][1]
		}

		_, err := similarityTransform(src, tpl)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "do not fit the template")
	})
}

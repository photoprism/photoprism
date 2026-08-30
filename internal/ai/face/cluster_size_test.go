package face

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestClusterSize pins the bar to the model's own input geometry, so "not invented by
// interpolation" keeps its meaning when the model changes.
func TestClusterSize(t *testing.T) {
	t.Run("AlignedModel", func(t *testing.T) {
		assert.Equal(t, ArcFaceTemplateSize, ClusterSize(ModelSFace))
		assert.Equal(t, ArcFaceTemplateSize, ClusterSize(ModelAuraFace))
	})
	t.Run("BoxAlignedModel", func(t *testing.T) {
		// FaceNet reads the rectangular crop rather than the landmark template, so its bar is the
		// larger 160: a shared 112 would admit a crop it has to stretch.
		assert.Equal(t, CropSize.Width, ClusterSize(ModelFaceNet))
		assert.Greater(t, ClusterSize(ModelFaceNet), ClusterSize(ModelSFace))
	})
	t.Run("Unknown", func(t *testing.T) {
		assert.Equal(t, ClusterSizeThresholdDefault, ClusterSize("nonexistent"))
		assert.Equal(t, ClusterSizeThresholdDefault, ClusterSize(ModelNone))
	})
}

// TestFace_SetThumbSize covers the extent recorded beside an embedding, which is measured in the
// image it was sampled from rather than in the one detection ran on.
func TestFace_SetThumbSize(t *testing.T) {
	t.Run("ScalesOntoTheSource", func(t *testing.T) {
		f := &Face{Cols: 720, Area: Area{Scale: 60}}
		f.SetThumbSize(1800)

		assert.Equal(t, 150, f.ThumbSize)
	})
	t.Run("SameRendition", func(t *testing.T) {
		f := &Face{Cols: 720, Area: Area{Scale: 112}}
		f.SetThumbSize(720)

		assert.Equal(t, 112, f.ThumbSize, "a face at the bar fills the template from the detection thumbnail")
	})
	t.Run("UnknownSourceLeavesItUnset", func(t *testing.T) {
		f := &Face{Cols: 720, Area: Area{Scale: 60}}
		f.SetThumbSize(0)

		assert.Zero(t, f.ThumbSize, "a guess is worse than no value, which falls back to the detection size")
	})
	t.Run("UnrecordedDetectionWidthLeavesItUnset", func(t *testing.T) {
		// CropArea rewrites an unset Cols to 1 before the embedding loop runs, and ImageScale
		// then returns the source width itself. Recording that would be Size x srcWidth, which
		// clears any bar. No detection image is one pixel wide, so refuse instead.
		f := &Face{Cols: 1, Area: Area{Scale: 60}}
		f.SetThumbSize(1920)

		assert.Zero(t, f.ThumbSize)
	})
	t.Run("Nil", func(t *testing.T) {
		assert.NotPanics(t, func() { (*Face)(nil).SetThumbSize(1800) })
	})
}

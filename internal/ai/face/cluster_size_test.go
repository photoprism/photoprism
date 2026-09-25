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

// TestEmbedDetail covers how much of the crop the embedder asked for a source could supply, which
// is what tells a vector drawn from real pixels from one interpolated up to the same size.
func TestEmbedDetail(t *testing.T) {
	t.Run("Upscaled", func(t *testing.T) {
		assert.Equal(t, 46, EmbedDetail(52, 112), "a 52 px face stretched onto a 112 px template")
		assert.Equal(t, 30, EmbedDetail(48, 160))
		assert.Equal(t, 95, EmbedDetail(106, 112), "close is still short")
	})
	t.Run("FullDetail", func(t *testing.T) {
		assert.Equal(t, 100, EmbedDetail(112, 112))
	})
	t.Run("ClampedAtFullDetail", func(t *testing.T) {
		// Headroom above the crop is spent on the resample, so it is not detail the model sees.
		assert.Equal(t, 100, EmbedDetail(1600, 112))
	})
	t.Run("NeverZeroWhenMeasured", func(t *testing.T) {
		// Zero is the "could not be measured" state, so a real ratio that rounds to it is 1.
		assert.Equal(t, 1, EmbedDetail(1, 4096))
	})
	t.Run("Unmeasurable", func(t *testing.T) {
		assert.Zero(t, EmbedDetail(0, 112))
		assert.Zero(t, EmbedDetail(52, 0))
		assert.Zero(t, EmbedDetail(-1, 112))
	})
}

// TestFace_SetEmbedDetail covers the detail recorded beside an embedding, which is measured at
// the crop rather than derived afterwards from a template that may since have changed.
func TestFace_SetEmbedDetail(t *testing.T) {
	t.Run("AlignedTemplate", func(t *testing.T) {
		f := &Face{Cols: 720, Area: Area{Scale: 60}}
		f.SetThumbSize(720)
		f.SetEmbedDetail(ArcFaceTemplateSize)

		assert.Equal(t, 60, f.ThumbSize)
		assert.Equal(t, 54, f.EmbedDetail)
	})
	t.Run("WiderRenditionRemovesTheUpscale", func(t *testing.T) {
		// The same face embedded from a 1920 px rendition instead of the detection thumbnail.
		f := &Face{Cols: 720, Area: Area{Scale: 60}}
		f.SetThumbSize(1920)
		f.SetEmbedDetail(ArcFaceTemplateSize)

		assert.Equal(t, 100, f.EmbedDetail)
	})
	t.Run("UnsampledLeavesItUnset", func(t *testing.T) {
		f := &Face{Cols: 720, Area: Area{Scale: 60}}
		f.SetEmbedDetail(ArcFaceTemplateSize)

		assert.Zero(t, f.EmbedDetail, "no extent was recorded to measure against")
	})
	t.Run("Nil", func(t *testing.T) {
		assert.NotPanics(t, func() { (*Face)(nil).SetEmbedDetail(112) })
	})
}

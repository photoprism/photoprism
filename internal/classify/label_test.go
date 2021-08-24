package classify

import (
	"github.com/photoprism/photoprism/internal/face"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLabel_NewLocationLabel(t *testing.T) {
	LocLabel := LocationLabel("locationtest", 23)
	t.Log(LocLabel)
	assert.Equal(t, "location", LocLabel.Source)
	assert.Equal(t, 23, LocLabel.Uncertainty)
	assert.Equal(t, "locationtest", LocLabel.Name)

	t.Run("locationtest / slash", func(t *testing.T) {
		LocLabel := LocationLabel("locationtest / slash", 24)
		t.Log(LocLabel)
		assert.Equal(t, "location", LocLabel.Source)
		assert.Equal(t, 24, LocLabel.Uncertainty)
		assert.Equal(t, "locationtest", LocLabel.Name)
	})

	t.Run("locationtest - minus", func(t *testing.T) {
		LocLabel := LocationLabel("locationtest - minus", 80)
		t.Log(LocLabel)
		assert.Equal(t, "location", LocLabel.Source)
		assert.Equal(t, 80, LocLabel.Uncertainty)
		assert.Equal(t, "locationtest", LocLabel.Name)
	})

	t.Run("label as name", func(t *testing.T) {
		LocLabel := LocationLabel("barracouta", 80)
		t.Log(LocLabel)
		assert.Equal(t, "location", LocLabel.Source)
		assert.Equal(t, 80, LocLabel.Uncertainty)
		assert.Equal(t, "barracouta", LocLabel.Name)
		assert.Equal(t, "water", LocLabel.Categories[0])
		assert.Equal(t, 0, LocLabel.Priority)
	})
}

func TestLabel_Title(t *testing.T) {
	t.Run("locationtest123", func(t *testing.T) {
		LocLabel := LocationLabel("locationtest123", 23)
		assert.Equal(t, "Locationtest123", LocLabel.Title())
	})

	t.Run("Berlin/Neukölln", func(t *testing.T) {
		LocLabel := LocationLabel("berlin/neukölln_hasenheide", 23)
		assert.Equal(t, "Berlin / Neukölln Hasenheide", LocLabel.Title())
	})
}

func TestFaceLabels(t *testing.T) {
	Face1 := face.Face{
		Rows:       0,
		Cols:       0,
		Score:      0,
		Face:       face.Point{},
		Eyes:       nil,
		Landmarks:  nil,
		Embeddings: nil,
	}
	Face2 := face.Face{
		Rows:       0,
		Cols:       0,
		Score:      0,
		Face:       face.Point{},
		Eyes:       nil,
		Landmarks:  nil,
		Embeddings: nil,
	}
	t.Run("count < 1", func(t *testing.T) {
		Faces := face.Faces{}
		FaceLabels := FaceLabels(Faces, "")
		t.Log(FaceLabels)
		assert.Equal(t, 0, FaceLabels.Len())
	})
	t.Run("count > 1", func(t *testing.T) {
		Faces := face.Faces{Face1, Face2}
		FaceLabels := FaceLabels(Faces, "")
		t.Log(FaceLabels)
		assert.Equal(t, "people", FaceLabels[0].Name)
		assert.Equal(t, "", FaceLabels[0].Source)
		assert.Equal(t, 50, FaceLabels[0].Uncertainty)
		assert.Equal(t, 0, FaceLabels[0].Priority)
		//assert.Equal(t, "", FaceLabels[0].Categories)
	})
	t.Run("count = 1", func(t *testing.T) {
		Faces := face.Faces{Face1}
		FaceLabels := FaceLabels(Faces, "test")
		t.Log(FaceLabels)
		assert.Equal(t, "portrait", FaceLabels[0].Name)
		assert.Equal(t, "test", FaceLabels[0].Source)
		assert.Equal(t, 50, FaceLabels[0].Uncertainty)
		assert.Equal(t, 0, FaceLabels[0].Priority)
		assert.Equal(t, "people", FaceLabels[0].Categories[0])
	})
}

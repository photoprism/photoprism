package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/ai/face"
	"github.com/photoprism/photoprism/internal/thumb/crop"
	"github.com/photoprism/photoprism/pkg/rnd"
)

func TestMarkers_Overlapping(t *testing.T) {
	ValidateFixtures(t)
	file := File{FileHash: "a6c46e43b83fc02309b1c49e1ed7273f1f414610"}
	// The existing marker (cropArea2) fully covers the smaller XMP probe (cropArea1).
	existing := *NewMarker(file, cropArea2, "ls6sg6b1wowuy1c1", SrcImage, MarkerFace, 100, 65)
	probe := *NewMarker(file, cropArea1, "", SrcXmp, MarkerFace, 50, 50)

	t.Run("Found", func(t *testing.T) {
		markers := Markers{existing}
		got := markers.Overlapping(probe)
		if assert.NotNil(t, got) {
			assert.Equal(t, "ls6sg6b1wowuy1c1", got.SubjUID)
		}
	})
	t.Run("NoOverlap", func(t *testing.T) {
		m3 := *NewMarker(file, cropArea3, "ls6sg6b1wowuy1c3", SrcImage, MarkerFace, 100, 65)
		markers := Markers{m3}
		assert.Nil(t, markers.Overlapping(probe))
	})
	t.Run("SkipsInvalid", func(t *testing.T) {
		invalid := existing
		invalid.MarkerInvalid = true
		markers := Markers{invalid}
		assert.Nil(t, markers.Overlapping(probe))
	})
}

func TestMarkers_OverlapsInvalid(t *testing.T) {
	ValidateFixtures(t)
	file := File{FileHash: "a6c46e43b83fc02309b1c49e1ed7273f1f414610"}
	probe := *NewMarker(file, cropArea1, "", SrcXmp, MarkerFace, 50, 50)

	t.Run("TrueWhenRejectedOverlaps", func(t *testing.T) {
		invalid := *NewMarker(file, cropArea2, "ls6sg6b1wowuy1c1", SrcImage, MarkerFace, 100, 65)
		invalid.MarkerInvalid = true
		markers := Markers{invalid}
		assert.True(t, markers.OverlapsInvalid(probe))
	})
	t.Run("FalseWhenValidOverlaps", func(t *testing.T) {
		valid := *NewMarker(file, cropArea2, "ls6sg6b1wowuy1c1", SrcImage, MarkerFace, 100, 65)
		markers := Markers{valid}
		assert.False(t, markers.OverlapsInvalid(probe))
	})
	t.Run("FalseWhenNoOverlap", func(t *testing.T) {
		invalid := *NewMarker(file, cropArea3, "ls6sg6b1wowuy1c3", SrcImage, MarkerFace, 100, 65)
		invalid.MarkerInvalid = true
		markers := Markers{invalid}
		assert.False(t, markers.OverlapsInvalid(probe))
	})
}

func TestSubjSrcSharesFace(t *testing.T) {
	ValidateFixtures(t)
	assert.False(t, subjSrcSharesFace(SrcAuto))
	assert.False(t, subjSrcSharesFace(SrcXmp))
	assert.True(t, subjSrcSharesFace(SrcManual))
	assert.True(t, subjSrcSharesFace(SrcImage))
	assert.True(t, subjSrcSharesFace(SrcMeta))
}

func TestMarker_SetSubjectLink(t *testing.T) {
	ValidateFixtures(t)
	t.Run("Link", func(t *testing.T) {
		m := &Marker{}
		subj := &Subject{SubjUID: "js6sg6b1wowuy3c5", SubjName: "Alice"}
		m.SetSubjectLink(subj)
		assert.Equal(t, "js6sg6b1wowuy3c5", m.SubjUID)
		assert.Same(t, subj, m.subject)
	})
	t.Run("Detach", func(t *testing.T) {
		m := &Marker{SubjUID: "js6sg6b1wowuy3c5"}
		m.SetSubjectLink(nil)
		assert.Equal(t, "", m.SubjUID)
		assert.Nil(t, m.subject)
	})
}

// ensure the crop import stays referenced if the shared areas ever move.
var _ = crop.Area{}

func TestFile_AddFace_UpgradesEmbeddinglessMarker(t *testing.T) {
	ValidateFixtures(t)
	photo := Photo{PhotoUID: rnd.GenerateUID('p'), PhotoName: "xmp-addface", PhotoType: MediaImage}
	require.NoError(t, photo.Save())
	t.Cleanup(func() {
		assert.NoError(t, UnscopedDb().Delete(&Details{}, "photo_id = ?", photo.ID).Error)
		assert.NoError(t, UnscopedDb().Delete(&photo).Error)
	})
	file := &File{
		PhotoID:     photo.ID,
		PhotoUID:    photo.PhotoUID,
		FileUID:     rnd.GenerateUID('f'),
		FileHash:    "adface00000000000000000000000000000000a1",
		FileName:    "xmp-addface/a1.jpg",
		FileRoot:    RootOriginals,
		FilePrimary: true,
		FileType:    "jpg",
	}
	require.NoError(t, file.Create())
	t.Cleanup(func() { assert.NoError(t, UnscopedDb().Delete(file).Error) })

	// Persist an embedding-less XMP marker (as a prior pass would have).
	xmpMarker := NewMarker(*file, cropArea1, "", SrcXmp, MarkerFace, 100, 30)
	require.NotNil(t, xmpMarker)
	xmpMarker.MarkerName = "Alice"
	xmpMarker.SubjSrc = SrcXmp
	require.NoError(t, xmpMarker.Create())
	t.Cleanup(func() { assert.NoError(t, UnscopedDb().Delete(xmpMarker).Error) })
	require.Empty(t, xmpMarker.EmbeddingsJSON)

	// A later detection pass finds a real face overlapping the XMP marker. The detector
	// records which model produced the vector, which is what the upgraded row must store.
	f := face.Face{
		Rows: 1000, Cols: 1000, Score: 100,
		Area:        face.Area{Name: "face", Row: 385, Col: 486, Scale: 356},
		DetectModel: face.EngineONNX,
		EmbedModel:  face.ModelFaceNet,
		Embeddings:  face.Embeddings{testEmbeddings[0]},
	}

	restoreModel := face.ConfiguredModel()

	t.Cleanup(func() {
		_ = face.ConfigureEmbedder(face.EmbedderSettings{Name: restoreModel, Model: face.FindEmbeddingModel(restoreModel)})
	})

	require.NoError(t, face.ConfigureEmbedder(face.EmbedderSettings{
		Name:  face.ModelFaceNet,
		Model: face.FindEmbeddingModel(face.ModelFaceNet),
	}))

	file.markers = nil // force reload from DB
	file.AddFace(f, "")

	saved, err := FindMarkers(file.FileUID)
	require.NoError(t, err)
	require.Len(t, saved, 1, "must upgrade in place, not create a duplicate")
	assert.NotEmpty(t, saved[0].EmbeddingsJSON, "embedding-less XMP marker must gain the detected embedding")
	assert.Equal(t, face.ModelFaceNet, saved[0].EmbedModel, "the upgraded row must record the model that produced the vector")
	assert.Equal(t, face.EngineONNX, saved[0].DetectModel, "the upgraded row must record the detector that produced the crop")
	assert.Equal(t, "Alice", saved[0].MarkerName, "XMP name must be preserved")
}

func TestFile_AddFace_RecordsProducerModel(t *testing.T) {
	ValidateFixtures(t)
	photo := Photo{PhotoUID: rnd.GenerateUID('p'), PhotoName: "xmp-addface3", PhotoType: MediaImage}
	require.NoError(t, photo.Save())
	t.Cleanup(func() {
		assert.NoError(t, UnscopedDb().Delete(&Details{}, "photo_id = ?", photo.ID).Error)
		assert.NoError(t, UnscopedDb().Delete(&photo).Error)
	})
	file := &File{
		PhotoID:     photo.ID,
		PhotoUID:    photo.PhotoUID,
		FileUID:     rnd.GenerateUID('f'),
		FileHash:    "adface00000000000000000000000000000000a3",
		FileName:    "xmp-addface3/a3.jpg",
		FileRoot:    RootOriginals,
		FilePrimary: true,
		FileType:    "jpg",
	}
	require.NoError(t, file.Create())
	t.Cleanup(func() { assert.NoError(t, UnscopedDb().Delete(file).Error) })

	restoreModel := face.ConfiguredModel()

	t.Cleanup(func() {
		_ = face.ConfigureEmbedder(face.EmbedderSettings{Name: restoreModel, Model: face.FindEmbeddingModel(restoreModel)})
	})

	// The configured model deliberately differs from the one that produced the vector.
	require.NoError(t, face.ConfigureEmbedder(face.EmbedderSettings{
		Name:  face.ModelFaceNet,
		Model: face.FindEmbeddingModel(face.ModelFaceNet),
	}))

	f := face.Face{
		Rows: 1000, Cols: 1000, Score: 100,
		Area:        face.Area{Name: "face", Row: 385, Col: 486, Scale: 356},
		DetectModel: face.EngineONNX,
		EmbedModel:  face.ModelArcFaceR50,
		Embeddings:  face.Embeddings{testEmbeddings[0]},
	}

	file.AddFace(f, "")

	added := file.Markers()
	require.Len(t, *added, 1)
	assert.Equal(t, face.ModelArcFaceR50, (*added)[0].EmbedModel, "provenance must come from the producer, not the configuration")
	assert.Equal(t, face.EngineONNX, (*added)[0].DetectModel, "the detector that produced the crop is recorded beside the embedding model")
}

func TestFile_AddFace_DoesNotResurrectRejected(t *testing.T) {
	ValidateFixtures(t)
	photo := Photo{PhotoUID: rnd.GenerateUID('p'), PhotoName: "xmp-addface2", PhotoType: MediaImage}
	require.NoError(t, photo.Save())
	t.Cleanup(func() {
		assert.NoError(t, UnscopedDb().Delete(&Details{}, "photo_id = ?", photo.ID).Error)
		assert.NoError(t, UnscopedDb().Delete(&photo).Error)
	})
	file := &File{
		PhotoID:     photo.ID,
		PhotoUID:    photo.PhotoUID,
		FileUID:     rnd.GenerateUID('f'),
		FileHash:    "adface00000000000000000000000000000000b2",
		FileName:    "xmp-addface/b2.jpg",
		FileRoot:    RootOriginals,
		FilePrimary: true,
		FileType:    "jpg",
	}
	require.NoError(t, file.Create())
	t.Cleanup(func() { assert.NoError(t, UnscopedDb().Delete(file).Error) })

	rejected := NewMarker(*file, cropArea1, "", SrcImage, MarkerFace, 100, 30)
	require.NotNil(t, rejected)
	rejected.MarkerInvalid = true
	require.NoError(t, rejected.Create())
	t.Cleanup(func() { assert.NoError(t, UnscopedDb().Delete(rejected).Error) })

	f := face.Face{
		Rows: 1000, Cols: 1000, Score: 100,
		Area:       face.Area{Name: "face", Row: 385, Col: 486, Scale: 356},
		Embeddings: face.Embeddings{testEmbeddings[0]},
	}

	file.markers = nil
	file.AddFace(f, "")

	saved, err := FindMarkers(file.FileUID)
	require.NoError(t, err)
	require.Len(t, saved, 1, "a detected face over a rejected marker must not add a new marker")
	assert.True(t, saved[0].MarkerInvalid, "rejected marker stays rejected")
	assert.Empty(t, saved[0].EmbeddingsJSON, "rejected marker must not be upgraded")
}

func TestMarker_SetFace_XmpNotShared(t *testing.T) {
	ValidateFixtures(t)
	// SetFace propagates a marker's subject onto the shared Face for clustering name sources such
	// as SrcManual, but must not for SrcXmp: an imported XMP name labels only its own marker.
	// SetSubjectUID mutates the passed Face in memory, so an unchanged SubjUID proves it was gated.
	setup := func(t *testing.T, subjSrc, hash, person string) (*Marker, *Face, string) {
		photo := Photo{PhotoUID: rnd.GenerateUID('p'), PhotoName: "xmp-setface-" + subjSrc, PhotoType: MediaImage}
		require.NoError(t, photo.Save())
		t.Cleanup(func() {
			assert.NoError(t, UnscopedDb().Delete(&Details{}, "photo_id = ?", photo.ID).Error)
			assert.NoError(t, UnscopedDb().Delete(&photo).Error)
		})
		file := File{
			PhotoID:     photo.ID,
			PhotoUID:    photo.PhotoUID,
			FileUID:     rnd.GenerateUID('f'),
			FileHash:    hash,
			FileName:    "xmp-setface/" + subjSrc + ".jpg",
			FileRoot:    RootOriginals,
			FilePrimary: true,
			FileType:    "jpg",
		}
		require.NoError(t, file.Create())
		t.Cleanup(func() { assert.NoError(t, UnscopedDb().Delete(file).Error) })

		subj := FirstOrCreateSubject(NewSubject(person, SubjPerson, SrcManual))
		require.NotNil(t, subj)
		t.Cleanup(func() { assert.NoError(t, UnscopedDb().Delete(subj).Error) })

		// Model a detected AI marker (MarkerSrc = SrcImage) that has since gained a
		// name from the given source, so SetFace exercises the box-vs-name split.
		m := NewMarker(file, cropArea1, subj.SubjUID, SrcImage, MarkerFace, 100, 100)
		require.NotNil(t, m)
		m.SubjSrc = subjSrc
		m.SetEmbeddings(face.Embeddings{testEmbeddings[0]}, face.EmbeddingModelName(), face.EngineONNX)
		require.NoError(t, m.Create())
		t.Cleanup(func() { assert.NoError(t, UnscopedDb().Delete(m).Error) })

		// A subjectless shared face to observe whether the marker's subject is
		// pushed onto it; a unique id keeps the manual-case DB write local.
		return m, &Face{ID: "XMPSETFACE" + rnd.GenerateUID('f'), SubjUID: ""}, subj.SubjUID
	}
	t.Run("XmpDoesNotPropagate", func(t *testing.T) {
		m, f, subjUID := setup(t, SrcXmp, "5eface00000000000000000000000000000000a1", "Xmp Setface Person")
		_, err := m.SetFace(f, 0.5)
		require.NoError(t, err)
		assert.Empty(t, f.SubjUID, "XMP name must not propagate onto the shared face")
		assert.Equal(t, subjUID, m.SubjUID, "marker keeps its own XMP subject")
		assert.Equal(t, SrcXmp, m.SubjSrc, "marker subject source stays SrcXmp")
	})
	t.Run("ManualDoesPropagate", func(t *testing.T) {
		m, f, subjUID := setup(t, SrcManual, "5eface00000000000000000000000000000000b2", "Manual Setface Person")
		_, err := m.SetFace(f, 0.5)
		require.NoError(t, err)
		assert.Equal(t, subjUID, f.SubjUID, "manual name must propagate onto the shared face")
	})
}

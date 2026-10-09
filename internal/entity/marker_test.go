package entity

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/ai/face"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/thumb/crop"
	"github.com/photoprism/photoprism/pkg/rnd"
)

func TestMarker_SameEmbeddingModel(t *testing.T) {
	restore := face.ConfiguredModel()
	t.Cleanup(func() {
		_ = face.ConfigureEmbedder(face.EmbedderSettings{Name: restore, Model: face.FindEmbeddingModel(restore)})
	})

	assert.NoError(t, face.ConfigureEmbedder(face.EmbedderSettings{Name: face.ModelFaceNet, Model: face.FindEmbeddingModel(face.ModelFaceNet)}))
	assert.True(t, (&Marker{EmbedModel: face.ModelFaceNet}).SameEmbeddingModel())
	assert.True(t, (&Marker{}).SameEmbeddingModel())
	assert.False(t, (&Marker{EmbedModel: face.ModelSFace}).SameEmbeddingModel())

	assert.NoError(t, face.ConfigureEmbedder(face.EmbedderSettings{Name: face.ModelSFace}))
	assert.False(t, (&Marker{}).SameEmbeddingModel())
}

var testArea = crop.Area{
	Name: "face",
	X:    0.308333,
	Y:    0.206944,
	W:    0.355556,
	H:    0.355556,
}

var invalidArea1 = crop.Area{
	Name: "face",
	X:    -1,
	Y:    0.206944,
	W:    0.355556,
	H:    0.355556,
}

var invalidArea2 = crop.Area{
	Name: "face",
	X:    0.1,
	Y:    0.206944,
	W:    0,
	H:    0.355556,
}

var invalidArea3 = crop.Area{
	Name: "face",
	X:    0.1,
	Y:    -0.206944,
	W:    0.1,
	H:    0.355556,
}

func TestMarker_TableName(t *testing.T) {
	m := &Marker{}
	assert.Contains(t, m.TableName(), "markers")
}

func TestNewMarker(t *testing.T) {
	m := NewMarker(FileFixtures.Get("exampleFileName.jpg"), testArea, "ls6sg6b1wowuy3c3", SrcImage, MarkerLabel, 100, 29)
	assert.IsType(t, &Marker{}, m)
	assert.Equal(t, "fs6sg6bw45bnlqdw", m.FileUID)
	assert.Equal(t, "2cad9168fa6acc5c5c2965ddf6ec465ca42fd818-1340ce163163", m.Thumb)
	assert.Equal(t, "ls6sg6b1wowuy3c3", m.SubjUID)
	assert.Equal(t, 29, m.Score)
	assert.Equal(t, SrcImage, m.MarkerSrc)
	assert.Equal(t, MarkerLabel, m.MarkerType)
}

func TestMarkerSize(t *testing.T) {
	area := crop.NewArea("face", 0.4, 0.4, 0.1, 0.1)
	t.Run("Landscape", func(t *testing.T) {
		// Fit720 draws a 4:3 original at 720x540, so a tenth of the frame spans 72 px.
		assert.Equal(t, 72, MarkerSize(area, File{FileWidth: 4000, FileHeight: 3000}))
	})
	t.Run("Portrait", func(t *testing.T) {
		assert.Equal(t, 72, MarkerSize(area, File{FileWidth: 3000, FileHeight: 4000}))
	})
	t.Run("SmallOriginal", func(t *testing.T) {
		// A fit thumbnail never enlarges, so a small original is detected at its own size.
		assert.Equal(t, 64, MarkerSize(area, File{FileWidth: 640, FileHeight: 480}))
	})
	t.Run("UnknownDimensions", func(t *testing.T) {
		assert.Equal(t, -1, MarkerSize(area, File{}))
	})
	t.Run("SubPixelArea", func(t *testing.T) {
		// Never 0: GORM omits it on insert, so the row would read back as -1 and a second pass
		// would see a change that did not happen.
		tiny := crop.NewArea("face", 0.4, 0.4, 0.0001, 0.0001)
		assert.Equal(t, 1, MarkerSize(tiny, File{FileWidth: 640, FileHeight: 480}))
	})
}

// TestNewMarkerReview pins what "needs review" means on the score scale: a marker that exists but
// cannot contribute to a cluster is one a person has to look at. Stated against the threshold
// rather than a literal, because a literal is what let this drift onto the wrong scale before.
func TestNewMarkerReview(t *testing.T) {
	file := FileFixtures.Get("exampleFileName.jpg")

	// The shared default rather than the configurable variable, which is what NewMarker reads:
	// the review flag is stored, so it cannot follow a threshold an operator changes later.
	below := NewMarker(file, testArea, "ls6sg6b1wowuy3c3", SrcImage, MarkerFace, 100, face.ClusterScoreThresholdDefault-1)
	require.NotNil(t, below)
	assert.True(t, below.MarkerReview, "a marker under the clustering bar needs review")

	atBar := NewMarker(file, testArea, "ls6sg6b1wowuy3c3", SrcImage, MarkerFace, 100, face.ClusterScoreThresholdDefault)
	require.NotNil(t, atBar)
	assert.False(t, atBar.MarkerReview, "a marker scored above the bar does not")
}

func TestMarker_SetName(t *testing.T) {
	t.Run("InvalidName", func(t *testing.T) {
		m := MarkerFixtures.Get("actress-a-1")
		assert.IsType(t, Marker{}, m)
		assert.Equal(t, "Actress A", m.MarkerName)
		changed, err := m.SetName("", SrcManual)

		if err != nil {
			t.Fatal(err)
		}

		assert.False(t, changed)
		assert.Equal(t, "Actress A", m.MarkerName)

		changed, err = m.SetName("Foo Bar", SrcAuto)

		if err != nil {
			t.Fatal(err)
		}

		assert.False(t, changed)
		assert.Equal(t, "Actress A", m.MarkerName)
	})
}

func TestMarker_SaveForm(t *testing.T) {
	t.Run("FaGeAddNewNameToMarkerThenRenameMarker", func(t *testing.T) {
		m := MarkerFixtures.Get("fa-gr-1")
		m2 := MarkerFixtures.Get("fa-gr-2")
		m3 := MarkerFixtures.Get("fa-gr-3")

		assert.Empty(t, m.SubjUID)
		assert.Empty(t, m2.SubjUID)
		assert.Empty(t, m3.SubjUID)

		m.MarkerInvalid = true
		m.Score = 50

		//set new name

		f := form.Marker{SubjSrc: SrcManual, MarkerName: "Jane Doe", MarkerInvalid: false}

		changed, err := m.SaveForm(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.True(t, changed)
		assert.NotEmpty(t, m.SubjUID)

		if s := m.Subject(); s != nil {
			assert.Equal(t, "Jane Doe", s.SubjName)
		}
		if m := FindMarker("ms6sg6b1wowuy777"); m != nil {
			assert.Equal(t, "Jane Doe", m.Subject().SubjName)
		}
		if m := FindMarker("ms6sg6b1wowuy888"); m != nil {
			assert.Equal(t, "Jane Doe", m.Subject().SubjName)
		}

		// Rename subject.
		f3 := form.Marker{SubjSrc: SrcManual, MarkerName: "Franzilein", MarkerInvalid: false}

		if m := FindMarker("ms6sg6b1wowuy777"); m == nil {
			t.Fatal("result is nil")
		} else if changed, err := m.SaveForm(f3); err != nil {
			t.Fatal(err)
		} else {
			assert.True(t, changed)
		}

		if m := FindMarker("ms6sg6b1wowuy666"); m != nil {
			assert.Equal(t, "Franzilein", m.Subject().SubjName)
		}
		if m := FindMarker("ms6sg6b1wowuy777"); m != nil {
			assert.Equal(t, "Franzilein", m.Subject().SubjName)
		}
		if m := FindMarker("ms6sg6b1wowuy888"); m != nil {
			assert.Equal(t, "Franzilein", m.Subject().SubjName)
		}
	})
}

func TestUpdateOrCreateMarker(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		m := NewMarker(FileFixtures.Get("exampleFileName.jpg"), testArea, "ls6sg6b1wowuy3c3", SrcImage, MarkerLabel, 100, 65)
		assert.IsType(t, &Marker{}, m)
		assert.Equal(t, "fs6sg6bw45bnlqdw", m.FileUID)
		assert.Equal(t, "ls6sg6b1wowuy3c3", m.SubjUID)
		assert.Equal(t, SrcImage, m.MarkerSrc)
		assert.Equal(t, MarkerLabel, m.MarkerType)

		m, err := CreateMarkerIfNotExists(m)

		if err != nil {
			t.Fatal(err)
		}

		if m == nil {
			t.Fatal("result must not be nil")
		}

		if m.MarkerUID == "" || m.FileUID == "" {
			t.Errorf("UIDs should not be empty")
		}
	})
}

func TestMarker_Delete(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		m := NewMarker(FileFixtures.Get("exampleFileName.jpg"), crop.Area{Name: "face", X: 0.01, Y: 0.01, W: 0.02, H: 0.02}, "", SrcXmp, MarkerFace, 100, 30)
		if err := m.Create(); err != nil {
			t.Fatal(err)
		}
		if m.MarkerUID == "" || FindMarker(m.MarkerUID) == nil {
			t.Fatal("created marker not found")
		}
		if err := m.Delete(); err != nil {
			t.Fatal(err)
		}
		if found := FindMarker(m.MarkerUID); found != nil {
			t.Errorf("deleted marker still exists: %s", found.MarkerUID)
		}
	})
	t.Run("EmptyUID", func(t *testing.T) {
		if err := (&Marker{}).Delete(); err == nil {
			t.Error("deleting a marker with an empty UID must return an error, not issue an unscoped delete")
		}
	})
}

func TestMarker_Updates(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		m := NewMarker(FileFixtures.Get("exampleFileName.jpg"), testArea, "ls6sg6b1wowuy3c4", SrcImage, MarkerLabel, 100, 65)
		m, err := CreateMarkerIfNotExists(m)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, SrcImage, m.MarkerSrc)
		assert.Equal(t, MarkerLabel, m.MarkerType)

		if err = m.Updates(Marker{MarkerSrc: SrcMeta}); err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, SrcMeta, m.MarkerSrc)
		assert.Equal(t, MarkerLabel, m.MarkerType)

		if m.MarkerUID == "" || m.FileUID == "" {
			t.Errorf("UIDs should not be empty")
		}
	})
}

func TestMarker_Update(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		m := NewMarker(FileFixtures.Get("exampleFileName.jpg"), testArea, "ls6sg6b1wowuy3c4", SrcImage, MarkerLabel, 100, 65)
		m, err := CreateMarkerIfNotExists(m)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, MarkerLabel, m.MarkerType)

		if err := m.Update("MarkerSrc", SrcMeta); err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, SrcMeta, m.MarkerSrc)
		assert.Equal(t, MarkerLabel, m.MarkerType)

		if m.MarkerUID == "" || m.FileUID == "" {
			t.Errorf("UIDs should not be empty")
		}
	})
}

func TestMarker_InvalidArea(t *testing.T) {
	t.Run("TestArea", func(t *testing.T) {
		m := NewMarker(FileFixtures.Get("exampleFileName.jpg"), testArea, "ls6sg6b1wowuy3c4", SrcImage, MarkerFace, 100, 65)
		assert.Nil(t, m.InvalidArea())
		m.MarkerType = MarkerUnknown
		assert.Nil(t, m.InvalidArea())
	})
	t.Run("InvalidArea1", func(t *testing.T) {
		m := NewMarker(FileFixtures.Get("exampleFileName.jpg"), invalidArea1, "ls6sg6b1wowuy3c4", SrcImage, MarkerFace, 100, 65)
		assert.EqualError(t, m.InvalidArea(), "invalid face crop area x=-100% y=20% w=35% h=35%")
		m.MarkerUID = "m345634636"
		assert.EqualError(t, m.InvalidArea(), "invalid face crop area x=-100% y=20% w=35% h=35%")
		m.MarkerType = MarkerUnknown
		assert.Nil(t, m.InvalidArea())
	})
	t.Run("InvalidArea2", func(t *testing.T) {
		m := NewMarker(FileFixtures.Get("exampleFileName.jpg"), invalidArea2, "ls6sg6b1wowuy3c4", SrcImage, MarkerFace, 100, 65)
		assert.Error(t, m.InvalidArea())
		m.MarkerType = MarkerUnknown
		assert.Nil(t, m.InvalidArea())
	})
	t.Run("InvalidArea3", func(t *testing.T) {
		m := NewMarker(FileFixtures.Get("exampleFileName.jpg"), invalidArea3, "ls6sg6b1wowuy3c4", SrcImage, MarkerFace, 100, 65)
		assert.Error(t, m.InvalidArea())
		m.MarkerType = MarkerUnknown
		assert.Nil(t, m.InvalidArea())
	})
}

// TODO fails on mariadb
func TestMarker_Save(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		m := NewMarker(FileFixtures.Get("exampleFileName.jpg"), testArea, "ls6sg6b1wowuy3c4", SrcImage, MarkerLabel, 100, 65)

		m, err := CreateMarkerIfNotExists(m)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, MarkerLabel, m.MarkerType)

		m.MarkerSrc = SrcMeta

		assert.Equal(t, SrcMeta, m.MarkerSrc)

		initialDate := m.UpdatedAt

		if err := m.Save(); err != nil {
			t.Fatal(err)
		}

		afterDate := m.UpdatedAt

		assert.Equal(t, SrcMeta, m.MarkerSrc)
		// Timestamps are stored with second precision, so a save within the same
		// second leaves UpdatedAt unchanged; assert it stays within a sane window.
		elapsed := afterDate.Sub(initialDate)
		assert.GreaterOrEqual(t, elapsed, time.Duration(0))
		assert.Less(t, elapsed, time.Minute)

		if m.MarkerUID == "" || m.FileUID == "" {
			t.Errorf("UIDs should not be empty")
		}

		p := PhotoFixtures.Get("19800101_000002_D640C559")
		assert.Empty(t, p.Files)
		p.PreloadFiles()
		assert.NotEmpty(t, p.Files)
	})
	t.Run("InvalidPosition", func(t *testing.T) {
		m := Marker{X: -1, Y: 0, W: 0.2, H: 0.133, MarkerType: MarkerFace}

		if err := m.Save(); err == nil {
			t.Fatal("error expected")
		} else {
			assert.Equal(t, "invalid face crop area x=-100% y=0% w=20% h=13%", err.Error())
		}

	})
}

func TestMarker_ClearSubject(t *testing.T) {
	t.Run("Num1000003Two", func(t *testing.T) {
		m := MarkerFixtures.Get("1000003-2")

		assert.NotEmpty(t, m.MarkerName)

		err := m.ClearSubject(SrcAuto)

		if err != nil {
			t.Fatal(err)
		}

		assert.Empty(t, m.MarkerName)
	})
	t.Run("ActorOne", func(t *testing.T) {
		m := MarkerFixtures.Get("actor-a-4")  // id 18
		m2 := MarkerFixtures.Get("actor-a-3") // id 17
		m3 := MarkerFixtures.Get("actor-a-2") // id 16
		m4 := MarkerFixtures.Get("actor-a-1") // id 15

		assert.Equal(t, "js6sg6b1h1njaaad", m.SubjUID)
		assert.Equal(t, "js6sg6b1h1njaaad", m2.SubjUID)
		assert.Equal(t, "js6sg6b1h1njaaad", m3.SubjUID)
		assert.Equal(t, "js6sg6b1h1njaaad", m4.SubjUID)
		assert.NotNil(t, m.Face())
		assert.NotNil(t, m2.Face())
		assert.NotNil(t, m3.Face())
		assert.NotNil(t, m4.Face())

		if m := FindMarker("ms6sg6b1wowu1002"); m == nil {
			t.Fatal("marker is nil")
		} else if f := m.Face(); f == nil {
			t.Fatal("face is nil")
		}

		assert.Equal(t, "PI6A2XGOTUXEFI7CBF4KCI5I2I3JEJHS", m.Face().ID)
		assert.Equal(t, "PI6A2XGOTUXEFI7CBF4KCI5I2I3JEJHS", m2.Face().ID)
		assert.Equal(t, "PI6A2XGOTUXEFI7CBF4KCI5I2I3JEJHS", m3.Face().ID)
		assert.Equal(t, "PI6A2XGOTUXEFI7CBF4KCI5I2I3JEJHS", m4.Face().ID)
		assert.Equal(t, int(0), FindMarker("ms6sg6b1wowu1002").Face().Collisions)

		// Reset face subject.
		err := m.ClearSubject(SrcAuto)

		if err != nil {
			t.Fatal(err)
		}

		assert.NotNil(t, FindMarker("ms6sg6b1wowu1004"))
		assert.NotNil(t, FindMarker("ms6sg6b1wowu1003"))
		assert.NotNil(t, FindMarker("ms6sg6b1wowu1002"))
		assert.NotNil(t, FindFace("PI6A2XGOTUXEFI7CBF4KCI5I2I3JEJHS"))

		assert.Empty(t, m.SubjUID)
		assert.Equal(t, "", FindMarker("ms6sg6b1wowu1004").SubjUID)
		assert.Equal(t, "", FindMarker("ms6sg6b1wowu1003").SubjUID)
		assert.Equal(t, "", FindMarker("ms6sg6b1wowu1002").SubjUID)
		assert.Empty(t, m.FaceID)
		assert.Equal(t, "", FindMarker("ms6sg6b1wowu1004").FaceID)
		assert.Equal(t, "", FindMarker("ms6sg6b1wowu1003").FaceID)
		assert.Equal(t, "", FindMarker("ms6sg6b1wowu1002").FaceID)
		assert.Equal(t, int(1), FindFace("PI6A2XGOTUXEFI7CBF4KCI5I2I3JEJHS").Collisions)
	})
}

func TestMarker_ClearFace(t *testing.T) {
	t.Run("Num1000003Two", func(t *testing.T) {
		m := MarkerFixtures.Get("1000003-2")

		assert.NotEmpty(t, m.FaceID)

		updated, err := m.ClearFace()

		if err != nil {
			t.Fatal(err)
		}

		assert.True(t, updated)
		assert.Empty(t, m.FaceID)
	})
	t.Run("EmptyFaceId", func(t *testing.T) {
		m := Marker{FaceID: "", MarkerUID: "IShouldntBeInDB"}

		updated, err := m.ClearFace()

		if err != nil {
			t.Fatal(err)
		}

		assert.False(t, updated)
		assert.Empty(t, m.FaceID)
	})

	t.Run("missing markeruid", func(t *testing.T) {
		m := Marker{FaceID: ""}

		updated, err := m.ClearFace()

		assert.ErrorContains(t, err, "markeruid required but not provided")
		assert.False(t, updated)
		assert.Empty(t, m.FaceID)
	})
	t.Run("SubjectSrcManual", func(t *testing.T) {
		m := Marker{MarkerUID: "mqyz9x61edicxf8j", FaceID: "123ab"}

		assert.NotEmpty(t, m.FaceID)
		assert.Empty(t, m.MatchedAt)
		updated, err := m.ClearFace()

		if err != nil {
			t.Fatal(err)
		}

		assert.True(t, updated)
		assert.Empty(t, m.FaceID)
		assert.NotEmpty(t, m.MatchedAt)
	})
	t.Run("ReturnsUpdateError", func(t *testing.T) {
		Db().AddError(errors.New("Force Gorm To Return Error"))
		t.Cleanup(func() {
			Db().Error = nil
		})

		m := Marker{
			FaceID:    "FACE-CLEAR-ERR-1",
			SubjSrc:   SrcAuto,
			MarkerUID: rnd.GenerateUID('m'),
		}

		updated, err := m.ClearFace()
		assert.True(t, updated)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Force Gorm To Return Error")

		Db().AddError(errors.New("Force Gorm To Return Error"))
		m = Marker{
			FaceID:    "FACE-CLEAR-ERR-2",
			SubjSrc:   SrcBatch,
			MarkerUID: rnd.GenerateUID('m'),
		}

		updated, err = m.ClearFace()
		assert.True(t, updated)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Force Gorm To Return Error")

	})
}

func TestMarker_SyncSubject(t *testing.T) {
	t.Run("NoFaceMarker", func(t *testing.T) {
		m := Marker{MarkerType: "test", subject: nil}
		assert.Nil(t, m.SyncSubject(false))
	})
	t.Run("SubjectIsNil", func(t *testing.T) {
		m := Marker{MarkerType: MarkerFace, subject: nil}
		assert.Nil(t, m.SyncSubject(false))
	})
	t.Run("UpdateKnownFaceError", func(t *testing.T) {
		Db().AddError(errors.New("Force Gorm To Return Error"))
		t.Cleanup(func() {
			Db().Error = nil
		})

		subjUID := "jsyncsubjecterror123"
		m := Marker{
			MarkerType: MarkerFace,
			FaceID:     "FACE-SYNC-ERR-1",
			SubjUID:    subjUID,
			SubjSrc:    SrcManual,
			subject: &Subject{
				SubjUID: subjUID,
			},
		}

		err := m.SyncSubject(false)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "update known face")
	})
}

func TestMarker_Create(t *testing.T) {
	t.Run("InvalidPosition", func(t *testing.T) {
		m := Marker{X: 0, Y: 0, MarkerType: MarkerFace}
		err := m.Create()
		if err == nil {
			t.Fatal("error expected")
		} else {
			assert.Equal(t, "invalid face crop area x=0% y=0% w=0% h=0%", err.Error())
		}
	})
}

func TestMarker_Embeddings(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// The fixtures are generated for whichever model a run resolves to, so what the
		// vector has to be is its width and its provenance, not a particular value.
		m := MarkerFixtures.Get("1000003-4")

		require.Len(t, m.Embeddings(), 1)
		assert.Len(t, m.Embeddings()[0], face.ExpectedDims())
		assert.True(t, m.SameEmbeddingModel())
	})
	t.Run("EmptyEmbedding", func(t *testing.T) {
		m := Marker{}
		m.EmbeddingsJSON = []byte("")

		assert.Empty(t, m.Embeddings())
	})
	t.Run("InvalidEmbeddingJson", func(t *testing.T) {
		m := Marker{}
		m.EmbeddingsJSON = []byte("[false]")

		assert.Empty(t, m.Embeddings()[0])
	})
}

func TestMarker_HasFace(t *testing.T) {
	t.Run("True", func(t *testing.T) {
		m := MarkerFixtures.Get("1000003-6")

		assert.True(t, m.HasFace(nil, -1))
		assert.True(t, m.HasFace(FaceFixtures.Pointer("joe-biden"), -1))
	})
	t.Run("False", func(t *testing.T) {
		m := MarkerFixtures.Get("1000003-6")

		assert.False(t, m.HasFace(FaceFixtures.Pointer("joe-biden"), 0.1))
	})
	t.Run("FaceIdEmpty", func(t *testing.T) {
		m := Marker{FaceID: ""}

		assert.False(t, m.HasFace(FaceFixtures.Pointer("joe-biden"), 0.1))
	})
	t.Run("FaceDistLessThanZero", func(t *testing.T) {
		m := Marker{FaceID: "123", FaceDist: -1}

		assert.False(t, m.HasFace(FaceFixtures.Pointer("joe-biden"), 0.1))
	})
	t.Run("FaceIdEqualFId", func(t *testing.T) {
		m := Marker{FaceID: "VF7ANLDET2BKZNT4VQWJMMC6HBEFDOG6"}

		assert.True(t, m.HasFace(FaceFixtures.Pointer("joe-biden"), 0.1))
	})
}

func TestMarker_Subject(t *testing.T) {
	t.Run("EmptySubjUID", func(t *testing.T) {
		m := Marker{SubjUID: "", subject: &Subject{SubjUID: "", SubjName: "Test Subject"}}

		if s := m.Subject(); s == nil {
			t.Fatal("return value must not be nil")
		} else {
			assert.Equal(t, "Test Subject", s.SubjName)
			assert.Equal(t, "", m.SubjUID)
			assert.Equal(t, "", s.SubjUID)
		}
	})
	t.Run("ConflictingSubjUID", func(t *testing.T) {
		m := Marker{SubjUID: "", subject: &Subject{SubjUID: "xyz", SubjName: "Test Subject"}}

		if s := m.Subject(); s != nil {
			t.Fatal("return value must be nil")
		}
	})
	t.Run("SubjSrcAuto", func(t *testing.T) {
		m := Marker{SubjSrc: SrcAuto, SubjUID: "", MarkerName: "Hans Mayer"}

		if s := m.Subject(); s != nil {
			t.Fatal("return value must be nil")
		} else {
			assert.Equal(t, "Hans Mayer", m.MarkerName)
			assert.Empty(t, m.SubjUID)
			assert.Equal(t, SrcAuto, m.SubjSrc)
		}
	})
	t.Run("SubjSrcManual", func(t *testing.T) {
		m := Marker{SubjSrc: SrcManual, SubjUID: "", MarkerName: "Hans Mayer"}

		if s := m.Subject(); s == nil {
			t.Fatal("return value must not be nil")
		} else {
			assert.Equal(t, "Hans Mayer", s.SubjName)
			assert.NotEmpty(t, s.SubjUID)
		}
	})
}

func TestMarker_GetFace(t *testing.T) {
	t.Run("ExistingFaceID", func(t *testing.T) {
		m := Marker{MarkerUID: "ms6sg6b14ahkyd24", FaceID: "1234", face: &Face{ID: "1234"}}

		if f := m.Face(); f == nil {
			t.Fatal("return value must not be nil")
		} else {
			assert.Equal(t, "1234", f.ID)
			assert.Equal(t, "1234", m.FaceID)
		}
	})
	t.Run("ConflictingFaceID", func(t *testing.T) {
		m := Marker{MarkerUID: "ms6sg6b14ahkyd24", FaceID: "8888", face: &Face{ID: "1234"}}

		if f := m.Face(); f != nil {
			t.Fatal("return value must be nil")
		} else {
			assert.Equal(t, "8888", m.FaceID)
			assert.Nil(t, m.face)
		}
	})
	t.Run("FindFaceWithId", func(t *testing.T) {
		m := Marker{MarkerUID: "ms6sg6b14ahkyd24", FaceID: "VF7ANLDET2BKZNT4VQWJMMC6HBEFDOG6"}

		if f := m.Face(); f == nil {
			t.Fatal("return value must not be nil")
		} else {
			assert.Equal(t, "VF7ANLDET2BKZNT4VQWJMMC6HBEFDOG6", f.ID)
		}
	})
	t.Run("LowQualityMarker", func(t *testing.T) {
		m := Marker{MarkerUID: "", FaceID: "", SubjSrc: SrcManual, Size: 130}

		assert.Nil(t, m.Face())
	})
	t.Run("CreateFace", func(t *testing.T) {
		m := Marker{
			MarkerUID:      "ms6sg6b14ahkyd24",
			FaceID:         "",
			EmbeddingsJSON: MarkerFixtures.Get("actress-a-1").EmbeddingsJSON,
			SubjSrc:        SrcManual,
			Size:           160,
			Score:          80,
		}

		if m.Face() == nil {
			t.Fatal("return value must not be nil")
		} else {
			assert.NotEmpty(t, m.Face().ID)
		}
	})
}

func TestFindMarker(t *testing.T) {
	t.Run("Nil", func(t *testing.T) {
		assert.Nil(t, FindMarker("0000"))
	})
}

// TestMarker_SetFace checks assignment inputs and persisted membership.
func TestMarker_SetFace(t *testing.T) {
	t.Run("FaceEqualNil", func(t *testing.T) {
		m := MarkerFixtures.Pointer("1000003-6")
		assert.Equal(t, "PN6QO5INYTUSAATOFL43LL2ABAV5ACZK", m.FaceID)
		updated, err := m.SetFace(nil, -1)
		assert.False(t, updated)
		assert.Equal(t, "PN6QO5INYTUSAATOFL43LL2ABAV5ACZK", m.FaceID)
		assert.Equal(t, fmt.Errorf("face is nil"), err)
	})
	t.Run("WrongMarkerType", func(t *testing.T) {
		m := Marker{MarkerType: "xxx"}
		updated, err := m.SetFace(&Face{ID: "99876"}, -1)
		assert.False(t, updated)
		assert.Equal(t, "", m.FaceID)
		assert.Equal(t, fmt.Errorf("not a face marker"), err)
	})
	t.Run("SkipSameFace", func(t *testing.T) {
		m := Marker{MarkerType: MarkerFace, SubjUID: "js6sg6b1qekk9jx8", FaceID: "99876uyt", X: 0.01, Y: 0.01, W: 0.01, H: 0.01}
		if err := m.Create(); err != nil {
			t.Error(err)
		}
		updated, err := m.SetFace(&Face{ID: "99876uyt", SubjUID: "js6sg6b1qekk9jx8"}, -1)
		assert.False(t, updated)
		assert.Equal(t, "99876uyt", m.FaceID)
		assert.Nil(t, err)
		assert.Nil(t, UnscopedDb().Delete(&m).Error)
	})
	t.Run("SetNewFace", func(t *testing.T) {
		m := Marker{MarkerUID: "mqyz9x61edicxf8j", MarkerType: MarkerFace, SubjUID: "", FaceID: ""}
		require.NoError(t, UnscopedDb().Create(&m).Error)
		t.Cleanup(func() { UnscopedDb().Delete(&Marker{}, "marker_uid = ?", m.MarkerUID) })

		updated, err := m.SetFace(FaceFixtures.Pointer("john-doe"), -1)
		assert.True(t, updated)
		assert.Equal(t, "PN6QO5INYTUSAATOFL43LL2ABAV5ACZK", m.FaceID)
		assert.Nil(t, err)
		updated2, err := m.ClearFace()

		if err != nil {
			t.Fatal(err)
		}

		assert.True(t, updated2)
		assert.Empty(t, m.FaceID)
	})
}

func TestMarker_RefreshPhotos(t *testing.T) {
	m := MarkerFixtures.Get("1000003-6")

	if err := m.RefreshPhotos(); err != nil {
		t.Fatal(err)
	}
}

func TestMarker_SurfaceRatio(t *testing.T) {
	m1 := *NewMarker(FileFixtures.Get("exampleFileName.jpg"), cropArea1, "ls6sg6b1wowuy1c1", SrcImage, MarkerFace, 100, 65)
	m2 := *NewMarker(FileFixtures.Get("exampleFileName.jpg"), cropArea2, "ls6sg6b1wowuy1c2", SrcImage, MarkerFace, 100, 65)
	m3 := *NewMarker(FileFixtures.Get("exampleFileName.jpg"), cropArea3, "ls6sg6b1wowuy1c3", SrcImage, MarkerFace, 100, 65)
	m4 := *NewMarker(FileFixtures.Get("exampleFileName.jpg"), cropArea4, "ls6sg6b1wowuy1c3", SrcImage, MarkerFace, 100, 65)

	assert.Equal(t, 99, int(m1.SurfaceRatio(m1.OverlapArea(m1))*100))
	assert.Equal(t, 99, int(m1.SurfaceRatio(m1.OverlapArea(m2))*100))
	assert.Equal(t, 29, int(m2.SurfaceRatio(m2.OverlapArea(m1))*100))
	assert.Equal(t, 0, int(m1.SurfaceRatio(m1.OverlapArea(m3))*100))
	assert.Equal(t, 30, int(m1.SurfaceRatio(m1.OverlapArea(m4))*100))
	assert.Equal(t, 0, int(m1.SurfaceRatio(m3.OverlapArea(m1))*100))
	assert.Equal(t, 30, int(m1.SurfaceRatio(m4.OverlapArea(m1))*100))
}

func TestMarker_OverlapArea(t *testing.T) {
	m1 := *NewMarker(FileFixtures.Get("exampleFileName.jpg"), cropArea1, "ls6sg6b1wowuy1c1", SrcImage, MarkerFace, 100, 65)
	m2 := *NewMarker(FileFixtures.Get("exampleFileName.jpg"), cropArea2, "ls6sg6b1wowuy1c2", SrcImage, MarkerFace, 100, 65)
	m3 := *NewMarker(FileFixtures.Get("exampleFileName.jpg"), cropArea3, "ls6sg6b1wowuy1c3", SrcImage, MarkerFace, 100, 65)
	m4 := *NewMarker(FileFixtures.Get("exampleFileName.jpg"), cropArea4, "ls6sg6b1wowuy1c3", SrcImage, MarkerFace, 100, 65)

	assert.Equal(t, 0.1264200823986168, m1.OverlapArea(m1))
	assert.Equal(t, int(m1.Surface()*10000), int(m1.OverlapArea(m1)*10000))
	assert.Equal(t, 0.1264200823986168, m1.OverlapArea(m2))
	assert.Equal(t, 0.1264200823986168, m2.OverlapArea(m1))
	assert.Equal(t, 0.0, m1.OverlapArea(m3))
	assert.Equal(t, 0.038166598943088825, m1.OverlapArea(m4))
}

func TestMarker_OverlapPercent(t *testing.T) {
	m1 := *NewMarker(FileFixtures.Get("exampleFileName.jpg"), cropArea1, "ls6sg6b1wowuy1c1", SrcImage, MarkerFace, 100, 65)
	m2 := *NewMarker(FileFixtures.Get("exampleFileName.jpg"), cropArea2, "ls6sg6b1wowuy1c2", SrcImage, MarkerFace, 100, 65)
	m3 := *NewMarker(FileFixtures.Get("exampleFileName.jpg"), cropArea3, "ls6sg6b1wowuy1c3", SrcImage, MarkerFace, 100, 65)
	m4 := *NewMarker(FileFixtures.Get("exampleFileName.jpg"), cropArea4, "ls6sg6b1wowuy1c3", SrcImage, MarkerFace, 100, 65)

	assert.Equal(t, 100, m1.OverlapPercent(m1))
	assert.Equal(t, 29, m1.OverlapPercent(m2))
	assert.Equal(t, 100, m2.OverlapPercent(m1))
	assert.Equal(t, 0, m1.OverlapPercent(m3))
	assert.Equal(t, 96, m1.OverlapPercent(m4))
}

func TestMarker_String(t *testing.T) {
	t.Run("Nil", func(t *testing.T) {
		var m *Marker
		assert.Equal(t, "Marker<nil>", m.String())
		//nolint:staticcheck // the point is that fmt reaches String(), which calling it cannot show.
		assert.Equal(t, "Marker<nil>", fmt.Sprintf("%s", m))
	})
	t.Run("New", func(t *testing.T) {
		m := &Marker{}
		assert.Equal(t, "*Marker", m.String())
		//nolint:staticcheck // the point is that fmt reaches String(), which calling it cannot show.
		assert.Equal(t, "*Marker", fmt.Sprintf("%s", m))
	})
	t.Run("Name", func(t *testing.T) {
		m := MarkerFixtures.Pointer("1000003-4")
		assert.Equal(t, "Jens Mander", m.String())
	})
}

func TestMarker_Matched(t *testing.T) {
	t.Run("missing markeruid", func(t *testing.T) {
		m := Marker{FileUID: "DummyValue"}
		if err := m.Matched(); err != nil {
			assert.Equal(t, "markeruid required but not provided", err.Error())

		}
	})
}

func TestMarker_SetEmbeddings(t *testing.T) {
	t.Run("RecordsTheProducingModel", func(t *testing.T) {
		// Provenance is what keeps two embedding spaces apart. A vector stored without it
		// reads as legacy FaceNet and would be admitted into FaceNet clusters whatever
		// model actually produced it.
		m := &Marker{MarkerType: MarkerFace}
		m.SetEmbeddings(face.Embeddings{face.RandomEmbedding()}, face.ModelSFace, face.EngineONNX)

		assert.Equal(t, face.ModelSFace, m.EmbedModel)
		assert.NotEmpty(t, m.EmbeddingsJSON)
		assert.False(t, m.Embeddings().Empty())
	})
	t.Run("RecordsTheProducingDetector", func(t *testing.T) {
		// The detector decides the landmarks and therefore the aligned crop, so a vector
		// whose detector is unknown cannot be told apart from one a legacy set produced.
		m := &Marker{MarkerType: MarkerFace}
		m.SetEmbeddings(face.Embeddings{face.RandomEmbedding()}, face.ModelSFace, face.EngineONNX)

		assert.Equal(t, face.EngineONNX, m.DetectModel)
	})
	t.Run("EmptyClearsTheModel", func(t *testing.T) {
		// A marker whose vector was cleared must not keep claiming a model, or a later
		// migration counts it as already done.
		m := &Marker{MarkerType: MarkerFace, EmbedModel: face.ModelSFace, DetectModel: face.EngineONNX}
		m.SetEmbeddings(face.Embeddings{}, face.ModelSFace, face.EngineONNX)

		assert.Empty(t, m.EmbedModel)
		assert.Empty(t, m.DetectModel)
	})
	t.Run("ReplacesAPreviousModel", func(t *testing.T) {
		m := &Marker{MarkerType: MarkerFace}
		m.SetEmbeddings(face.Embeddings{face.RandomEmbedding()}, face.ModelFaceNet, face.EngineONNX)
		m.SetEmbeddings(face.Embeddings{face.RandomEmbedding()}, face.ModelSFace, face.EngineONNX)

		assert.Equal(t, face.ModelSFace, m.EmbedModel)
	})
}

func TestMarker_Clusterable(t *testing.T) {
	t.Run("ClearsBothBars", func(t *testing.T) {
		m := &Marker{Size: face.ClusterSizeThreshold, Score: 100}
		assert.True(t, m.Clusterable())
	})
	t.Run("TooSmall", func(t *testing.T) {
		m := &Marker{Size: face.ClusterSizeThreshold - 1, Score: 100}
		assert.False(t, m.Clusterable())
	})
	t.Run("TooLowScoring", func(t *testing.T) {
		m := &Marker{Size: face.ClusterSizeThreshold, Score: 0}
		assert.False(t, m.Clusterable())
	})
	t.Run("ScoreBarFollowsTheDetector", func(t *testing.T) {
		// A library holds markers from more than one detector and nothing recomputes a score, so
		// judging one by the active detector's bar would exclude it for a calibration it was
		// never scored against.
		score := face.ClusterScore(face.DetectorSCRFD)
		m := &Marker{Size: face.ClusterSizeThreshold, Score: score, DetectModel: face.DetectorSCRFD}

		assert.True(t, m.Clusterable())
		assert.Equal(t, score >= face.ClusterScore(""), (&Marker{Size: m.Size, Score: score}).Clusterable())
	})
	t.Run("NilMarker", func(t *testing.T) {
		assert.False(t, (*Marker)(nil).Clusterable())
	})
}

// TestMarker_Unmatched pins the flag a conflict has to leave behind. ClearFace stamps, which is
// right where the matcher found no face and wrong after a cluster narrowed underneath a marker:
// a stamped marker is in neither matching pass's set and waits for a forced run.
func TestMarker_Unmatched(t *testing.T) {
	m := &Marker{
		FileUID:        "fs6sg6bw45bnlqdw",
		MarkerType:     MarkerFace,
		MarkerSrc:      SrcImage,
		Size:           100,
		Score:          100,
		EmbedModel:     face.EmbeddingModelName(),
		EmbeddingsJSON: face.Embeddings{face.RandomEmbedding()}.JSON(),
		W:              0.1,
		H:              0.1,
	}
	require.NoError(t, Db().Create(m).Error)
	t.Cleanup(func() { UnscopedDb().Delete(m) })

	require.NoError(t, m.Matched())
	require.NotNil(t, m.MatchedAt)

	stored := FindMarker(m.MarkerUID)
	require.NotNil(t, stored)
	require.NotNil(t, stored.MatchedAt)

	require.NoError(t, m.Unmatched())

	assert.Nil(t, m.MatchedAt)

	stored = FindMarker(m.MarkerUID)
	require.NotNil(t, stored)
	assert.Nil(t, stored.MatchedAt, "the column must be cleared, not only the field")
}

// TestMarker_NamesFace covers the predicate deciding whether choosing a cluster also names it.
//
// Narrower than "a person set this subject": an XMP name labels its own marker only, so it cannot
// mint an identity and must not be withheld as though it could.
func TestMarker_NamesFace(t *testing.T) {
	t.Run("Manual", func(t *testing.T) {
		assert.True(t, (&Marker{SubjUID: "js6sg6b1qekk9jx8", SubjSrc: SrcManual}).NamesFace())
	})
	t.Run("Image", func(t *testing.T) {
		assert.True(t, (&Marker{SubjUID: "js6sg6b1qekk9jx8", SubjSrc: SrcImage}).NamesFace())
	})
	t.Run("Automatic", func(t *testing.T) {
		assert.False(t, (&Marker{SubjUID: "js6sg6b1qekk9jx8", SubjSrc: SrcAuto}).NamesFace())
	})
	t.Run("Xmp", func(t *testing.T) {
		// SetFace refuses to propagate it, so nothing is minted and nothing is withheld.
		assert.False(t, (&Marker{SubjUID: "js6sg6b1qekk9jx8", SubjSrc: SrcXmp}).NamesFace())
	})
	t.Run("NoSubject", func(t *testing.T) {
		assert.False(t, (&Marker{SubjSrc: SrcManual}).NamesFace())
	})
	t.Run("NilMarker", func(t *testing.T) {
		assert.False(t, (*Marker)(nil).NamesFace())
	})
	t.Run("MatchesTheAdoptionBranch", func(t *testing.T) {
		// The point of the method: it must not drift from what SetFace gates the adoption on.
		for _, src := range []string{SrcAuto, SrcXmp, SrcManual, SrcImage, SrcMeta, SrcMarker} {
			m := &Marker{SubjUID: "js6sg6b1qekk9jx8", SubjSrc: src}
			assert.Equal(t, subjSrcSharesFace(src), m.NamesFace(), "source %q", src)
		}
	})
}

// TestMarker_SyncSubjectRelated pins that a manual name updates a cluster's automatic markers only
// when the cluster carries that person, and reports a marker named after someone else to the cluster.
func TestMarker_SyncSubjectRelated(t *testing.T) {
	newSubject := func(t *testing.T, name string) *Subject {
		t.Helper()

		s := NewSubject(name, SubjPerson, SrcManual)
		require.NotNil(t, s)
		require.NoError(t, s.Create())
		t.Cleanup(func() { UnscopedDb().Delete(&Subject{}, "subj_uid = ?", s.SubjUID) })

		return s
	}

	newFace := func(t *testing.T, subjUID string, seed uint64) *Face {
		t.Helper()

		f := NewFace(subjUID, SrcAuto, face.Embeddings{face.FixtureEmbedding(seed)}, face.EmbeddingModelName())
		require.NotNil(t, f)
		require.NoError(t, f.Create())
		t.Cleanup(func() { UnscopedDb().Delete(&Face{}, "id = ?", f.ID) })

		return f
	}

	// newMarkers stores n markers at the given fraction of the distance the cluster accepts.
	newMarkers := func(t *testing.T, f *Face, n int, subjUID, subjSrc string, fraction float64) []string {
		t.Helper()

		uids := make([]string, 0, n)
		dist := fraction * f.AcceptDist()

		for i := range n {
			emb := face.Embeddings{face.FixtureEmbeddingAt(f.Embedding(), dist, uint64(9000+i))}
			m := Marker{
				MarkerUID:      rnd.GenerateUID('m'),
				MarkerType:     MarkerFace,
				SubjUID:        subjUID,
				SubjSrc:        subjSrc,
				FaceID:         f.ID,
				FaceDist:       dist,
				EmbeddingsJSON: emb.JSON(),
				EmbedModel:     f.EmbedModel,
				Size:           face.ClusterSizeThreshold,
				Score:          face.ClusterScore("") + 10,
				MatchedAt:      TimeStamp(),
				W:              0.1,
				H:              0.1,
			}

			require.NoError(t, UnscopedDb().Create(&m).Error)
			t.Cleanup(func() { UnscopedDb().Delete(&Marker{}, "marker_uid = ?", m.MarkerUID) })

			uids = append(uids, m.MarkerUID)
		}

		return uids
	}

	subjects := func(t *testing.T, uids []string) []string {
		t.Helper()

		result := make([]string, len(uids))

		for i, uid := range uids {
			m := FindMarker(uid)
			require.NotNil(t, m, uid)
			result[i] = m.SubjUID
		}

		return result
	}

	repeat := func(s string, n int) []string {
		result := make([]string, n)

		for i := range result {
			result[i] = s
		}

		return result
	}

	// anchored checks that a corrected marker left the cluster for a face of its new person.
	anchored := func(t *testing.T, m *Marker, f *Face, subjUID string) {
		t.Helper()

		t.Cleanup(func() { UnscopedDb().Delete(&Face{}, "subj_uid = ? AND id <> ?", subjUID, f.ID) })
		require.NotEmpty(t, m.FaceID, "the corrected marker gets a face of its own person")
		assert.NotEqual(t, f.ID, m.FaceID, "and leaves the cluster")
		assert.NotNil(t, m.MatchedAt)

		if own := FindFace(m.FaceID); assert.NotNil(t, own) {
			assert.Equal(t, subjUID, own.SubjUID)
		}
	}

	setName := func(t *testing.T, uid, name string) *Marker {
		t.Helper()

		m := FindMarker(uid)
		require.NotNil(t, m)

		changed, err := m.SetName(name, SrcManual)
		require.NoError(t, err)
		require.True(t, changed)
		require.NoError(t, m.Save())

		return FindMarker(uid)
	}

	t.Run("OtherPersonsCluster", func(t *testing.T) {
		carol := newSubject(t, "Sync Related Carol")
		dave := newSubject(t, "Sync Related Dave")
		f := newFace(t, carol.SubjUID, 7101)
		require.Greater(t, 0.6*f.AcceptDist()-face.Epsilon, max(face.AmbiguityDist(), face.CollisionDist), "the correction must narrow")
		near := newMarkers(t, f, 2, carol.SubjUID, SrcAuto, 0.2)
		corrected := newMarkers(t, f, 1, carol.SubjUID, SrcAuto, 0.6)[0]
		far := newMarkers(t, f, 2, carol.SubjUID, SrcAuto, 0.9)

		got := setName(t, corrected, dave.SubjName)
		assert.Equal(t, dave.SubjUID, got.SubjUID)
		assert.Equal(t, SrcManual, got.SubjSrc)
		anchored(t, got, f, dave.SubjUID)

		cluster := FindFace(f.ID)
		require.NotNil(t, cluster)
		assert.Equal(t, carol.SubjUID, cluster.SubjUID)
		assert.Equal(t, 1, cluster.Collisions)
		assert.Equal(t, int(face.RegularFace), cluster.FaceKind, "narrowed rather than ambiguous")
		assert.Less(t, cluster.CollisionRadius, got.Embeddings().Dist(f.Embedding()))

		assert.Equal(t, repeat(carol.SubjUID, 2), subjects(t, near), "markers inside the new radius keep their person")
		assert.Equal(t, repeat("", 2), subjects(t, far), "markers beyond it are released")
		assert.Equal(t, carol.SubjName, FindSubject(carol.SubjUID).SubjName)
	})
	t.Run("AmbiguousCluster", func(t *testing.T) {
		gina := newSubject(t, "Sync Related Gina")
		hank := newSubject(t, "Sync Related Hank")
		f := newFace(t, gina.SubjUID, 7105)
		others := newMarkers(t, f, 2, gina.SubjUID, SrcAuto, 0.5)
		corrected := newMarkers(t, f, 1, gina.SubjUID, SrcAuto, face.AmbiguityDist()/2/f.AcceptDist())[0]

		got := setName(t, corrected, hank.SubjName)
		assert.Equal(t, hank.SubjUID, got.SubjUID)
		anchored(t, got, f, hank.SubjUID)

		cluster := FindFace(f.ID)
		require.NotNil(t, cluster)
		assert.Equal(t, int(face.AmbiguousFace), cluster.FaceKind)
		assert.Equal(t, 1, cluster.Collisions)
		assert.Equal(t, repeat(gina.SubjUID, 2), subjects(t, others))
	})
	t.Run("RejectedMatchInOtherPersonsCluster", func(t *testing.T) {
		xena := newSubject(t, "Sync Related Xena")
		f := newFace(t, xena.SubjUID, 7102)
		auto := newMarkers(t, f, 3, xena.SubjUID, SrcAuto, 0.2)
		unnamed := newMarkers(t, f, 1, "", SrcAuto, 0.2)
		rejected := newMarkers(t, f, 1, "", SrcManual, 0.6)

		got := setName(t, rejected[0], "Sync Related Yuri")
		t.Cleanup(func() { UnscopedDb().Delete(&Subject{}, "subj_uid = ?", got.SubjUID) })
		require.NotEmpty(t, got.SubjUID)
		assert.NotEqual(t, xena.SubjUID, got.SubjUID)
		anchored(t, got, f, got.SubjUID)
		assert.Equal(t, repeat(xena.SubjUID, 3), subjects(t, auto))
		assert.Equal(t, []string{""}, subjects(t, unnamed))
		assert.Equal(t, xena.SubjUID, FindFace(f.ID).SubjUID)
		assert.Equal(t, 1, FindFace(f.ID).Collisions)
	})
	t.Run("UnnamedCluster", func(t *testing.T) {
		erin := newSubject(t, "Sync Related Erin")
		f := newFace(t, "", 7103)
		uids := newMarkers(t, f, 4, "", SrcAuto, 0.5)

		got := setName(t, uids[0], erin.SubjName)
		assert.Equal(t, erin.SubjUID, got.SubjUID)
		assert.Equal(t, f.ID, got.FaceID)
		assert.Equal(t, erin.SubjUID, FindFace(f.ID).SubjUID, "a manual name still names the cluster")
		assert.Zero(t, FindFace(f.ID).Collisions)
		assert.Equal(t, repeat(erin.SubjUID, 3), subjects(t, uids[1:]), "and its automatic markers")
	})
	t.Run("SamePersonsCluster", func(t *testing.T) {
		fred := newSubject(t, "Sync Related Fred")
		f := newFace(t, fred.SubjUID, 7104)
		named := newMarkers(t, f, 2, fred.SubjUID, SrcAuto, 0.5)
		unnamed := newMarkers(t, f, 2, "", SrcAuto, 0.5)

		got := setName(t, unnamed[0], fred.SubjName)
		assert.Equal(t, fred.SubjUID, got.SubjUID)
		assert.Equal(t, f.ID, got.FaceID)
		assert.Zero(t, FindFace(f.ID).Collisions)
		assert.Equal(t, repeat(fred.SubjUID, 2), subjects(t, named))
		assert.Equal(t, []string{fred.SubjUID}, subjects(t, unnamed[1:]), "related markers still follow the cluster's own person")
	})
	t.Run("UnownedNameRenames", func(t *testing.T) {
		ivan := newSubject(t, "Sync Related Ivan")
		f := newFace(t, ivan.SubjUID, 7106)
		uids := newMarkers(t, f, 3, ivan.SubjUID, SrcAuto, 0.5)

		got := setName(t, uids[0], "Sync Related Ivo")
		assert.Equal(t, ivan.SubjUID, got.SubjUID, "a name nobody owns renames the person")
		assert.Equal(t, "Sync Related Ivo", FindSubject(ivan.SubjUID).SubjName)
		assert.Equal(t, f.ID, got.FaceID)
		assert.Zero(t, FindFace(f.ID).Collisions)
		assert.Equal(t, repeat(ivan.SubjUID, 2), subjects(t, uids[1:]))
	})
}

func TestMarker_resolveSubjectCollision(t *testing.T) {
	carol := NewSubject("Resolve Collision Carol", SubjPerson, SrcManual)
	require.NoError(t, carol.Create())
	dave := NewSubject("Resolve Collision Dave", SubjPerson, SrcManual)
	require.NoError(t, dave.Create())
	t.Cleanup(func() { UnscopedDb().Delete(&Subject{}, "subj_uid IN (?)", []string{carol.SubjUID, dave.SubjUID}) })

	newFace := func(t *testing.T, subjUID string, seed uint64) *Face {
		t.Helper()

		f := NewFace(subjUID, SrcAuto, face.Embeddings{face.FixtureEmbedding(seed)}, face.EmbeddingModelName())
		require.NoError(t, f.Create())
		t.Cleanup(func() { UnscopedDb().Delete(&Face{}, "id = ?", f.ID) })

		return f
	}

	// newMarker stores a marker of Dave in the cluster, with an embedding unless dist is negative.
	newMarker := func(t *testing.T, f *Face, dist float64) *Marker {
		t.Helper()

		m := &Marker{
			MarkerUID:  rnd.GenerateUID('m'),
			MarkerType: MarkerFace,
			SubjUID:    dave.SubjUID,
			SubjSrc:    SrcManual,
			FaceID:     f.ID,
			FaceDist:   dist,
			EmbedModel: f.EmbedModel,
			MatchedAt:  TimeStamp(),
			W:          0.1,
			H:          0.1,
		}

		if dist >= 0 {
			m.EmbeddingsJSON = face.Embeddings{face.FixtureEmbeddingAt(f.Embedding(), dist, 1)}.JSON()
		}

		require.NoError(t, UnscopedDb().Create(m).Error)
		t.Cleanup(func() { UnscopedDb().Delete(&Marker{}, "marker_uid = ?", m.MarkerUID) })

		return m
	}

	// detached checks that a marker left its cluster and waits for matching.
	detached := func(t *testing.T, m *Marker) {
		t.Helper()

		assert.Empty(t, m.FaceID)
		assert.Nil(t, m.MatchedAt)

		stored := FindMarker(m.MarkerUID)
		require.NotNil(t, stored)
		assert.Empty(t, stored.FaceID, "detached in the database, not only in memory")
		assert.Nil(t, stored.MatchedAt)
		assert.Equal(t, dave.SubjUID, stored.SubjUID)
	}

	t.Run("OtherPerson", func(t *testing.T) {
		f := newFace(t, carol.SubjUID, 7501)
		m := newMarker(t, f, 0.6*f.AcceptDist())

		require.NoError(t, m.resolveSubjectCollision(FindFace(m.FaceID)))
		detached(t, m)
		assert.Equal(t, 1, FindFace(f.ID).Collisions)
	})
	t.Run("Anchored", func(t *testing.T) {
		f := newFace(t, carol.SubjUID, 7507)
		m := newMarker(t, f, 0.6*f.AcceptDist())
		require.NoError(t, m.Updates(Values{"size": face.ClusterSizeThreshold, "score": face.ClusterScore("") + 10}))
		t.Cleanup(func() { UnscopedDb().Delete(&Face{}, "subj_uid = ?", dave.SubjUID) })

		require.NoError(t, m.resolveSubjectCollision(FindFace(m.FaceID)))
		assert.Equal(t, 1, FindFace(f.ID).Collisions)

		stored := FindMarker(m.MarkerUID)
		require.NotNil(t, stored)
		assert.Equal(t, m.FaceID, stored.FaceID)
		assert.NotEqual(t, f.ID, stored.FaceID)
		assert.NotNil(t, stored.MatchedAt)

		if own := FindFace(stored.FaceID); assert.NotNil(t, own) {
			assert.Equal(t, dave.SubjUID, own.SubjUID)
		}
	})
	t.Run("OwnSeed", func(t *testing.T) {
		// A face named by hand without a cluster seeds a face from its own embedding, which a face of
		// the new person would share.
		m := &Marker{
			MarkerUID:      rnd.GenerateUID('m'),
			MarkerType:     MarkerFace,
			EmbeddingsJSON: face.Embeddings{face.FixtureEmbedding(7509)}.JSON(),
			EmbedModel:     face.EmbeddingModelName(),
			Size:           face.ClusterSizeThreshold,
			Score:          face.ClusterScore("") + 10,
			W:              0.1,
			H:              0.1,
		}
		require.NoError(t, UnscopedDb().Create(m).Error)
		t.Cleanup(func() { UnscopedDb().Delete(&Marker{}, "marker_uid = ?", m.MarkerUID) })

		changed, err := m.SetName(carol.SubjName, SrcManual)
		require.NoError(t, err)
		require.True(t, changed)
		require.NoError(t, m.Save())
		seed := m.FaceID
		require.NotEmpty(t, seed)
		t.Cleanup(func() { UnscopedDb().Delete(&Face{}, "id = ?", seed) })

		changed, err = m.SetName(dave.SubjName, SrcManual)
		require.NoError(t, err)
		require.True(t, changed)
		require.NoError(t, m.Save())

		stored := FindMarker(m.MarkerUID)
		require.NotNil(t, stored)
		assert.Equal(t, dave.SubjUID, stored.SubjUID)
		assert.NotEqual(t, seed, stored.FaceID, "the marker does not stay on the other person's face")
		assert.Empty(t, stored.FaceID, "and is left for matching")
		assert.Nil(t, stored.MatchedAt)
	})
	t.Run("Invalid", func(t *testing.T) {
		f := newFace(t, carol.SubjUID, 7508)
		m := newMarker(t, f, 0.6*f.AcceptDist())
		m.MarkerInvalid = true

		require.NoError(t, m.resolveSubjectCollision(FindFace(m.FaceID)))
		detached(t, m)
		assert.Zero(t, FindFace(f.ID).Collisions, "a region that is not a face is no evidence")
	})
	t.Run("NoEmbeddings", func(t *testing.T) {
		f := newFace(t, carol.SubjUID, 7502)
		m := newMarker(t, f, -1)

		require.NoError(t, m.resolveSubjectCollision(FindFace(m.FaceID)))
		detached(t, m)
		assert.Zero(t, FindFace(f.ID).Collisions, "nothing to compare, so nothing is reported")
	})
	t.Run("ReportFails", func(t *testing.T) {
		f := newFace(t, carol.SubjUID, 7503)
		m := newMarker(t, f, 0.6*f.AcceptDist())
		require.NoError(t, UnscopedDb().Model(&Face{}).Where("id = ?", f.ID).UpdateColumn("embedding_json", []byte{}).Error)

		require.NoError(t, m.resolveSubjectCollision(FindFace(m.FaceID)), "the name is kept")
		detached(t, m)
	})
	t.Run("SamePerson", func(t *testing.T) {
		f := newFace(t, dave.SubjUID, 7504)
		m := newMarker(t, f, 0.6*f.AcceptDist())

		require.NoError(t, m.resolveSubjectCollision(FindFace(m.FaceID)))
		assert.Equal(t, f.ID, FindMarker(m.MarkerUID).FaceID)
		assert.Zero(t, FindFace(f.ID).Collisions)
	})
	t.Run("UnnamedCluster", func(t *testing.T) {
		f := newFace(t, "", 7505)
		m := newMarker(t, f, 0.6*f.AcceptDist())

		require.NoError(t, m.resolveSubjectCollision(FindFace(m.FaceID)))
		assert.Equal(t, f.ID, FindMarker(m.MarkerUID).FaceID)
	})
	t.Run("NoFace", func(t *testing.T) {
		m := &Marker{MarkerUID: rnd.GenerateUID('m'), MarkerType: MarkerFace, SubjUID: dave.SubjUID}
		require.NoError(t, m.resolveSubjectCollision(FindFace(m.FaceID)))
	})
	t.Run("NoUID", func(t *testing.T) {
		f := newFace(t, carol.SubjUID, 7506)
		m := &Marker{MarkerType: MarkerFace, SubjUID: dave.SubjUID, FaceID: f.ID, MatchedAt: TimeStamp()}

		require.NoError(t, m.resolveSubjectCollision(FindFace(m.FaceID)))
		assert.Empty(t, m.FaceID)
		assert.Nil(t, m.MatchedAt)
	})
}

// TestMarker_SetName_Unlinked pins that entering the same name again links a marker without a person,
// while a linked marker with that name is left alone.
func TestMarker_SetName_Unlinked(t *testing.T) {
	t.Cleanup(func() {
		UnscopedDb().Delete(&Subject{}, "subj_name IN (?)", []string{"SetName Unlinked Carl", "SetName Unlinked Xmp", "SetName Unlinked Invalid"})
	})

	t.Run("Unlinked", func(t *testing.T) {
		m := &Marker{MarkerUID: rnd.GenerateUID('m'), MarkerType: MarkerFace, SubjSrc: SrcXmp, MarkerName: "SetName Unlinked Carl"}

		changed, err := m.SetName("SetName Unlinked Carl", SrcManual)
		require.NoError(t, err)
		assert.True(t, changed)
		assert.NotEmpty(t, m.SubjUID)
		assert.Equal(t, SrcManual, m.SubjSrc)
	})
	t.Run("XmpResent", func(t *testing.T) {
		// What any update of an unlinked XMP marker sends back, such as rejecting it.
		m := &Marker{MarkerUID: rnd.GenerateUID('m'), MarkerType: MarkerFace, SubjSrc: SrcXmp, MarkerName: "SetName Unlinked Xmp"}

		changed, err := m.SetName("SetName Unlinked Xmp", SrcXmp)
		require.NoError(t, err)
		assert.False(t, changed)
		assert.Empty(t, m.SubjUID)
		assert.Nil(t, FindSubjectByName("SetName Unlinked Xmp", false), "no person is created")
	})
	t.Run("Invalid", func(t *testing.T) {
		m := &Marker{MarkerUID: rnd.GenerateUID('m'), MarkerType: MarkerFace, SubjSrc: SrcManual, MarkerName: "SetName Unlinked Invalid", MarkerInvalid: true}

		changed, err := m.SetName("SetName Unlinked Invalid", SrcManual)
		require.NoError(t, err)
		assert.False(t, changed)
		assert.Nil(t, FindSubjectByName("SetName Unlinked Invalid", false), "a rejected face names no person")
	})
	t.Run("Linked", func(t *testing.T) {
		subj := FindSubjectByName("SetName Unlinked Carl", false)
		require.NotNil(t, subj)
		m := &Marker{MarkerUID: rnd.GenerateUID('m'), MarkerType: MarkerFace, SubjSrc: SrcAuto, SubjUID: subj.SubjUID, MarkerName: subj.SubjName}

		changed, err := m.SetName(subj.SubjName, SrcManual)
		require.NoError(t, err)
		assert.False(t, changed)
		assert.Equal(t, SrcAuto, m.SubjSrc, "confirming a linked name does not make it manual")
	})
}

func TestMarker_SourceNamesFace(t *testing.T) {
	t.Run("Manual", func(t *testing.T) {
		assert.True(t, (&Marker{SubjSrc: SrcManual}).SourceNamesFace())
	})
	t.Run("Automatic", func(t *testing.T) {
		assert.False(t, (&Marker{SubjSrc: SrcAuto}).SourceNamesFace())
	})
	t.Run("Xmp", func(t *testing.T) {
		assert.False(t, (&Marker{SubjSrc: SrcXmp, MarkerName: "Jane Doe"}).SourceNamesFace())
	})
	t.Run("NilMarker", func(t *testing.T) {
		assert.False(t, (*Marker)(nil).SourceNamesFace())
	})
	t.Run("MatchesThePolicy", func(t *testing.T) {
		for _, src := range []string{SrcAuto, SrcXmp, SrcManual, SrcImage, SrcMeta, SrcMarker} {
			assert.Equal(t, subjSrcSharesFace(src), (&Marker{SubjSrc: src}).SourceNamesFace(), "source %q", src)
		}
	})
}

func TestMarker_RejectedMatch(t *testing.T) {
	t.Run("Rejected", func(t *testing.T) {
		assert.True(t, (&Marker{MarkerType: MarkerFace, SubjSrc: SrcManual}).RejectedMatch())
	})
	t.Run("Automatic", func(t *testing.T) {
		assert.False(t, (&Marker{MarkerType: MarkerFace, SubjSrc: SrcAuto}).RejectedMatch())
	})
	t.Run("Xmp", func(t *testing.T) {
		assert.False(t, (&Marker{MarkerType: MarkerFace, SubjSrc: SrcXmp}).RejectedMatch())
	})
	t.Run("ManualSubject", func(t *testing.T) {
		assert.False(t, (&Marker{MarkerType: MarkerFace, SubjSrc: SrcManual, SubjUID: "js6sg6b1qekk9jx8"}).RejectedMatch())
	})
	t.Run("ManualName", func(t *testing.T) {
		assert.False(t, (&Marker{MarkerType: MarkerFace, SubjSrc: SrcManual, MarkerName: "Jane Doe"}).RejectedMatch())
	})
	t.Run("LabelMarker", func(t *testing.T) {
		assert.False(t, (&Marker{MarkerType: MarkerLabel, SubjSrc: SrcManual}).RejectedMatch())
	})
	t.Run("NilMarker", func(t *testing.T) {
		assert.False(t, (*Marker)(nil).RejectedMatch())
	})
}

func TestRejectedMatchCond(t *testing.T) {
	// Stored rather than built, so the condition is compared with what the database returns.
	shapes := map[string]Marker{
		"Rejected":      {MarkerType: MarkerFace, SubjSrc: SrcManual},
		"Automatic":     {MarkerType: MarkerFace, SubjSrc: SrcAuto},
		"ManualSubject": {MarkerType: MarkerFace, SubjSrc: SrcManual, SubjUID: "js6sg6b1qekk9jx8"},
		"ManualName":    {MarkerType: MarkerFace, SubjSrc: SrcManual, MarkerName: "Jane Doe"},
		"LabelMarker":   {MarkerType: MarkerLabel, SubjSrc: SrcManual},
		"XmpNoName":     {MarkerType: MarkerFace, SubjSrc: SrcXmp},
		"NullSubject":   {MarkerType: MarkerFace, SubjSrc: SrcManual},
	}

	uids := make(map[string]string, len(shapes))

	for name, m := range shapes {
		m.MarkerUID = rnd.GenerateUID('m')
		m.FileUID = rnd.GenerateUID(FileUID)
		m.W, m.H = 0.1, 0.1
		require.NoError(t, UnscopedDb().Create(&m).Error)
		uids[name] = m.MarkerUID

		t.Cleanup(func() { UnscopedDb().Delete(&Marker{}, "marker_uid = ?", m.MarkerUID) })
	}

	require.NoError(t, UnscopedDb().Exec("UPDATE markers SET subj_uid = NULL, marker_name = NULL WHERE marker_uid = ?", uids["NullSubject"]).Error)

	cond, args := RejectedMatchCond()

	for name, uid := range uids {
		var m Marker
		require.NoError(t, UnscopedDb().Where("marker_uid = ?", uid).First(&m).Error)

		var n int64
		require.NoError(t, UnscopedDb().Model(&Marker{}).Where("marker_uid = ?", uid).Where(cond, args...).Count(&n).Error)
		assert.Equal(t, m.RejectedMatch(), n == int64(1), name)
	}
}

func TestMarker_Embeddings_Normalized(t *testing.T) {
	t.Run("ScalesToUnitLength", func(t *testing.T) {
		// Distances are stated for unit vectors, so a stored vector of another length has to
		// be scaled on read rather than compared as it is.
		m := &Marker{EmbeddingsJSON: []byte(`[[0.1,0.2,0.3,0.4]]`)}
		result := m.Embeddings()
		assert.Len(t, result, 1)
		assert.True(t, result[0].Unit())
		assert.InDelta(t, 0.1/0.5477225575, result[0][0], 1e-9)
	})
	t.Run("KeepsUnitVector", func(t *testing.T) {
		m := &Marker{EmbeddingsJSON: []byte(`[[0.6,0.8]]`)}
		assert.Equal(t, face.Embedding{0.6, 0.8}, m.Embeddings()[0])
	})
	t.Run("Empty", func(t *testing.T) {
		m := &Marker{}
		assert.Empty(t, m.Embeddings())
	})
	t.Run("InvalidJSON", func(t *testing.T) {
		m := &Marker{EmbeddingsJSON: []byte(`{`)}
		assert.Empty(t, m.Embeddings())
	})
}

func TestMarker_CropArea(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		m := Marker{X: 0.1, Y: 0.2, W: 0.3, H: 0.4}
		result := m.CropArea()

		assert.Equal(t, "face", result.Name)
		assert.Equal(t, float32(0.1), result.X)
		assert.Equal(t, float32(0.4), result.H)
	})
	t.Run("Empty", func(t *testing.T) {
		m := Marker{}

		assert.Zero(t, m.CropArea().W)
	})
}

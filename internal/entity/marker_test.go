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
	assert.False(t, atBar.MarkerReview, "a marker that can contribute to a cluster does not")
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
		assert.Equal(t, "Marker<nil>", fmt.Sprintf("%s", m))
	})
	t.Run("New", func(t *testing.T) {
		m := &Marker{}
		assert.Equal(t, "*Marker", m.String())
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

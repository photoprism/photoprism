package entity

import (
	"testing"

	"github.com/photoprism/photoprism/internal/form"
	"github.com/stretchr/testify/assert"
)

func TestMarker_TableName(t *testing.T) {
	m := &Marker{}
	assert.Contains(t, m.TableName(), "markers")
}

func TestNewMarker(t *testing.T) {
	m := NewMarker("ft8es39w45bnlqdw", "lt9k3pw1wowuy3c3", SrcImage, MarkerLabel, 0.308333, 0.206944, 0.355556, 0.355556)
	assert.IsType(t, &Marker{}, m)
	assert.Equal(t, "ft8es39w45bnlqdw", m.FileUID)
	assert.Equal(t, "lt9k3pw1wowuy3c3", m.SubjectUID)
	assert.Equal(t, SrcImage, m.MarkerSrc)
	assert.Equal(t, MarkerLabel, m.MarkerType)
}

func TestMarker_SaveForm(t *testing.T) {
	t.Run("fa-ge add new name to marker then rename marker", func(t *testing.T) {
		m := MarkerFixtures.Get("fa-gr-1")
		m2 := MarkerFixtures.Get("fa-gr-2")
		m3 := MarkerFixtures.Get("fa-gr-3")

		assert.Empty(t, m.SubjectUID)
		assert.Empty(t, m2.SubjectUID)
		assert.Empty(t, m3.SubjectUID)

		m.MarkerInvalid = true
		m.Score = 50

		//set new name

		f := form.Marker{SubjectSrc: SrcManual, MarkerName: "Jane Doe", MarkerInvalid: false}

		err := m.SaveForm(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.NotEmpty(t, m.SubjectUID)

		if s := m.Subject(); s != nil {
			assert.Equal(t, "Jane Doe", s.SubjectName)
		}
		if m := FindMarker("mt9k3pw1wowuy777"); m != nil {
			assert.Equal(t, "Jane Doe", m.Subject().SubjectName)
		}
		if m := FindMarker("mt9k3pw1wowuy888"); m != nil {
			assert.Equal(t, "Jane Doe", m.Subject().SubjectName)
		}

		// Rename subject.
		f3 := form.Marker{SubjectSrc: SrcManual, MarkerName: "Franzilein", MarkerInvalid: false}

		if m := FindMarker("mt9k3pw1wowuy777"); m == nil {
			t.Fatal("result is nil")
		} else if err := m.SaveForm(f3); err != nil {
			t.Fatal(err)
		}

		if m := FindMarker("mt9k3pw1wowuy666"); m != nil {
			assert.Equal(t, "Franzilein", m.Subject().SubjectName)
		}
		if m := FindMarker("mt9k3pw1wowuy777"); m != nil {
			assert.Equal(t, "Franzilein", m.Subject().SubjectName)
		}
		if m := FindMarker("mt9k3pw1wowuy888"); m != nil {
			assert.Equal(t, "Franzilein", m.Subject().SubjectName)
		}
	})
}

func TestUpdateOrCreateMarker(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		m := NewMarker("ft8es39w45bnlqdw", "lt9k3pw1wowuy3c3", SrcImage, MarkerLabel, 0.308333, 0.206944, 0.355556, 0.355556)
		assert.IsType(t, &Marker{}, m)
		assert.Equal(t, "ft8es39w45bnlqdw", m.FileUID)
		assert.Equal(t, "lt9k3pw1wowuy3c3", m.SubjectUID)
		assert.Equal(t, SrcImage, m.MarkerSrc)
		assert.Equal(t, MarkerLabel, m.MarkerType)

		m, err := UpdateOrCreateMarker(m)

		if err != nil {
			t.Fatal(err)
		}

		if m == nil {
			t.Fatal("result should not be nil")
		}

		if m.MarkerUID == "" || m.FileUID == "" {
			t.Errorf("UIDs should not be empty")
		}
	})
}

func TestMarker_Updates(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		m := NewMarker("ft8es39w45bnlqdw", "lt9k3pw1wowuy3c4", SrcImage, MarkerLabel, 0.308333, 0.206944, 0.355556, 0.355556)
		m, err := UpdateOrCreateMarker(m)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, SrcImage, m.MarkerSrc)
		assert.Equal(t, MarkerLabel, m.MarkerType)

		if err := m.Updates(Marker{MarkerSrc: SrcMeta}); err != nil {
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
	t.Run("success", func(t *testing.T) {
		m := NewMarker("ft8es39w45bnlqdw", "lt9k3pw1wowuy3c4", SrcImage, MarkerLabel, 0.308333, 0.206944, 0.355556, 0.355556)
		m, err := UpdateOrCreateMarker(m)

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

func TestMarker_Save(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		m := NewMarker("ft8es39w45bnlqdw", "lt9k3pw1wowuy3c4", SrcImage, MarkerLabel, 0.308333, 0.206944, 0.355556, 0.355556)
		m, err := UpdateOrCreateMarker(m)

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
		assert.True(t, afterDate.After(initialDate))

		if m.MarkerUID == "" || m.FileUID == "" {
			t.Errorf("UIDs should not be empty")
		}

		p := PhotoFixtures.Get("19800101_000002_D640C559")
		assert.Empty(t, p.Files)
		p.PreloadFiles()
		assert.NotEmpty(t, p.Files)

		// t.Logf("FILES: %#v", p.Files)
	})
	t.Run("invalid position", func(t *testing.T) {
		m := Marker{X: 0, Y: 0}
		err := m.Save()
		assert.Equal(t, "marker: invalid position", err.Error())
	})
}

func TestMarker_ClearSubject(t *testing.T) {
	t.Run("1000003-2", func(t *testing.T) {
		m := MarkerFixtures.Get("1000003-2")

		assert.NotEmpty(t, m.MarkerName)

		err := m.ClearSubject(SrcAuto)

		if err != nil {
			t.Fatal(err)
		}

		assert.Empty(t, m.MarkerName)
	})
	t.Run("actor-1", func(t *testing.T) {
		m := MarkerFixtures.Get("actor-a-4")  // id 18
		m2 := MarkerFixtures.Get("actor-a-3") // id 17
		m3 := MarkerFixtures.Get("actor-a-2") // id 16
		m4 := MarkerFixtures.Get("actor-a-1") // id 15

		assert.Equal(t, "jqy1y111h1njaaad", m.SubjectUID)
		assert.Equal(t, "jqy1y111h1njaaad", m2.SubjectUID)
		assert.Equal(t, "jqy1y111h1njaaad", m3.SubjectUID)
		assert.Equal(t, "jqy1y111h1njaaad", m4.SubjectUID)
		assert.NotNil(t, m.Face())
		assert.NotNil(t, m2.Face())
		assert.NotNil(t, m3.Face())
		assert.NotNil(t, m4.Face())

		if m := FindMarker("mt9k3pw1wowu1002"); m == nil {
			t.Fatal("marker is nil")
		} else if f := m.Face(); f == nil {
			t.Fatal("face is nil")
		}

		assert.Equal(t, "PI6A2XGOTUXEFI7CBF4KCI5I2I3JEJHS", m.Face().ID)
		assert.Equal(t, "PI6A2XGOTUXEFI7CBF4KCI5I2I3JEJHS", m2.Face().ID)
		assert.Equal(t, "PI6A2XGOTUXEFI7CBF4KCI5I2I3JEJHS", m3.Face().ID)
		assert.Equal(t, "PI6A2XGOTUXEFI7CBF4KCI5I2I3JEJHS", m4.Face().ID)
		assert.Equal(t, int(0), FindMarker("mt9k3pw1wowu1002").Face().Collisions)

		// Reset face subject.
		err := m.ClearSubject(SrcAuto)

		if err != nil {
			t.Fatal(err)
		}

		assert.NotNil(t, FindMarker("mt9k3pw1wowu1004"))
		assert.NotNil(t, FindMarker("mt9k3pw1wowu1003"))
		assert.NotNil(t, FindMarker("mt9k3pw1wowu1002"))
		assert.NotNil(t, FindFace("PI6A2XGOTUXEFI7CBF4KCI5I2I3JEJHS"))

		assert.Empty(t, m.SubjectUID)
		assert.Equal(t, "", FindMarker("mt9k3pw1wowu1004").SubjectUID)
		assert.Equal(t, "", FindMarker("mt9k3pw1wowu1003").SubjectUID)
		assert.Equal(t, "", FindMarker("mt9k3pw1wowu1002").SubjectUID)
		assert.Empty(t, m.FaceID)
		assert.Equal(t, "", FindMarker("mt9k3pw1wowu1004").FaceID)
		assert.Equal(t, "", FindMarker("mt9k3pw1wowu1003").FaceID)
		assert.Equal(t, "", FindMarker("mt9k3pw1wowu1002").FaceID)
		assert.Equal(t, int(1), FindFace("PI6A2XGOTUXEFI7CBF4KCI5I2I3JEJHS").Collisions)
	})
}

func TestMarker_ClearFace(t *testing.T) {
	t.Run("1000003-2", func(t *testing.T) {
		m := MarkerFixtures.Get("1000003-2")

		assert.NotEmpty(t, m.FaceID)

		updated, err := m.ClearFace()

		if err != nil {
			t.Fatal(err)
		}

		assert.True(t, updated)
		assert.Empty(t, m.FaceID)
	})
	t.Run("empty face id", func(t *testing.T) {
		m := Marker{FaceID: ""}

		updated, err := m.ClearFace()

		if err != nil {
			t.Fatal(err)
		}

		assert.False(t, updated)
		assert.Empty(t, m.FaceID)
	})
	t.Run("subject src manual", func(t *testing.T) {
		m := Marker{FaceID: "123ab"}

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
}

func TestMarker_SyncSubject(t *testing.T) {
	t.Run("no face marker", func(t *testing.T) {
		m := Marker{MarkerType: "test", subject: nil}
		assert.Nil(t, m.SyncSubject(false))
	})
	t.Run("subject is nil", func(t *testing.T) {
		m := Marker{MarkerType: MarkerFace, subject: nil}
		assert.Nil(t, m.SyncSubject(false))
	})
}

func TestMarker_Create(t *testing.T) {
	t.Run("invalid position", func(t *testing.T) {
		m := Marker{X: 0, Y: 0}
		err := m.Create()
		assert.Equal(t, "marker: invalid position", err.Error())
	})
}

func TestMarker_Embeddings(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		m := MarkerFixtures.Get("1000003-4")

		assert.Equal(t, 0.013083286379677253, m.Embeddings()[0][0])
	})
	t.Run("empty embedding", func(t *testing.T) {
		m := Marker{}
		m.EmbeddingsJSON = []byte("")

		assert.Empty(t, m.Embeddings())
	})
	t.Run("invalid embedding json", func(t *testing.T) {
		m := Marker{}
		m.EmbeddingsJSON = []byte("[false]")

		assert.Empty(t, m.Embeddings()[0])
	})
}

func TestMarker_HasFace(t *testing.T) {
	t.Run("true", func(t *testing.T) {
		m := MarkerFixtures.Get("1000003-6")

		assert.True(t, m.HasFace(nil, -1))
		assert.True(t, m.HasFace(FaceFixtures.Pointer("joe-biden"), -1))
	})
	t.Run("false", func(t *testing.T) {
		m := MarkerFixtures.Get("1000003-6")

		assert.False(t, m.HasFace(FaceFixtures.Pointer("joe-biden"), 0.1))
	})
	t.Run("face id empty", func(t *testing.T) {
		m := Marker{FaceID: ""}

		assert.False(t, m.HasFace(FaceFixtures.Pointer("joe-biden"), 0.1))
	})
	t.Run("face dist < 0", func(t *testing.T) {
		m := Marker{FaceID: "123", FaceDist: -1}

		assert.False(t, m.HasFace(FaceFixtures.Pointer("joe-biden"), 0.1))
	})
	t.Run("face id = f.ID", func(t *testing.T) {
		m := Marker{FaceID: "VF7ANLDET2BKZNT4VQWJMMC6HBEFDOG6"}

		assert.True(t, m.HasFace(FaceFixtures.Pointer("joe-biden"), 0.1))
	})
}

func TestMarker_Subject(t *testing.T) {
	t.Run("EmptySubjectUID", func(t *testing.T) {
		m := Marker{SubjectUID: "", subject: &Subject{SubjectUID: "", SubjectName: "Test Subject"}}

		if s := m.Subject(); s == nil {
			t.Fatal("return value must not be nil")
		} else {
			assert.Equal(t, "Test Subject", s.SubjectName)
			assert.Equal(t, "", m.SubjectUID)
			assert.Equal(t, "", s.SubjectUID)
		}
	})
	t.Run("ConflictingSubjectUID", func(t *testing.T) {
		m := Marker{SubjectUID: "", subject: &Subject{SubjectUID: "xyz", SubjectName: "Test Subject"}}

		if s := m.Subject(); s != nil {
			t.Fatal("return value must be nil")
		}
	})
	t.Run("SubjectSrcAuto", func(t *testing.T) {
		m := Marker{SubjectSrc: SrcAuto, SubjectUID: "", MarkerName: "Hans Mayer"}

		if s := m.Subject(); s != nil {
			t.Fatal("return value must be nil")
		} else {
			assert.Equal(t, "Hans Mayer", m.MarkerName)
			assert.Empty(t, m.SubjectUID)
			assert.Equal(t, SrcAuto, m.SubjectSrc)
		}
	})
	t.Run("SubjectSrcManual", func(t *testing.T) {
		m := Marker{SubjectSrc: SrcManual, SubjectUID: "", MarkerName: "Hans Mayer"}

		if s := m.Subject(); s == nil {
			t.Fatal("return value must not be nil")
		} else {
			assert.Equal(t, "Hans Mayer", s.SubjectName)
			assert.NotEmpty(t, s.SubjectUID)
		}
	})
}

func TestMarker_GetFace(t *testing.T) {
	t.Run("ExistingFaceID", func(t *testing.T) {
		m := Marker{face: &Face{ID: "1234"}, FaceID: "1234"}

		if f := m.Face(); f == nil {
			t.Fatal("return value must not be nil")
		} else {
			assert.Equal(t, "1234", f.ID)
			assert.Equal(t, "1234", m.FaceID)
		}
	})
	t.Run("ConflictingFaceID", func(t *testing.T) {
		m := Marker{face: &Face{ID: "1234"}, FaceID: "8888"}

		if f := m.Face(); f != nil {
			t.Fatal("return value must be nil")
		} else {
			assert.Equal(t, "8888", m.FaceID)
			assert.Nil(t, m.face)
		}
	})
	t.Run("find face with ID", func(t *testing.T) {
		m := Marker{FaceID: "VF7ANLDET2BKZNT4VQWJMMC6HBEFDOG6"}

		if f := m.Face(); f == nil {
			t.Fatal("return value must not be nil")
		} else {
			assert.Equal(t, "jqy3y652h8njw0sx", f.SubjectUID)
		}
	})
	t.Run("low quality marker", func(t *testing.T) {
		m := Marker{FaceID: "", SubjectSrc: SrcManual, Size: 130}

		assert.Nil(t, m.Face())
	})
	t.Run("create face", func(t *testing.T) {
		m := Marker{
			FaceID:         "",
			EmbeddingsJSON: MarkerFixtures.Get("actress-a-1").EmbeddingsJSON,
			SubjectSrc:     SrcManual,
			Size:           160,
			Score:          40,
		}

		if m.Face() == nil {
			t.Fatal("return value must not be nil")
		} else {
			assert.NotEmpty(t, m.Face().ID)
		}
	})
}

func TestFindMarker(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		assert.Nil(t, FindMarker("0000"))
	})
}

func TestMarker_SetFace(t *testing.T) {
	t.Run("face == nil", func(t *testing.T) {
		m := MarkerFixtures.Pointer("1000003-6")
		assert.Equal(t, "PN6QO5INYTUSAATOFL43LL2ABAV5ACZK", m.FaceID)
		updated, _ := m.SetFace(nil, -1)
		assert.False(t, updated)
		assert.Equal(t, "PN6QO5INYTUSAATOFL43LL2ABAV5ACZK", m.FaceID)
	})
	t.Run("wrong marker type", func(t *testing.T) {
		m := Marker{MarkerType: "xxx"}
		updated, _ := m.SetFace(&Face{ID: "99876"}, -1)
		assert.False(t, updated)
		assert.Equal(t, "", m.FaceID)
	})
	t.Run("skip same face", func(t *testing.T) {
		m := Marker{MarkerType: MarkerFace, SubjectUID: "jqu0xs11qekk9jx8", FaceID: "99876uyt"}
		updated, _ := m.SetFace(&Face{ID: "99876uyt", SubjectUID: "jqu0xs11qekk9jx8"}, -1)
		assert.False(t, updated)
		assert.Equal(t, "99876uyt", m.FaceID)
	})
	t.Run("set new face", func(t *testing.T) {
		m := Marker{MarkerType: MarkerFace, SubjectUID: "", FaceID: ""}

		updated, _ := m.SetFace(FaceFixtures.Pointer("john-doe"), -1)
		assert.True(t, updated)
		assert.Equal(t, "PN6QO5INYTUSAATOFL43LL2ABAV5ACZK", m.FaceID)
		updated2, err := m.ClearFace()

		if err != nil {
			t.Fatal(err)
		}

		assert.True(t, updated2)
		assert.Empty(t, m.FaceID)
	})

}

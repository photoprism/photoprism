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
	m := NewMarker(1000000, "lt9k3pw1wowuy3c3", SrcImage, MarkerLabel, 0.308333, 0.206944, 0.355556, 0.355556)
	assert.IsType(t, &Marker{}, m)
	assert.Equal(t, uint(1000000), m.FileID)
	assert.Equal(t, "lt9k3pw1wowuy3c3", m.SubjectUID)
	assert.Equal(t, SrcImage, m.MarkerSrc)
	assert.Equal(t, MarkerLabel, m.MarkerType)
}

func TestUpdateOrCreateMarker(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		m := NewMarker(1000000, "lt9k3pw1wowuy3c3", SrcImage, MarkerLabel, 0.308333, 0.206944, 0.355556, 0.355556)
		assert.IsType(t, &Marker{}, m)
		assert.Equal(t, uint(1000000), m.FileID)
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

		if m.ID <= 0 {
			t.Errorf("ID should be > 0")
		}
	})
}

func TestMarker_Updates(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		m := NewMarker(1000000, "lt9k3pw1wowuy3c4", SrcImage, MarkerLabel, 0.308333, 0.206944, 0.355556, 0.355556)
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

		if m.ID <= 0 {
			t.Errorf("ID should be > 0")
		}
	})
}

func TestMarker_Update(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		m := NewMarker(1000000, "lt9k3pw1wowuy3c4", SrcImage, MarkerLabel, 0.308333, 0.206944, 0.355556, 0.355556)
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

		if m.ID <= 0 {
			t.Errorf("ID should be > 0")
		}
	})
}

func TestMarker_Save(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		m := NewMarker(1000000, "lt9k3pw1wowuy3c4", SrcImage, MarkerLabel, 0.308333, 0.206944, 0.355556, 0.355556)
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

		if m.ID <= 0 {
			t.Errorf("ID should be > 0")
		}

		p := PhotoFixtures.Get("19800101_000002_D640C559")
		assert.Empty(t, p.Files)
		p.PreloadFiles(true)
		assert.NotEmpty(t, p.Files)

		t.Logf("FILES: %#v", p.Files)
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
}

func TestMarker_SaveForm(t *testing.T) {
	t.Run("fa-ge add name of not yet existing subject to marker without subject", func(t *testing.T) {
		m := MarkerFixtures.Get("fa-gr-1")
		m2 := MarkerFixtures.Get("fa-gr-2")
		m3 := MarkerFixtures.Get("fa-gr-3")

		assert.Empty(t, m.SubjectUID)
		assert.Empty(t, m2.SubjectUID)
		assert.Empty(t, m3.SubjectUID)

		m.MarkerInvalid = true
		m.Score = 50

		//set new name

		f := form.Marker{SubjectSrc: SrcManual, MarkerName: "Franki", MarkerInvalid: false}

		err := m.SaveForm(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.NotEmpty(t, m.SubjectUID)
		assert.Equal(t, "Franki", m.GetSubject().SubjectName)
		assert.Equal(t, "Franki", FindMarker(9).GetSubject().SubjectName)
		assert.Equal(t, "Franki", FindMarker(10).GetSubject().SubjectName)

		//rename
		f3 := form.Marker{SubjectSrc: SrcManual, MarkerName: "Franzilein", MarkerInvalid: false}

		err3 := FindMarker(9).SaveForm(f3)

		if err3 != nil {
			t.Fatal(err3)
		}

		assert.Equal(t, "Franzilein", FindMarker(8).GetSubject().SubjectName)
		assert.Equal(t, "Franzilein", FindMarker(9).GetSubject().SubjectName)
		assert.Equal(t, "Franzilein", FindMarker(10).GetSubject().SubjectName)
	})
}

func TestMarker_SyncSubject(t *testing.T) {
	t.Run("no face marker", func(t *testing.T) {
		m := Marker{MarkerType: "test", Subject: nil}
		assert.Nil(t, m.SyncSubject(false))
	})
	t.Run("subject is nil", func(t *testing.T) {
		m := Marker{MarkerType: MarkerFace, Subject: nil}
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
}

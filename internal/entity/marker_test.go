package entity

import (
	"github.com/photoprism/photoprism/internal/form"
	"testing"

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
	t.Run("fa-ge add name of not yet existing subject", func(t *testing.T) {
		m := MarkerFixtures.Get("fa-gr-1")
		m2 := MarkerFixtures.Get("fa-gr-2")
		m3 := MarkerFixtures.Get("fa-gr-3")

		assert.Empty(t, m.SubjectUID)
		assert.Empty(t, m2.SubjectUID)
		assert.Empty(t, m3.SubjectUID)

		f := form.Marker{SubjectSrc: SrcManual, MarkerName: "Franzi", MarkerInvalid: false}

		err := m.SaveForm(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.NotEmpty(t, m.SubjectUID)
		assert.Equal(t, "Franzi", m.GetSubject().SubjectName)
		assert.Equal(t, "Franzi", FindMarker(9).GetSubject().SubjectName)
		assert.Equal(t, "Franzi", FindMarker(10).GetSubject().SubjectName)

	})
}

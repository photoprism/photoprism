package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMarker_TableName(t *testing.T) {
	fileSync := &Marker{}
	assert.Equal(t, "markers_dev", fileSync.TableName())
}

func TestNewMarker(t *testing.T) {
	m := NewMarker("ft8es39w45bnlqdw", "lt9k3pw1wowuy3c3", SrcImage, MarkerLabel)
	assert.IsType(t, &Marker{}, m)
	assert.Equal(t, "ft8es39w45bnlqdw", m.FileUID)
	assert.Equal(t, "lt9k3pw1wowuy3c3", m.RefUID)
	assert.Equal(t, SrcImage, m.MarkerSrc)
	assert.Equal(t, MarkerLabel, m.MarkerType)
}

func TestFirstOrCreateMarker(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		m := NewMarker("ft8es39w45bnlqdw", "lt9k3pw1wowuy3c3", SrcImage, MarkerLabel)
		assert.IsType(t, &Marker{}, m)
		assert.Equal(t, "ft8es39w45bnlqdw", m.FileUID)
		assert.Equal(t, "lt9k3pw1wowuy3c3", m.RefUID)
		assert.Equal(t, SrcImage, m.MarkerSrc)
		assert.Equal(t, MarkerLabel, m.MarkerType)

		m = FirstOrCreateMarker(m)

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
		m := NewMarker("ft8es39w45bnlqdw", "lt9k3pw1wowuy3c4", SrcImage, MarkerLabel)
		m = FirstOrCreateMarker(m)

		assert.Equal(t, SrcImage, m.MarkerSrc)
		assert.Equal(t, MarkerLabel, m.MarkerType)

		err := m.Updates(Marker{MarkerSrc: SrcMeta})

		if err != nil {
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
		m := NewMarker("ft8es39w45bnlqdw", "lt9k3pw1wowuy3c4", SrcImage, MarkerLabel)
		m = FirstOrCreateMarker(m)

		assert.Equal(t, MarkerLabel, m.MarkerType)

		err := m.Update("MarkerSrc", SrcMeta)

		if err != nil {
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
		m := NewMarker("ft8es39w45bnlqdw", "lt9k3pw1wowuy3c4", SrcImage, MarkerLabel)
		m = FirstOrCreateMarker(m)

		assert.Equal(t, MarkerLabel, m.MarkerType)

		m.MarkerSrc = SrcMeta

		assert.Equal(t, SrcMeta, m.MarkerSrc)

		initialDate := m.UpdatedAt

		err := m.Save()

		if err != nil {
			t.Fatal(err)
		}

		afterDate := m.UpdatedAt

		assert.Equal(t, SrcMeta, m.MarkerSrc)
		assert.True(t, afterDate.After(initialDate))

		if m.ID <= 0 {
			t.Errorf("ID should be > 0")
		}
	})
}

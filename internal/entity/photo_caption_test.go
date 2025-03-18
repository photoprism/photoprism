package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPhoto_HasCaption(t *testing.T) {
	t.Run("False", func(t *testing.T) {
		photo := Photo{PhotoCaption: ""}
		assert.False(t, photo.HasCaption())
	})
	t.Run("True", func(t *testing.T) {
		photo := Photo{PhotoCaption: "bcss"}
		assert.True(t, photo.HasCaption())
	})
}

func TestPhoto_NoCaption(t *testing.T) {
	t.Run("True", func(t *testing.T) {
		photo := Photo{PhotoCaption: ""}
		assert.True(t, photo.NoCaption())
	})
	t.Run("False", func(t *testing.T) {
		photo := Photo{PhotoCaption: "bcss"}
		assert.False(t, photo.NoCaption())
	})
}

func TestPhoto_GetCaption(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo15")
		assert.Equal(t, "photo caption non-photographic", m.PhotoCaption)
		assert.Equal(t, "photo caption non-photographic", m.GetCaption())
		assert.Equal(t, SrcMeta, m.CaptionSrc)
		assert.Equal(t, SrcMeta, m.GetCaptionSrc())
		assert.Equal(t, false, m.NoCaption())
		assert.Equal(t, true, m.HasCaption())
		assert.Equal(t, false, m.NormalizeValues())
		assert.Equal(t, "photo caption non-photographic", m.PhotoCaption)
		assert.Equal(t, "photo caption non-photographic", m.GetCaption())
		assert.Equal(t, SrcMeta, m.CaptionSrc)
		assert.Equal(t, SrcMeta, m.GetCaptionSrc())
		assert.Equal(t, false, m.NoCaption())
		assert.Equal(t, true, m.HasCaption())
	})
	t.Run("RestoreDescription", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo25")
		assert.Equal(t, "", m.PhotoCaption)
		assert.Equal(t, "", m.CaptionSrc)
		assert.Equal(t, "legacy description", m.PhotoDescription)
		assert.Equal(t, "meta", m.DescriptionSrc)
		assert.Equal(t, "", m.GetCaption())
		assert.Equal(t, "", m.GetCaptionSrc())
		assert.Equal(t, true, m.NoCaption())
		assert.Equal(t, false, m.HasCaption())
		assert.Equal(t, true, m.NormalizeValues())
		assert.Equal(t, "legacy description", m.GetCaption())
		assert.Equal(t, "meta", m.GetCaptionSrc())
		assert.Equal(t, false, m.NoCaption())
		assert.Equal(t, true, m.HasCaption())
		assert.Equal(t, false, m.NormalizeValues())
		assert.Equal(t, "legacy description", m.GetCaption())
		assert.Equal(t, "meta", m.GetCaptionSrc())
		assert.Equal(t, false, m.NoCaption())
		assert.Equal(t, true, m.HasCaption())
	})
}

func TestPhoto_UpdateCaptionLabels(t *testing.T) {
	FirstOrCreateLabel(NewLabel("Food", 1))
	FirstOrCreateLabel(NewLabel("Wine", 2))
	FirstOrCreateLabel(&Label{LabelName: "Bar", LabelSlug: "bar", CustomSlug: "bar", DeletedAt: TimeStamp()})

	t.Run("Success", func(t *testing.T) {
		details := &Details{Keywords: "snake, otter", KeywordsSrc: SrcMeta}
		photo := Photo{ID: 234667, PhotoTitle: "I was in a nice Bar!", TitleSrc: SrcName, PhotoCaption: "globe, wine, food", CaptionSrc: SrcMeta, Details: details}

		if err := photo.Save(); err != nil {
			t.Fatal(err)
		}

		p := FindPhoto(photo)

		assert.Equal(t, 0, len(p.Labels))

		if err := p.UpdateCaptionLabels(); err != nil {
			t.Fatal(err)
		}

		p = FindPhoto(*p)

		assert.Equal(t, "I was in a nice Bar!", p.PhotoTitle)
		assert.Equal(t, "globe, wine, food", p.PhotoCaption)
		assert.Equal(t, "snake, otter", p.Details.Keywords)
		assert.Equal(t, 2, len(p.Labels))
	})
	t.Run("EmptyCaption", func(t *testing.T) {
		details := &Details{Keywords: "snake, otter, food", KeywordsSrc: SrcMeta}
		photo := Photo{ID: 234668, PhotoTitle: "cow, wine, food", TitleSrc: SrcName, PhotoCaption: "", CaptionSrc: SrcMeta, Details: details}

		if err := photo.Save(); err != nil {
			t.Fatal(err)
		}

		p := FindPhoto(photo)

		assert.Equal(t, 0, len(p.Labels))

		if err := p.UpdateCaptionLabels(); err != nil {
			t.Fatal(err)
		}

		p = FindPhoto(*p)

		assert.Equal(t, "cow, wine, food", p.PhotoTitle)
		assert.Equal(t, "", p.PhotoCaption)
		assert.Equal(t, "snake, otter, food", p.Details.Keywords)
		assert.Equal(t, 0, len(p.Labels))
	})
}

func TestPhoto_SetCaption(t *testing.T) {
	t.Run("SetEmptyCaption", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo15")
		assert.Equal(t, "photo caption non-photographic", m.PhotoCaption)
		m.SetCaption("", SrcManual)
		assert.Equal(t, "photo caption non-photographic", m.PhotoCaption)
	})
	t.Run("DescriptionNotFromTheSameSource", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo15")
		assert.Equal(t, "photo caption non-photographic", m.PhotoCaption)
		m.SetCaption("new photo description", SrcName)
		assert.Equal(t, "photo caption non-photographic", m.PhotoCaption)
	})
	t.Run("Ok", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo15")
		assert.Equal(t, "photo caption non-photographic", m.PhotoCaption)
		m.SetCaption("new photo description", SrcMeta)
		assert.Equal(t, "new photo description", m.PhotoCaption)
	})
}

package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/ai/classify"
)

func TestPhoto_HasTitle(t *testing.T) {
	t.Run("false", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo03")
		assert.False(t, m.HasTitle())
	})
	t.Run("true", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo04")
		assert.True(t, m.HasTitle())
	})
}

func TestPhoto_NoTitle(t *testing.T) {
	t.Run("true", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo03")
		assert.True(t, m.NoTitle())
	})
	t.Run("false", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo04")
		assert.False(t, m.NoTitle())
	})
}

func TestPhoto_SetTitle(t *testing.T) {
	t.Run("ManuallyDeleteTitle", func(t *testing.T) {
		// Photo15 has title source "name" (SrcName).
		m := PhotoFixtures.Get("Photo15")
		assert.Equal(t, "TitleToBeSet", m.PhotoTitle)
		// Manually delete existing title.
		m.SetTitle("", SrcManual)
		assert.Equal(t, "", m.PhotoTitle)
	})
	t.Run("LowerSourcePriority", func(t *testing.T) {
		// Photo15 has title source "name" (SrcName).
		m := PhotoFixtures.Get("Photo15")
		assert.Equal(t, "TitleToBeSet", m.PhotoTitle)
		// Set title with lower source priority.
		m.SetTitle("NewTitleSet", SrcAuto)
		assert.Equal(t, "TitleToBeSet", m.PhotoTitle)
	})
	t.Run("SameSourcePriority", func(t *testing.T) {
		// Photo15 has title source "name" (SrcName).
		m := PhotoFixtures.Get("Photo15")
		assert.Equal(t, "TitleToBeSet", m.PhotoTitle)
		// Try to delete existing title with same source priority.
		m.SetTitle("", SrcName)
		assert.Equal(t, "TitleToBeSet", m.PhotoTitle)
		// Replace existing title with same source priority.
		m.SetTitle("NewTitleSet", SrcName)
		assert.Equal(t, "NewTitleSet", m.PhotoTitle)
	})
}

func TestPhoto_GenerateTitle(t *testing.T) {
	t.Run("wont update title was modified", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo08")
		classifyLabels := &classify.Labels{}
		assert.Equal(t, "Black beach", m.PhotoTitle)
		err := m.GenerateTitle(*classifyLabels)
		if err == nil {
			t.Fatal()
		}
		assert.Equal(t, "Black beach", m.PhotoTitle)
	})
	t.Run("photo with location without city and label", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo10")
		classifyLabels := &classify.Labels{{Name: "tree", Uncertainty: 30, Source: "manual", Priority: 5, Categories: []string{"plant"}}}
		assert.Equal(t, "Title", m.PhotoTitle)
		err := m.GenerateTitle(*classifyLabels)
		if err != nil {
			t.Fatal(err)
		}
		// TODO: Unstable
		if len(m.SubjectNames()) > 0 {
			assert.Equal(t, "Actor A / Germany / 2016", m.PhotoTitle)
		} else {
			assert.Equal(t, "Tree / Germany / 2016", m.PhotoTitle)
		}
	})
	t.Run("photo with location and short city and label", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo09")
		classifyLabels := &classify.Labels{{Name: "tree", Uncertainty: 30, Source: "manual", Priority: 5, Categories: []string{"plant"}}}
		assert.Equal(t, "Title", m.PhotoTitle)
		err := m.GenerateTitle(*classifyLabels)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, "Tree / Teotihuacán / 2016", m.PhotoTitle)
	})
	t.Run("photo with location and locname >45", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo13")
		classifyLabels := &classify.Labels{}
		assert.Equal(t, "Title", m.PhotoTitle)
		err := m.GenerateTitle(*classifyLabels)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, "LonglonglonglonglonglonglonglonglonglonglonglonglongName", m.PhotoTitle)
	})
	t.Run("photo with location and locname >20", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo14")
		classifyLabels := &classify.Labels{}
		assert.Equal(t, "Title", m.PhotoTitle)
		err := m.GenerateTitle(*classifyLabels)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, "longlonglonglonglonglongName / 2018", m.PhotoTitle)
	})

	t.Run("photo with location and short city", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo09")
		classifyLabels := &classify.Labels{}
		assert.Equal(t, "Title", m.PhotoTitle)
		err := m.GenerateTitle(*classifyLabels)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, "Adosada Platform / Teotihuacán / 2016", m.PhotoTitle)
	})
	t.Run("photo with location without city", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo10")
		classifyLabels := &classify.Labels{}
		assert.Equal(t, "Title", m.PhotoTitle)
		err := m.GenerateTitle(*classifyLabels)
		if err != nil {
			t.Fatal(err)
		}

		// TODO: Unstable
		if len(m.SubjectNames()) > 0 {
			assert.Equal(t, "Actor A / Germany / 2016", m.PhotoTitle)
		} else {
			assert.Equal(t, "Holiday Park / Germany / 2016", m.PhotoTitle)
		}
	})

	t.Run("photo with location without  loc name and long city", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo11")
		classifyLabels := &classify.Labels{}
		assert.Equal(t, "Title", m.PhotoTitle)
		err := m.GenerateTitle(*classifyLabels)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, "longlonglonglonglongcity / 2016", m.PhotoTitle)
	})
	t.Run("photo with location without loc name and short city", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo12")
		classifyLabels := &classify.Labels{}
		assert.Equal(t, "Title", m.PhotoTitle)
		err := m.GenerateTitle(*classifyLabels)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, "shortcity / Germany / 2016", m.PhotoTitle)
	})
	t.Run("no location original name", func(t *testing.T) {
		m := PhotoFixtures.Get("19800101_000002_D640C559")
		classifyLabels := &classify.Labels{{Name: "classify", Uncertainty: 30, Source: SrcManual, Priority: 5, Categories: []string{"flower", "plant"}}}
		assert.Equal(t, "Lake / 2790", m.PhotoTitle)
		err := m.GenerateTitle(*classifyLabels)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, "Franzilein & Actress A / 2008", m.PhotoTitle)
	})
	t.Run("no location", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo01")
		classifyLabels := &classify.Labels{{Name: "classify", Uncertainty: 30, Source: SrcManual, Priority: 5, Categories: []string{"flower", "plant"}}}
		assert.Equal(t, "", m.PhotoTitle)
		err := m.GenerateTitle(*classifyLabels)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, "Classify / Germany / 2006", m.PhotoTitle)
	})

	t.Run("no location no labels", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo02")
		classifyLabels := &classify.Labels{}
		assert.Equal(t, "", m.PhotoTitle)
		err := m.GenerateTitle(*classifyLabels)
		if err != nil {
			t.Fatal(err)
		}

		// TODO: Unstable
		if len(m.SubjectNames()) > 0 {
			assert.Equal(t, "Actress A / 1990", m.PhotoTitle)
		} else {
			assert.Equal(t, "Bridge1 / 1990", m.PhotoTitle)
		}
	})
	t.Run("no location no labels no takenAt", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo20")
		classifyLabels := &classify.Labels{}
		assert.Equal(t, "", m.PhotoTitle)
		err := m.GenerateTitle(*classifyLabels)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, "Photo20", m.PhotoTitle)
	})
	t.Run("OnePerson", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo10")

		assert.Equal(t, SrcAuto, m.TitleSrc)
		assert.Equal(t, SrcAuto, m.CaptionSrc)
		assert.Equal(t, "Title", m.PhotoTitle)
		assert.Equal(t, "", m.PhotoCaption)

		err := m.GenerateTitle(classify.Labels{})

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, SrcAuto, m.TitleSrc)
		assert.Equal(t, SrcAuto, m.CaptionSrc)

		// TODO: Unstable
		if len(m.SubjectNames()) > 0 {
			assert.Equal(t, "Actor A / Germany / 2016", m.PhotoTitle)
		} else {
			assert.Equal(t, "Holiday Park / Germany / 2016", m.PhotoTitle)
		}

		assert.Equal(t, "", m.PhotoCaption)
	})
	t.Run("People", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo04")

		assert.Equal(t, SrcAuto, m.TitleSrc)
		assert.Equal(t, SrcAuto, m.CaptionSrc)
		assert.Equal(t, "Neckarbrücke", m.PhotoTitle)
		assert.Equal(t, "", m.PhotoCaption)

		err := m.GenerateTitle(classify.Labels{})

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, SrcAuto, m.TitleSrc)
		assert.Equal(t, SrcAuto, m.CaptionSrc)
		assert.Equal(t, "Corn McCornface & Jens Mander / 2014", m.PhotoTitle)
		assert.Equal(t, "", m.PhotoCaption)
	})
}

func TestPhoto_FileTitle(t *testing.T) {
	t.Run("NonLatin", func(t *testing.T) {
		photo := Photo{PhotoName: "桥", PhotoPath: "", OriginalName: ""}
		result := photo.FileTitle()
		assert.Equal(t, "桥", result)
	})
	t.Run("Slash", func(t *testing.T) {
		photo := Photo{PhotoName: "20200102_194030_9EFA9E5E", PhotoPath: "2000/05", OriginalName: "flickr import/changing-of-the-guard--buckingham-palace_7925318070_o.jpg"}
		result := photo.FileTitle()
		assert.Equal(t, "Changing of the Guard / Buckingham Palace", result)
	})
	t.Run("Empty", func(t *testing.T) {
		photo := Photo{PhotoName: "", PhotoPath: "", OriginalName: ""}
		result := photo.FileTitle()
		assert.Equal(t, "", result)
	})
	t.Run("Name", func(t *testing.T) {
		photo := Photo{PhotoName: "sun, beach, fun", PhotoPath: "", OriginalName: "", PhotoTitle: ""}
		result := photo.FileTitle()
		assert.Equal(t, "Sun, Beach, Fun", result)
	})
	t.Run("Path", func(t *testing.T) {
		photo := Photo{PhotoName: "", PhotoPath: "vacation", OriginalName: "20200102_194030_9EFA9E5E", PhotoTitle: ""}
		result := photo.FileTitle()
		assert.Equal(t, "Vacation", result)
	})
}

func TestPhoto_UpdateTitleLabels(t *testing.T) {
	FirstOrCreateLabel(NewLabel("Food", 1))
	FirstOrCreateLabel(NewLabel("Wine", 2))
	FirstOrCreateLabel(&Label{LabelName: "Bar", LabelSlug: "bar", CustomSlug: "bar", DeletedAt: TimeStamp()})

	t.Run("Success", func(t *testing.T) {
		details := &Details{Keywords: "snake, otter, food", KeywordsSrc: SrcMeta}
		photo := Photo{ID: 234567, PhotoTitle: "I was in a nice Wine Bar!", TitleSrc: SrcName, PhotoCaption: "cow, flower, food", CaptionSrc: SrcMeta, Details: details}

		if err := photo.Save(); err != nil {
			t.Fatal(err)
		}

		p := FindPhoto(photo)

		assert.Equal(t, 0, len(p.Labels))

		if err := p.UpdateTitleLabels(); err != nil {
			t.Fatal(err)
		}

		t.Logf("(1) %#v", p.Labels)

		p = FindPhoto(*p)

		t.Logf("(2) %#v", p.Labels)

		assert.Equal(t, "I was in a nice Wine Bar!", p.PhotoTitle)
		assert.Equal(t, "cow, flower, food", p.PhotoCaption)
		assert.Equal(t, "snake, otter, food", p.Details.Keywords)
		assert.Equal(t, 1, len(p.Labels))
	})
	t.Run("EmptyTitle", func(t *testing.T) {
		details := &Details{Keywords: "snake, otter, food", KeywordsSrc: SrcMeta}
		photo := Photo{ID: 234568, PhotoTitle: "", TitleSrc: SrcName, PhotoCaption: "cow, flower, food", CaptionSrc: SrcMeta, Details: details}

		if err := photo.Save(); err != nil {
			t.Fatal(err)
		}

		p := FindPhoto(photo)

		assert.Equal(t, 0, len(p.Labels))

		if err := p.UpdateTitleLabels(); err != nil {
			t.Fatal(err)
		}

		p = FindPhoto(*p)

		assert.Equal(t, "", p.PhotoTitle)
		assert.Equal(t, "cow, flower, food", p.PhotoCaption)
		assert.Equal(t, "snake, otter, food", p.Details.Keywords)
		assert.Equal(t, 0, len(p.Labels))
	})
}

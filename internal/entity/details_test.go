package entity

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestFirstOrCreateDetails(t *testing.T) {
	t.Run("not existing details", func(t *testing.T) {
		details := &Details{PhotoID: 123, Keywords: ""}
		details = FirstOrCreateDetails(details)

		if details == nil {
			t.Fatal("details should not be nil")
		}
	})
	t.Run("existing details", func(t *testing.T) {
		details := &Details{PhotoID: 1000000}
		details = FirstOrCreateDetails(details)

		if details == nil {
			t.Fatal("details should not be nil")
		}
	})
	t.Run("error", func(t *testing.T) {
		details := &Details{PhotoID: 0}
		assert.Nil(t, FirstOrCreateDetails(details))
	})
}

func TestDetails_NoKeywords(t *testing.T) {
	t.Run("no keywords", func(t *testing.T) {
		description := &Details{PhotoID: 123, Keywords: ""}

		assert.Equal(t, true, description.NoKeywords())
	})
	t.Run("keywords", func(t *testing.T) {
		description := &Details{PhotoID: 123, Keywords: "test cat dog", Subject: "animals", Artist: "Bender", Notes: "notes", Copyright: "copy"}

		assert.Equal(t, false, description.NoKeywords())
	})
}

func TestDetails_NoSubject(t *testing.T) {
	t.Run("no subject", func(t *testing.T) {
		description := &Details{PhotoID: 123, Subject: ""}

		assert.Equal(t, true, description.NoSubject())
	})
	t.Run("subject", func(t *testing.T) {
		description := &Details{PhotoID: 123, Keywords: "test cat dog", Subject: "animals", Artist: "Bender", Notes: "notes", Copyright: "copy"}

		assert.Equal(t, false, description.NoSubject())
	})
}

func TestDetails_NoNotes(t *testing.T) {
	t.Run("no notes", func(t *testing.T) {
		description := &Details{PhotoID: 123, Notes: ""}

		assert.Equal(t, true, description.NoNotes())
	})
	t.Run("notes", func(t *testing.T) {
		description := &Details{PhotoID: 123, Keywords: "test cat dog", Subject: "animals", Artist: "Bender", Notes: "notes", Copyright: "copy"}

		assert.Equal(t, false, description.NoNotes())
	})
}

func TestDetails_NoArtist(t *testing.T) {
	t.Run("no artist", func(t *testing.T) {
		description := &Details{PhotoID: 123, Artist: ""}

		assert.Equal(t, true, description.NoArtist())
	})
	t.Run("artist", func(t *testing.T) {
		description := &Details{PhotoID: 123, Keywords: "test cat dog", Subject: "animals", Artist: "Bender", Notes: "notes", Copyright: "copy"}

		assert.Equal(t, false, description.NoArtist())
	})
}

func TestDetails_NoCopyright(t *testing.T) {
	t.Run("no copyright", func(t *testing.T) {
		description := &Details{PhotoID: 123, Copyright: ""}

		assert.Equal(t, true, description.NoCopyright())
	})
	t.Run("copyright", func(t *testing.T) {
		description := &Details{PhotoID: 123, Keywords: "test cat dog", Subject: "animals", Artist: "Bender", Notes: "notes", Copyright: "copy"}

		assert.Equal(t, false, description.NoCopyright())
	})
}

func TestNewDetails(t *testing.T) {
	t.Run("add to photo", func(t *testing.T) {
		p := NewPhoto()
		d := NewDetails(p)
		p.Details = &d
		d.Subject = "Foo Bar"
		d.Keywords = "Baz"

		err := p.Save()

		if err != nil {
			t.Fatal(err)
		}

		t.Logf("PHOTO: %#v", p)
		t.Logf("DETAILS: %#v", d)
	})
}

func TestDetails_Create(t *testing.T) {
	t.Run("error", func(t *testing.T) {
		details := Details{PhotoID: 0}

		assert.Error(t, details.Create())
	})
	t.Run("success", func(t *testing.T) {
		details := Details{PhotoID: 1236799955432}

		err := details.Create()

		if err != nil {
			t.Fatal(err)
		}
	})
}

func TestDetails_Save(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		details := Details{PhotoID: 123678955432, UpdatedAt: time.Date(2020, 2, 1, 0, 0, 0, 0, time.UTC)}
		initialDate := details.UpdatedAt

		err := details.Save()

		if err != nil {
			t.Fatal(err)
		}
		afterDate := details.UpdatedAt

		assert.True(t, afterDate.After(initialDate))
	})

	t.Run("error", func(t *testing.T) {
		details := Details{PhotoID: 0}

		assert.Error(t, details.Save())
	})
}

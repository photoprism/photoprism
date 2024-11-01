package entity

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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
		assert.False(t, description.HasKeywords())
	})
	t.Run("keywords", func(t *testing.T) {
		description := &Details{PhotoID: 123, Keywords: "test cat dog", Subject: "animals", Artist: "Bender", Notes: "notes", Copyright: "copy"}

		assert.Equal(t, false, description.NoKeywords())
		assert.True(t, description.HasKeywords())
	})
}

func TestDetails_NoSubject(t *testing.T) {
	t.Run("no subject", func(t *testing.T) {
		description := &Details{PhotoID: 123, Subject: ""}

		assert.Equal(t, true, description.NoSubject())
		assert.False(t, description.HasSubject())
	})
	t.Run("subject", func(t *testing.T) {
		description := &Details{PhotoID: 123, Keywords: "test cat dog", Subject: "animals", Artist: "Bender", Notes: "notes", Copyright: "copy"}

		assert.Equal(t, false, description.NoSubject())
		assert.True(t, description.HasSubject())
	})
}

func TestDetails_NoNotes(t *testing.T) {
	t.Run("no notes", func(t *testing.T) {
		description := &Details{PhotoID: 123, Notes: ""}

		assert.Equal(t, true, description.NoNotes())
		assert.False(t, description.HasNotes())
	})
	t.Run("notes", func(t *testing.T) {
		description := &Details{PhotoID: 123, Keywords: "test cat dog", Subject: "animals", Artist: "Bender", Notes: "notes", Copyright: "copy"}

		assert.Equal(t, false, description.NoNotes())
		assert.True(t, description.HasNotes())
	})
}

func TestDetails_NoArtist(t *testing.T) {
	t.Run("no artist", func(t *testing.T) {
		description := &Details{PhotoID: 123, Artist: ""}

		assert.Equal(t, true, description.NoArtist())
		assert.False(t, description.HasArtist())

	})
	t.Run("artist", func(t *testing.T) {
		description := &Details{PhotoID: 123, Keywords: "test cat dog", Subject: "animals", Artist: "Bender", Notes: "notes", Copyright: "copy"}

		assert.Equal(t, false, description.NoArtist())
		assert.True(t, description.HasArtist())
	})
}

func TestDetails_NoCopyright(t *testing.T) {
	t.Run("no copyright", func(t *testing.T) {
		description := &Details{PhotoID: 123, Copyright: ""}

		assert.Equal(t, true, description.NoCopyright())
		assert.False(t, description.HasCopyright())
	})
	t.Run("copyright", func(t *testing.T) {
		description := &Details{PhotoID: 123, Keywords: "test cat dog", Subject: "animals", Artist: "Bender", Notes: "notes", Copyright: "copy"}

		assert.Equal(t, false, description.NoCopyright())
		assert.True(t, description.HasCopyright())
	})
}

func TestDetails_NoLicense(t *testing.T) {
	t.Run("no license", func(t *testing.T) {
		description := &Details{PhotoID: 123, License: ""}

		assert.Equal(t, true, description.NoLicense())
		assert.False(t, description.HasLicense())
	})
	t.Run("license", func(t *testing.T) {
		description := &Details{PhotoID: 123, Keywords: "test cat dog", Subject: "animals", Artist: "Bender", Notes: "notes", License: "copy"}

		assert.Equal(t, false, description.NoLicense())
		assert.True(t, description.HasLicense())
	})
}

func TestNewDetails(t *testing.T) {
	t.Run("add to photo", func(t *testing.T) {
		p := NewPhoto(true)

		assert.Equal(t, UnknownTitle, p.PhotoTitle)

		d := NewDetails(p)
		p.Details = &d
		d.Subject = "Foo Bar"
		d.Keywords = "Baz"

		err := p.Save()

		if err != nil {
			t.Fatal(err)
		}

		// t.Logf("PHOTO: %#v", p)
		// t.Logf("DETAILS: %#v", d)
	})
}

// TODO fails on mariadb
func TestDetails_Create(t *testing.T) {
	t.Run("error", func(t *testing.T) {
		details := Details{PhotoID: 0}

		assert.Error(t, details.Create())
	})
	t.Run("Success", func(t *testing.T) {
		details := Details{PhotoID: 1236799955432}

		err := details.Create()

		if err != nil {
			t.Fatal(err)
		}
	})
}

// TODO fails on mariadb
func TestDetails_Save(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
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

func TestDetails_SetKeywords(t *testing.T) {
	t.Run("no keywords", func(t *testing.T) {
		description := &Details{PhotoID: 123, Keywords: ""}
		assert.False(t, description.HasKeywords())

		description.SetKeywords("", "manual")
		assert.False(t, description.HasKeywords())
	})
	t.Run("new keywords have no priority", func(t *testing.T) {
		description := &Details{PhotoID: 123, Keywords: "cat, brown", KeywordsSrc: SrcManual}
		assert.Equal(t, "cat, brown", description.Keywords)

		description.SetKeywords("dog", SrcMeta)
		assert.Equal(t, "cat, brown", description.Keywords)
	})
	t.Run("new keywords set - merge", func(t *testing.T) {
		description := &Details{PhotoID: 123, Keywords: "cat, brown", KeywordsSrc: SrcMeta}
		assert.Equal(t, "cat, brown", description.Keywords)

		description.SetKeywords("dog", SrcMeta)
		assert.Equal(t, "brown, cat, dog", description.Keywords)
	})
	t.Run("new keywords overwrite", func(t *testing.T) {
		description := &Details{PhotoID: 123, Keywords: "cat, brown", KeywordsSrc: SrcMeta}
		assert.Equal(t, "cat, brown", description.Keywords)

		description.SetKeywords("dog", SrcManual)
		assert.Equal(t, "dog", description.Keywords)
	})
}

func TestDetails_SetSubject(t *testing.T) {
	t.Run("no subject", func(t *testing.T) {
		description := &Details{PhotoID: 123, Subject: ""}
		assert.False(t, description.HasSubject())

		description.SetSubject("", "manual")
		assert.False(t, description.HasSubject())
	})
	t.Run("new subject has no priority", func(t *testing.T) {
		description := &Details{PhotoID: 123, Subject: "My cat", SubjectSrc: SrcManual}
		assert.Equal(t, "My cat", description.Subject)

		description.SetSubject("My dog", SrcMeta)
		assert.Equal(t, "My cat", description.Subject)
	})
	t.Run("new subject set", func(t *testing.T) {
		description := &Details{PhotoID: 123, Subject: "My cat", SubjectSrc: SrcMeta}
		assert.Equal(t, "My cat", description.Subject)

		description.SetSubject("My dog", SrcMeta)
		assert.Equal(t, "My dog", description.Subject)
	})
}

func TestDetails_SetNotes(t *testing.T) {
	t.Run("no notes", func(t *testing.T) {
		description := &Details{PhotoID: 123, Notes: ""}
		assert.False(t, description.HasNotes())

		description.SetNotes("", "manual")
		assert.False(t, description.HasNotes())
	})
	t.Run("new notes has no priority", func(t *testing.T) {
		description := &Details{PhotoID: 123, Notes: "My old notes", NotesSrc: SrcManual}
		assert.Equal(t, "My old notes", description.Notes)

		description.SetNotes("My new notes", SrcAuto)
		assert.Equal(t, "My old notes", description.Notes)
	})
	t.Run("new notes set", func(t *testing.T) {
		description := &Details{PhotoID: 123, Notes: "My old notes", NotesSrc: SrcMeta}
		assert.Equal(t, "My old notes", description.Notes)

		description.SetNotes("My new notes", SrcManual)
		assert.Equal(t, "My new notes", description.Notes)
	})
}

func TestDetails_SetArtist(t *testing.T) {
	t.Run("no artist", func(t *testing.T) {
		description := &Details{PhotoID: 123, Artist: ""}
		assert.False(t, description.HasArtist())

		description.SetArtist("", "manual")
		assert.False(t, description.HasArtist())
	})
	t.Run("new artist has no priority", func(t *testing.T) {
		description := &Details{PhotoID: 123, Artist: "Hans", ArtistSrc: SrcManual}
		assert.Equal(t, "Hans", description.Artist)

		description.SetArtist("Maria", SrcAuto)
		assert.Equal(t, "Hans", description.Artist)
	})
	t.Run("new artist set", func(t *testing.T) {
		description := &Details{PhotoID: 123, Artist: "Hans", ArtistSrc: SrcMeta}
		assert.Equal(t, "Hans", description.Artist)

		description.SetArtist("Maria", SrcManual)
		assert.Equal(t, "Maria", description.Artist)
	})
}

func TestDetails_SetCopyright(t *testing.T) {
	t.Run("no copyright", func(t *testing.T) {
		description := &Details{PhotoID: 123, Copyright: ""}
		assert.False(t, description.HasCopyright())

		description.SetCopyright("", "manual")
		assert.False(t, description.HasCopyright())
	})
	t.Run("new copyright has no priority", func(t *testing.T) {
		description := &Details{PhotoID: 123, Copyright: "2018", CopyrightSrc: SrcManual}
		assert.Equal(t, "2018", description.Copyright)

		description.SetCopyright("3000", SrcAuto)
		assert.Equal(t, "2018", description.Copyright)
	})
	t.Run("new copyright set", func(t *testing.T) {
		description := &Details{PhotoID: 123, Copyright: "2018", CopyrightSrc: SrcMeta}
		assert.Equal(t, "2018", description.Copyright)

		description.SetCopyright("3000", SrcManual)
		assert.Equal(t, "3000", description.Copyright)
	})
}

func TestDetails_SetLicense(t *testing.T) {
	t.Run("no license", func(t *testing.T) {
		description := &Details{PhotoID: 123, License: ""}
		assert.False(t, description.HasLicense())

		description.SetLicense("", "manual")
		assert.False(t, description.HasLicense())
	})
	t.Run("new license has no priority", func(t *testing.T) {
		description := &Details{PhotoID: 123, License: "old", LicenseSrc: SrcManual}
		assert.Equal(t, "old", description.License)

		description.SetLicense("new", SrcAuto)
		assert.Equal(t, "old", description.License)
	})
	t.Run("new license set", func(t *testing.T) {
		description := &Details{PhotoID: 123, License: "old", LicenseSrc: SrcMeta}
		assert.Equal(t, "old", description.License)

		description.SetLicense("new", SrcManual)
		assert.Equal(t, "new", description.License)
	})
}

func TestDetails_SetSoftware(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		description := &Details{PhotoID: 123, Software: ""}
		assert.False(t, description.HasSoftware())

		description.SetSoftware("", "manual")
		assert.False(t, description.HasSoftware())
	})
	t.Run("NoPriority", func(t *testing.T) {
		description := &Details{PhotoID: 123, Software: "old", SoftwareSrc: SrcManual}
		assert.Equal(t, "old", description.Software)

		description.SetSoftware("new", SrcAuto)
		assert.Equal(t, "old", description.Software)
	})
	t.Run("NewValue", func(t *testing.T) {
		description := &Details{PhotoID: 123, Software: "old", SoftwareSrc: SrcMeta}
		assert.Equal(t, "old", description.Software)

		description.SetSoftware("new", SrcManual)
		assert.Equal(t, "new", description.Software)
	})
}

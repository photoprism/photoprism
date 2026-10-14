package entity

import (
	"fmt"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/pkg/txt"
)

// longKeywords returns n distinct keywords joined like stored keywords.
func longKeywords(n int) string {
	w := make([]string, n)

	for i := range w {
		w[i] = fmt.Sprintf("schl\u00fcsselwort%03d", i)
	}

	return strings.Join(w, ", ")
}

// findDetails returns the stored details of a photo.
func findDetails(t *testing.T, photoID uint) Details {
	t.Helper()

	var found Details
	require.NoError(t, UnscopedDb().Where("photo_id = ?", photoID).First(&found).Error)

	return found
}

// deleteDetails removes the details of a photo after the test.
func deleteDetails(t *testing.T, photoID uint) {
	t.Cleanup(func() { _ = UnscopedDb().Where("photo_id = ?", photoID).Delete(&Details{}).Error })
}

func TestFirstOrCreateDetails(t *testing.T) {
	t.Run("NotExistingDetails", func(t *testing.T) {
		details := &Details{PhotoID: 123, Keywords: ""}
		details = FirstOrCreateDetails(details)

		if details == nil {
			t.Fatal("details must not be nil")
		}
	})
	t.Run("ExistingDetails", func(t *testing.T) {
		details := &Details{PhotoID: 1000000}
		details = FirstOrCreateDetails(details)

		if details == nil {
			t.Fatal("details must not be nil")
		}
	})
	t.Run("Error", func(t *testing.T) {
		details := &Details{PhotoID: 0}
		assert.Nil(t, FirstOrCreateDetails(details))
	})
}

func TestDetails_NoKeywords(t *testing.T) {
	t.Run("NoKeywords", func(t *testing.T) {
		description := &Details{PhotoID: 123, Keywords: ""}

		assert.Equal(t, true, description.NoKeywords())
		assert.False(t, description.HasKeywords())
	})
	t.Run("Keywords", func(t *testing.T) {
		description := &Details{PhotoID: 123, Keywords: "test cat dog", Subject: "animals", Artist: "Bender", Notes: "notes", Copyright: "copy"}

		assert.Equal(t, false, description.NoKeywords())
		assert.True(t, description.HasKeywords())
	})
}

func TestDetails_NoSubject(t *testing.T) {
	t.Run("NoSubject", func(t *testing.T) {
		description := &Details{PhotoID: 123, Subject: ""}

		assert.Equal(t, true, description.NoSubject())
		assert.False(t, description.HasSubject())
	})
	t.Run("Subject", func(t *testing.T) {
		description := &Details{PhotoID: 123, Keywords: "test cat dog", Subject: "animals", Artist: "Bender", Notes: "notes", Copyright: "copy"}

		assert.Equal(t, false, description.NoSubject())
		assert.True(t, description.HasSubject())
	})
}

func TestDetails_NoNotes(t *testing.T) {
	t.Run("NoNotes", func(t *testing.T) {
		description := &Details{PhotoID: 123, Notes: ""}

		assert.Equal(t, true, description.NoNotes())
		assert.False(t, description.HasNotes())
	})
	t.Run("Notes", func(t *testing.T) {
		description := &Details{PhotoID: 123, Keywords: "test cat dog", Subject: "animals", Artist: "Bender", Notes: "notes", Copyright: "copy"}

		assert.Equal(t, false, description.NoNotes())
		assert.True(t, description.HasNotes())
	})
}

func TestDetails_NoArtist(t *testing.T) {
	t.Run("NoArtist", func(t *testing.T) {
		description := &Details{PhotoID: 123, Artist: ""}

		assert.Equal(t, true, description.NoArtist())
		assert.False(t, description.HasArtist())

	})
	t.Run("Artist", func(t *testing.T) {
		description := &Details{PhotoID: 123, Keywords: "test cat dog", Subject: "animals", Artist: "Bender", Notes: "notes", Copyright: "copy"}

		assert.Equal(t, false, description.NoArtist())
		assert.True(t, description.HasArtist())
	})
}

func TestDetails_NoCopyright(t *testing.T) {
	t.Run("NoCopyright", func(t *testing.T) {
		description := &Details{PhotoID: 123, Copyright: ""}

		assert.Equal(t, true, description.NoCopyright())
		assert.False(t, description.HasCopyright())
	})
	t.Run("Copyright", func(t *testing.T) {
		description := &Details{PhotoID: 123, Keywords: "test cat dog", Subject: "animals", Artist: "Bender", Notes: "notes", Copyright: "copy"}

		assert.Equal(t, false, description.NoCopyright())
		assert.True(t, description.HasCopyright())
	})
}

func TestDetails_NoLicense(t *testing.T) {
	t.Run("NoLicense", func(t *testing.T) {
		description := &Details{PhotoID: 123, License: ""}

		assert.Equal(t, true, description.NoLicense())
		assert.False(t, description.HasLicense())
	})
	t.Run("License", func(t *testing.T) {
		description := &Details{PhotoID: 123, Keywords: "test cat dog", Subject: "animals", Artist: "Bender", Notes: "notes", License: "copy"}

		assert.Equal(t, false, description.NoLicense())
		assert.True(t, description.HasLicense())
	})
}

func TestNewDetails(t *testing.T) {
	t.Run("AddToPhoto", func(t *testing.T) {
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
	t.Run("Error", func(t *testing.T) {
		details := Details{PhotoID: 0}

		assert.Error(t, details.Create())
	})
	t.Run("LongText", func(t *testing.T) {
		details := Details{PhotoID: 900000007, Keywords: longKeywords(320), Notes: strings.Repeat("\u00e4", txt.ClipText+1)}
		deleteDetails(t, details.PhotoID)

		require.NoError(t, details.Create())

		found := findDetails(t, details.PhotoID)
		assert.True(t, strings.HasPrefix(longKeywords(320), found.Keywords+", "))
		assert.Equal(t, txt.ClipText, utf8.RuneCountInString(found.Notes))
	})
	t.Run("Success", func(t *testing.T) {
		details := Details{PhotoID: 900000001}

		err := details.Create()

		if err != nil {
			t.Fatal(err)
		}
	})
}

func TestDetails_Save(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		details := Details{PhotoID: 900000002, UpdatedAt: time.Date(2020, 2, 1, 0, 0, 0, 0, time.UTC)}
		initialDate := details.UpdatedAt

		err := details.Save()

		if err != nil {
			t.Fatal(err)
		}
		afterDate := details.UpdatedAt

		assert.True(t, afterDate.After(initialDate))
	})
	t.Run("Error", func(t *testing.T) {
		details := Details{PhotoID: 0}

		assert.Error(t, details.Save())
	})
	t.Run("LongText", func(t *testing.T) {
		long := strings.Repeat("\u00e4", txt.ClipText+10)
		details := Details{
			PhotoID:   900000004,
			Keywords:  longKeywords(320),
			Notes:     long,
			Subject:   long,
			Artist:    long,
			Copyright: long,
			License:   long,
			Software:  long,
			ArtistSrc: "manualxyz",
			NotesSrc:  SrcManual,
		}
		deleteDetails(t, details.PhotoID)

		require.NoError(t, details.Save())

		found := findDetails(t, details.PhotoID)
		assert.Equal(t, details.Keywords, found.Keywords)
		assert.LessOrEqual(t, utf8.RuneCountInString(found.Keywords), txt.ClipText)
		assert.True(t, strings.HasPrefix(longKeywords(320), found.Keywords+", "))
		assert.Equal(t, txt.ClipText, utf8.RuneCountInString(found.Notes))
		assert.Equal(t, txt.ClipShortText, utf8.RuneCountInString(found.Subject))
		assert.Equal(t, txt.ClipShortText, utf8.RuneCountInString(found.Artist))
		assert.Equal(t, txt.ClipShortText, utf8.RuneCountInString(found.Copyright))
		assert.Equal(t, txt.ClipShortText, utf8.RuneCountInString(found.License))
		assert.Equal(t, txt.ClipShortText, utf8.RuneCountInString(found.Software))
		assert.Equal(t, "", found.ArtistSrc)
		assert.Equal(t, SrcManual, found.NotesSrc)
	})
}

func TestDetails_Updates(t *testing.T) {
	t.Run("LongText", func(t *testing.T) {
		details := Details{PhotoID: 900000005, Keywords: "cat"}
		deleteDetails(t, details.PhotoID)
		require.NoError(t, details.Create())

		values := Values{"keywords": longKeywords(320), "subject": strings.Repeat("a", txt.ClipShortText+1), "keywords_src": SrcBatch}
		require.NoError(t, details.Updates(values))

		found := findDetails(t, details.PhotoID)
		assert.Equal(t, values["keywords"], found.Keywords)
		assert.LessOrEqual(t, utf8.RuneCountInString(found.Keywords), txt.ClipText)
		assert.Equal(t, txt.ClipShortText, len(found.Subject))
		assert.Equal(t, SrcBatch, found.KeywordsSrc)
	})
	t.Run("Nil", func(t *testing.T) {
		details := Details{PhotoID: 900000005}
		assert.NoError(t, details.Updates(nil))
	})
}

func TestDetails_Update(t *testing.T) {
	t.Run("LongText", func(t *testing.T) {
		details := Details{PhotoID: 900000006}
		deleteDetails(t, details.PhotoID)
		require.NoError(t, details.Create())

		require.NoError(t, details.Update("notes", strings.Repeat("a", txt.ClipText+1)))
		assert.Equal(t, txt.ClipText, len(findDetails(t, details.PhotoID).Notes))
	})
	t.Run("InvalidID", func(t *testing.T) {
		details := Details{}
		assert.Error(t, details.Update("notes", "a"))
	})
}

func TestClipDetailsText(t *testing.T) {
	t.Run("Short", func(t *testing.T) {
		assert.Equal(t, " cat, dog ", clipDetailsText("keywords", " cat, dog "))
		assert.Equal(t, " note ", clipDetailsText("notes", " note "))
	})
	t.Run("Keywords", func(t *testing.T) {
		s := longKeywords(320)
		result := clipDetailsText("keywords", s)

		assert.LessOrEqual(t, utf8.RuneCountInString(result), txt.ClipText)
		assert.True(t, strings.HasPrefix(s, result+", "))
	})
	t.Run("KeywordEndsAtLimit", func(t *testing.T) {
		s := strings.Repeat("a", txt.ClipText-5) + ", dog, cat"
		assert.Equal(t, strings.Repeat("a", txt.ClipText-5)+", dog", clipDetailsText("keywords", s))
	})
	t.Run("MultibyteKeywordEndsAtLimit", func(t *testing.T) {
		s := strings.Repeat("\u00fc", txt.ClipText-5) + ", d\u00f6g, cat"
		assert.Equal(t, strings.Repeat("\u00fc", txt.ClipText-5)+", d\u00f6g", clipDetailsText("keywords", s))
	})
	t.Run("FieldName", func(t *testing.T) {
		assert.Equal(t, txt.ClipShortText, len(clipDetailsText("Artist", strings.Repeat("a", txt.ClipShortText+1))))
		assert.Equal(t, "aa, bb", clipDetailsText("Keywords", "aa, bb, "+strings.Repeat("c", txt.ClipText)))
	})
	t.Run("SingleKeyword", func(t *testing.T) {
		assert.Equal(t, strings.Repeat("a", txt.ClipText), clipDetailsText("keywords", strings.Repeat("a", txt.ClipText+1)))
	})
	t.Run("ShortText", func(t *testing.T) {
		for _, col := range []string{"subject", "artist", "copyright", "license", "software"} {
			assert.Equal(t, strings.Repeat("\u00e4", txt.ClipShortText), clipDetailsText(col, strings.Repeat("\u00e4", txt.ClipShortText+1)), col)
		}
	})
	t.Run("Source", func(t *testing.T) {
		assert.Equal(t, SrcManual, clipDetailsText("keywords_src", SrcManual))
		assert.Equal(t, "estimate", clipDetailsText("NotesSrc", "estimate"))
		assert.Equal(t, "", clipDetailsText("artist_src", "manualxyz"))
		assert.Equal(t, "", clipDetailsText("KeywordsSrc", "manualxyz"))
		assert.Equal(t, "", clipDetailsText("software_src", "\u00e4\u00e4\u00e4\u00e4\u00e4"))
	})
	t.Run("OtherColumn", func(t *testing.T) {
		s := strings.Repeat("a", txt.ClipText+1)
		assert.Equal(t, s, clipDetailsText("photo_id", s))
	})
}

func TestDetails_SetKeywords(t *testing.T) {
	t.Run("NoKeywords", func(t *testing.T) {
		description := &Details{PhotoID: 123, Keywords: ""}
		assert.False(t, description.HasKeywords())

		description.SetKeywords("", "manual")
		assert.False(t, description.HasKeywords())
	})
	t.Run("NewKeywordsHaveNoPriority", func(t *testing.T) {
		description := &Details{PhotoID: 123, Keywords: "cat, brown", KeywordsSrc: SrcManual}
		assert.Equal(t, "cat, brown", description.Keywords)

		description.SetKeywords("dog", SrcMeta)
		assert.Equal(t, "cat, brown", description.Keywords)
	})
	t.Run("NewKeywordsSetMerge", func(t *testing.T) {
		description := &Details{PhotoID: 123, Keywords: "cat, brown", KeywordsSrc: SrcMeta}
		assert.Equal(t, "cat, brown", description.Keywords)

		description.SetKeywords("dog", SrcMeta)
		assert.Equal(t, "brown, cat, dog", description.Keywords)
	})
	t.Run("NewKeywordsOverwrite", func(t *testing.T) {
		description := &Details{PhotoID: 123, Keywords: "cat, brown", KeywordsSrc: SrcMeta}
		assert.Equal(t, "cat, brown", description.Keywords)

		description.SetKeywords("dog", SrcManual)
		assert.Equal(t, "dog", description.Keywords)
	})
}

func TestDetails_SetSubject(t *testing.T) {
	t.Run("NoSubject", func(t *testing.T) {
		description := &Details{PhotoID: 123, Subject: ""}
		assert.False(t, description.HasSubject())

		description.SetSubject("", "manual")
		assert.False(t, description.HasSubject())
	})
	t.Run("NewSubjectHasNoPriority", func(t *testing.T) {
		description := &Details{PhotoID: 123, Subject: "My cat", SubjectSrc: SrcManual}
		assert.Equal(t, "My cat", description.Subject)

		description.SetSubject("My dog", SrcMeta)
		assert.Equal(t, "My cat", description.Subject)
	})
	t.Run("NewSubjectSet", func(t *testing.T) {
		description := &Details{PhotoID: 123, Subject: "My cat", SubjectSrc: SrcMeta}
		assert.Equal(t, "My cat", description.Subject)

		description.SetSubject("My dog", SrcMeta)
		assert.Equal(t, "My dog", description.Subject)
	})
}

func TestDetails_SetNotes(t *testing.T) {
	t.Run("NoNotes", func(t *testing.T) {
		description := &Details{PhotoID: 123, Notes: ""}
		assert.False(t, description.HasNotes())

		description.SetNotes("", "manual")
		assert.False(t, description.HasNotes())
	})
	t.Run("NewNotesHasNoPriority", func(t *testing.T) {
		description := &Details{PhotoID: 123, Notes: "My old notes", NotesSrc: SrcManual}
		assert.Equal(t, "My old notes", description.Notes)

		description.SetNotes("My new notes", SrcAuto)
		assert.Equal(t, "My old notes", description.Notes)
	})
	t.Run("NewNotesSet", func(t *testing.T) {
		description := &Details{PhotoID: 123, Notes: "My old notes", NotesSrc: SrcMeta}
		assert.Equal(t, "My old notes", description.Notes)

		description.SetNotes("My new notes", SrcManual)
		assert.Equal(t, "My new notes", description.Notes)
	})
}

func TestDetails_SetArtist(t *testing.T) {
	t.Run("NoArtist", func(t *testing.T) {
		description := &Details{PhotoID: 123, Artist: ""}
		assert.False(t, description.HasArtist())

		description.SetArtist("", "manual")
		assert.False(t, description.HasArtist())
	})
	t.Run("NewArtistHasNoPriority", func(t *testing.T) {
		description := &Details{PhotoID: 123, Artist: "Hans", ArtistSrc: SrcManual}
		assert.Equal(t, "Hans", description.Artist)

		description.SetArtist("Maria", SrcAuto)
		assert.Equal(t, "Hans", description.Artist)
	})
	t.Run("NewArtistSet", func(t *testing.T) {
		description := &Details{PhotoID: 123, Artist: "Hans", ArtistSrc: SrcMeta}
		assert.Equal(t, "Hans", description.Artist)

		description.SetArtist("Maria", SrcManual)
		assert.Equal(t, "Maria", description.Artist)
	})
}

func TestDetails_SetCopyright(t *testing.T) {
	t.Run("NoCopyright", func(t *testing.T) {
		description := &Details{PhotoID: 123, Copyright: ""}
		assert.False(t, description.HasCopyright())

		description.SetCopyright("", "manual")
		assert.False(t, description.HasCopyright())
	})
	t.Run("NewCopyrightHasNoPriority", func(t *testing.T) {
		description := &Details{PhotoID: 123, Copyright: "2018", CopyrightSrc: SrcManual}
		assert.Equal(t, "2018", description.Copyright)

		description.SetCopyright("3000", SrcAuto)
		assert.Equal(t, "2018", description.Copyright)
	})
	t.Run("NewCopyrightSet", func(t *testing.T) {
		description := &Details{PhotoID: 123, Copyright: "2018", CopyrightSrc: SrcMeta}
		assert.Equal(t, "2018", description.Copyright)

		description.SetCopyright("3000", SrcManual)
		assert.Equal(t, "3000", description.Copyright)
	})
}

func TestDetails_SetLicense(t *testing.T) {
	t.Run("NoLicense", func(t *testing.T) {
		description := &Details{PhotoID: 123, License: ""}
		assert.False(t, description.HasLicense())

		description.SetLicense("", "manual")
		assert.False(t, description.HasLicense())
	})
	t.Run("NewLicenseHasNoPriority", func(t *testing.T) {
		description := &Details{PhotoID: 123, License: "old", LicenseSrc: SrcManual}
		assert.Equal(t, "old", description.License)

		description.SetLicense("new", SrcAuto)
		assert.Equal(t, "old", description.License)
	})
	t.Run("NewLicenseSet", func(t *testing.T) {
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

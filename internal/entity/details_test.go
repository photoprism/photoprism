package entity

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFirstOrCreateDetails(t *testing.T) {
	details := &Details{PhotoID: 123, Keywords: ""}
	details = FirstOrCreateDetails(details)

	if details == nil {
		t.Fatal("details should not be nil")
	}
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

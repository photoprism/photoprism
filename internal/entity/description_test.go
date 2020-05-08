package entity

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDescription_FirstOrCreate(t *testing.T) {
	description := &Description{PhotoID: 123, PhotoDescription: ""}
	err := description.FirstOrCreate()
	if err != nil {
		t.Fatal(err)
	}
}

func TestDescription_NoDescription(t *testing.T) {
	t.Run("no description", func(t *testing.T) {
		description := &Description{PhotoID: 123, PhotoDescription: ""}

		assert.Equal(t, true, description.NoDescription())
	})
	t.Run("description", func(t *testing.T) {
		description := &Description{PhotoID: 123, PhotoDescription: "test", PhotoKeywords: "test cat dog", PhotoSubject: "animals", PhotoArtist: "Bender", PhotoNotes: "notes", PhotoCopyright: "copy"}

		assert.Equal(t, false, description.NoDescription())
	})
}

func TestDescription_NoKeywords(t *testing.T) {
	t.Run("no keywords", func(t *testing.T) {
		description := &Description{PhotoID: 123, PhotoDescription: ""}

		assert.Equal(t, true, description.NoKeywords())
	})
	t.Run("keywords", func(t *testing.T) {
		description := &Description{PhotoID: 123, PhotoDescription: "test", PhotoKeywords: "test cat dog", PhotoSubject: "animals", PhotoArtist: "Bender", PhotoNotes: "notes", PhotoCopyright: "copy"}

		assert.Equal(t, false, description.NoKeywords())
	})
}

func TestDescription_NoSubject(t *testing.T) {
	t.Run("no subject", func(t *testing.T) {
		description := &Description{PhotoID: 123, PhotoDescription: ""}

		assert.Equal(t, true, description.NoSubject())
	})
	t.Run("subject", func(t *testing.T) {
		description := &Description{PhotoID: 123, PhotoDescription: "test", PhotoKeywords: "test cat dog", PhotoSubject: "animals", PhotoArtist: "Bender", PhotoNotes: "notes", PhotoCopyright: "copy"}

		assert.Equal(t, false, description.NoSubject())
	})
}

func TestDescription_NoNotes(t *testing.T) {
	t.Run("no notes", func(t *testing.T) {
		description := &Description{PhotoID: 123, PhotoDescription: ""}

		assert.Equal(t, true, description.NoNotes())
	})
	t.Run("notes", func(t *testing.T) {
		description := &Description{PhotoID: 123, PhotoDescription: "test", PhotoKeywords: "test cat dog", PhotoSubject: "animals", PhotoArtist: "Bender", PhotoNotes: "notes", PhotoCopyright: "copy"}

		assert.Equal(t, false, description.NoNotes())
	})
}

func TestDescription_NoArtist(t *testing.T) {
	t.Run("no artist", func(t *testing.T) {
		description := &Description{PhotoID: 123, PhotoDescription: ""}

		assert.Equal(t, true, description.NoArtist())
	})
	t.Run("artist", func(t *testing.T) {
		description := &Description{PhotoID: 123, PhotoDescription: "test", PhotoKeywords: "test cat dog", PhotoSubject: "animals", PhotoArtist: "Bender", PhotoNotes: "notes", PhotoCopyright: "copy"}

		assert.Equal(t, false, description.NoArtist())
	})
}

func TestDescription_NoCopyright(t *testing.T) {
	t.Run("no copyright", func(t *testing.T) {
		description := &Description{PhotoID: 123, PhotoDescription: ""}

		assert.Equal(t, true, description.NoCopyright())
	})
	t.Run("copyright", func(t *testing.T) {
		description := &Description{PhotoID: 123, PhotoDescription: "test", PhotoKeywords: "test cat dog", PhotoSubject: "animals", PhotoArtist: "Bender", PhotoNotes: "notes", PhotoCopyright: "copy"}

		assert.Equal(t, false, description.NoCopyright())
	})
}

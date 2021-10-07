package form

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSubject(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		var m = struct {
			SubjName     string `json:"Name"`
			SubjAlias    string `json:"Alias"`
			SubjBio      string `json:"Bio"`
			SubjNotes    string `json:"Notes"`
			SubjFavorite bool   `json:"Favorite"`
			SubjHidden   bool   `json:"Hidden"`
			SubjPrivate  bool   `json:"Private"`
			SubjExcluded bool   `json:"Excluded"`
		}{
			SubjName:     "Foo",
			SubjAlias:    "bar",
			SubjFavorite: true,
			SubjHidden:   true,
			SubjExcluded: false,
		}

		f, err := NewSubject(m)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "Foo", f.SubjName)
		assert.Equal(t, "bar", f.SubjAlias)
		assert.Equal(t, true, f.SubjFavorite)
		assert.Equal(t, true, f.SubjHidden)
		assert.Equal(t, false, f.SubjExcluded)
	})
}

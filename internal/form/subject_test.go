package form

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewSubject(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
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
	t.Run("Birthday", func(t *testing.T) {
		// A pointer field, which the form has to own rather than share with the model it copies.
		born := time.Date(1990, 8, 1, 0, 0, 0, 0, time.UTC)

		m := struct {
			SubjName     string     `json:"Name"`
			SubjBirthday *time.Time `json:"Birthday"`
		}{SubjName: "Foo", SubjBirthday: &born}

		f, err := NewSubject(m)

		if err != nil {
			t.Fatal(err)
		}

		if assert.NotNil(t, f.SubjBirthday) {
			assert.Equal(t, born, *f.SubjBirthday)
		}

		// Detached, so binding a request writes only into the form.
		assert.NotSame(t, m.SubjBirthday, f.SubjBirthday)

		if err = json.Unmarshal([]byte(`{"Birthday":"1991-09-02T00:00:00Z"}`), f); err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, born, *m.SubjBirthday, "the model keeps the date it was read with")
		assert.Equal(t, time.Date(1991, 9, 2, 0, 0, 0, 0, time.UTC), *f.SubjBirthday)
	})
}

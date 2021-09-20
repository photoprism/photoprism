package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewPerson(t *testing.T) {
	t.Run("BillGates", func(t *testing.T) {
		subj := Subject{
			SubjUID:      "jqytw12v8jjeu3e6",
			SubjName:     "William Henry Gates III",
			SubjAlias:    "Windows Guru",
			SubjFavorite: true,
		}

		m := NewPerson(subj)

		assert.Equal(t, "jqytw12v8jjeu3e6", m.SubjUID)
		assert.Equal(t, "William Henry Gates III", m.SubjName)
		assert.Equal(t, "Windows Guru", m.SubjAlias)
		assert.Equal(t, true, m.SubjFavorite)

		if j, err := m.MarshalJSON(); err != nil {
			t.Fatal(err)
		} else {
			s := string(j)

			expected := "{\"UID\":\"jqytw12v8jjeu3e6\",\"Name\":\"William Henry Gates III\"," +
				"\"Keywords\":[\"william\",\"henry\",\"gates\",\"iii\",\"windows\",\"guru\"]," +
				"\"Favorite\":true}"

			assert.Equal(t, expected, s)
			t.Logf("person json: %s", s)
		}
	})
}

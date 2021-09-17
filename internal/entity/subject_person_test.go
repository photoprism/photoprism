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
			Thumb:        "622c7287967f2800e873fbc55f0328973056ce1d",
		}

		m := NewPerson(subj)

		assert.Equal(t, "jqytw12v8jjeu3e6", m.SubjUID)
		assert.Equal(t, "William Henry Gates III", m.SubjName)
		assert.Equal(t, "Windows Guru", m.SubjAlias)
		assert.Equal(t, true, m.SubjFavorite)
		assert.Equal(t, "622c7287967f2800e873fbc55f0328973056ce1d", m.Thumb)

		if j, err := m.MarshalJSON(); err != nil {
			t.Fatal(err)
		} else {
			s := string(j)

			expected := "{\"UID\":\"jqytw12v8jjeu3e6\",\"Name\":\"William Henry Gates III\"," +
				"\"Keywords\":[\"william\",\"henry\",\"gates\",\"iii\",\"windows\",\"guru\"]," +
				"\"Favorite\":true,\"Thumb\":\"622c7287967f2800e873fbc55f0328973056ce1d\"}"

			assert.Equal(t, expected, s)
			t.Logf("person json: %s", s)
		}
	})
}

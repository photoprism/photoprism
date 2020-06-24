package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFindPersonByUserName(t *testing.T) {
	t.Run("admin", func(t *testing.T) {
		m := FindPersonByUserName("admin")

		if m == nil {
			t.Fatal("result should not be nil")
		}

		assert.Equal(t, 1, m.ID)
		assert.NotEmpty(t, m.PersonUID)
		assert.Equal(t, "admin", m.UserName)
		assert.Equal(t, "Admin", m.DisplayName)
		assert.NotEmpty(t, m.CreatedAt)
		assert.NotEmpty(t, m.UpdatedAt)
	})

	t.Run("unknown", func(t *testing.T) {
		m := FindPersonByUserName("")

		if m != nil {
			t.Fatal("result should be nil")
		}
	})
}

func TestPerson_InvalidPassword(t *testing.T) {
	t.Run("admin", func(t *testing.T) {
		m := FindPersonByUserName("admin")

		if m == nil {
			t.Fatal("result should not be nil")
		}

		assert.False(t, m.InvalidPassword("photoprism"))
	})
}

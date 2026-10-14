package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCount(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		m := PhotoFixtures.Pointer("Photo01")
		_, keys, err := ModelValues(m, "ID", "PhotoUID")

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, 1, Count(m, []string{"ID", "PhotoUID"}, keys))
	})
	t.Run("KeyMismatch", func(t *testing.T) {
		m := PhotoFixtures.Pointer("Photo01")
		other := PhotoFixtures.Pointer("Photo02")

		// Every key must match, not only the primary key of the model.
		assert.Equal(t, 0, Count(m, []string{"ID", "PhotoUID"}, []any{m.ID, other.PhotoUID}))
	})
	t.Run("NonPrimaryKey", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo01")

		// Without a primary key, the count is limited to the rows matching the given key.
		assert.Equal(t, 1, Count(&Photo{}, []string{"PhotoUID"}, []any{m.PhotoUID}))
		assert.Equal(t, 0, Count(&Photo{}, []string{"PhotoUID"}, []any{"ps6sg6be2lvl0ynomatch"}))
	})
	t.Run("InvalidParams", func(t *testing.T) {
		assert.Equal(t, -1, Count(nil, []string{"ID"}, []any{1}))
		assert.Equal(t, -1, Count(&Photo{}, []string{"ID"}, nil))
	})
}

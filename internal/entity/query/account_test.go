package query

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAccountByID(t *testing.T) {
	t.Run("existing account", func(t *testing.T) {
		r, err := AccountByID(uint(1000001))

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "Test Account2", r.AccName)

	})
	t.Run("record not found", func(t *testing.T) {
		r, err := AccountByID(uint(123))

		if err == nil {
			t.Fatal()
		}
		assert.Equal(t, "record not found", err.Error())
		assert.Empty(t, r)
	})
}

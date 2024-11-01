package query

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TODO find duplicates
func TestDuplicates(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		if files, err := Duplicates(10, 0, ""); err != nil {
			t.Fatal(err)
		} else if files == nil {
			t.Fatal("files must not be nil")
		}
	})
	t.Run("pathname not empty", func(t *testing.T) {
		files, err := Duplicates(10, 0, "/holiday/sea.jpg")

		if err != nil {
			t.Fatal(err)
		}

		assert.Empty(t, files)
	})
}

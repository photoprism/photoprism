package query

import (
	"testing"
)

func TestCellIDs(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		result, err := CellIDs()

		if err != nil {
			t.Fatal(err)
		}

		t.Logf("cell count: %v", len(result))
	})
}

func TestUpdatePlaceIDs(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		fixed, err := UpdatePlaceIDs()

		if err != nil {
			t.Fatal(err)
		}

		if fixed < 0 {
			t.Fatal("fixed must be a positive integer")
		}
	})
}

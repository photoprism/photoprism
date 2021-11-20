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

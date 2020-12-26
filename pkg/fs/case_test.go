package fs

import (
	"os"
	"testing"
)

func TestCaseInsensitive(t *testing.T) {
	t.Run("temp", func(t *testing.T) {
		if result, err := CaseInsensitive(os.TempDir()); err != nil {
			t.Fatal(err)
		} else {
			t.Logf("tmp fs case-insensitive: %t", result)
		}
	})
}

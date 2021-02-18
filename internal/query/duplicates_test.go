package query

import (
	"testing"
)

func TestDuplicates(t *testing.T) {
	if files, err := Duplicates(10, 0, ""); err != nil {
		t.Fatal(err)
	} else if files == nil {
		t.Fatal("files must not be nil")
	}
}

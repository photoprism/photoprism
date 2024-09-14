package entity

import (
	"testing"
)

func TestLabelCounts(t *testing.T) {
	results := LabelCounts()

	if len(results) == 0 {
		t.Fatal("at least one result expected")
	}

	for _, result := range results {
		t.Logf("LABEL COUNT: %+v", result)
	}
}

func TestUpdateCounts(t *testing.T) {
	err := UpdateCounts()

	if err != nil {
		t.Fatal(err)
	}
}

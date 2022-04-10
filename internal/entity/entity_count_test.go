package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCount(t *testing.T) {
	m := PhotoFixtures.Pointer("Photo01")
	_, keys, err := ModelValues(m, "ID", "PhotoUID")

	if err != nil {
		t.Fatal(err)
	}

	result := Count(m, []string{"ID", "PhotoUID"}, keys)

	assert.Equal(t, 1, result)
}

func TestLabelCounts(t *testing.T) {
	results := LabelCounts()

	if len(results) == 0 {
		t.Fatal("at least one result expected")
	}

	for _, result := range results {
		t.Logf("LABEL COUNT: %+v", result)
	}
}

func TestUpdatePhotoCounts(t *testing.T) {
	err := UpdateCounts()

	if err != nil {
		t.Fatal(err)
	}
}

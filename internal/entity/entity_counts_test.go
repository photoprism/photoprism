package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
	// countsBusy.Store(true)
	UpdateCountsAsync()
	assert.NoError(t, UpdateCounts())
}

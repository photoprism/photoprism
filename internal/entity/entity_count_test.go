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

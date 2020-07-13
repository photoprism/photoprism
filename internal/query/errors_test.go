package query

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestErrors(t *testing.T) {
	errors, err := Errors(1000, 0, "notexistingErrorString")
	if err != nil {
		t.Fatal(err)
	}
	assert.Empty(t, errors)
}

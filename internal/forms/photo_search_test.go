package forms

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPhotoSearchForm(t *testing.T) {
	form := &PhotoSearchForm{}

	assert.IsType(t, new(PhotoSearchForm), form)
}

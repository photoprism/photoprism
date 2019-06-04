package forms

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	log "github.com/sirupsen/logrus"
)

func TestPhotoSearchForm(t *testing.T) {
	form := &PhotoSearchForm{}

	assert.IsType(t, new(PhotoSearchForm), form)
}

func TestParseQueryString(t *testing.T) {
	form := &PhotoSearchForm{Query: "label:cat query:\"fooBar baz\" before:2019-01-15 camera:23"}

	err := form.ParseQueryString()

	log.Debugf("%+v\n", form)

	assert.Nil(t, err)
	assert.Equal(t, "cat", form.Label)
	assert.Equal(t, "foobar baz", form.Query)
	assert.Equal(t, 23, form.Camera)
	assert.Equal(t, time.Date(2019, 01, 15, 0, 0, 0, 0, time.UTC), form.Before)
}

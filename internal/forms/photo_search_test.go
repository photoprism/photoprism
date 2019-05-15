package forms

import (
	"testing"

	"github.com/stretchr/testify/assert"

	log "github.com/sirupsen/logrus"
)

func TestPhotoSearchForm(t *testing.T) {
	form := &PhotoSearchForm{}

	assert.IsType(t, new(PhotoSearchForm), form)
}

func TestParseQueryString(t *testing.T) {
	form := &PhotoSearchForm{Query: "tags:foo,bar query:\"fooBar baz\" camera:1"}

	err := form.ParseQueryString()

	log.Debugf("%+v\n", form)

	assert.Nil(t, err)
	assert.Equal(t, "foo,bar", form.Tags)
	assert.Equal(t, "foobar baz", form.Query)

}

package form

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAsJson(t *testing.T) {
	form := &SearchAlbums{Query: "slug:album1 favorite:true", Year: "2020"}
	assert.Contains(t, AsJson(form), "{\"Query\":\"slug:album1")
}

func TestAsReader(t *testing.T) {
	form := &SearchAlbums{Query: "slug:album1 favorite:true", Year: "2020"}
	assert.IsType(t, AsReader(form), &strings.Reader{})
}

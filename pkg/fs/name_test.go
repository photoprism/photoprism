package fs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRelativeName(t *testing.T) {
	t.Run("/some/path", func(t *testing.T) {
		assert.Equal(t, "foo/bar.baz", RelativeName("/some/path/foo/bar.baz", "/some/path"))
	})
	t.Run("/some/path/", func(t *testing.T) {
		assert.Equal(t, "foo/bar.baz", RelativeName("/some/path/foo/bar.baz", "/some/path/"))
	})
	t.Run("/some/path/bar", func(t *testing.T) {
		assert.Equal(t, "/some/path/foo/bar.baz", RelativeName("/some/path/foo/bar.baz", "/some/path/bar"))
	})
}

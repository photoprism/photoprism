package list

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAdd(t *testing.T) {
	assert.Equal(t, []string{}, Add([]string{}, ""))
	assert.Equal(t, []string{"bar"}, Add([]string{}, "bar"))
	assert.Equal(t, []string{"foo", "bar"}, Add([]string{"foo", "bar"}, ""))
	assert.Equal(t, []string{"foo", "bar"}, Add([]string{"foo", "bar"}, "foo"))
	assert.Equal(t, []string{"foo", "bar", "zzz"}, Add([]string{"foo", "bar"}, "zzz"))
	assert.Equal(t, []string{"foo", "bar", " "}, Add([]string{"foo", "bar"}, " "))
	assert.Equal(t, []string{"foo", "bar", "645656"}, Add([]string{"foo", "bar"}, "645656"))
	assert.Equal(t, []string{"foo", "bar ", "foo ", "baz", "bar"}, Add([]string{"foo", "bar ", "foo ", "baz"}, "bar"))
	assert.Equal(t, []string{"foo", "bar", "foo ", "baz", "bar "}, Add([]string{"foo", "bar", "foo ", "baz"}, "bar "))
}

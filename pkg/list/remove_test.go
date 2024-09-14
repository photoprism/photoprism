package list

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRemove(t *testing.T) {
	assert.Equal(t, []string{}, Remove([]string{}, ""))
	assert.Equal(t, []string{}, Remove([]string{"foo", "bar"}, "*"))
	assert.Equal(t, []string{}, Remove([]string{}, "bar"))
	assert.Equal(t, []string{"foo", "bar"}, Remove([]string{"foo", "bar"}, ""))
	assert.Equal(t, []string{"bar"}, Remove([]string{"foo", "bar"}, "foo"))
	assert.Equal(t, []string{"foo", "bar"}, Remove([]string{"foo", "bar"}, "zzz"))
	assert.Equal(t, []string{"foo", "bar"}, Remove([]string{"foo", "bar"}, " "))
	assert.Equal(t, []string{"foo", "bar"}, Remove([]string{"foo", "bar"}, "645656"))
	assert.Equal(t, []string{"foo", "bar ", "foo ", "baz"}, Remove([]string{"foo", "bar ", "foo ", "baz"}, "bar"))
	assert.Equal(t, []string{"foo", "bar", "foo ", "baz"}, Remove([]string{"foo", "bar", "foo ", "baz"}, "bar "))
}

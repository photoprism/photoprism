package list

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJoin(t *testing.T) {
	assert.Equal(t, []string{""}, Join([]string{}, []string{""}))
	assert.Equal(t, []string{"bar"}, Join([]string{}, []string{"bar"}))
	assert.Equal(t, []string{""}, Join([]string{""}, []string{}))
	assert.Equal(t, []string{"bar"}, Join([]string{"bar"}, []string{}))
	assert.Equal(t, []string{"foo", "bar"}, Join([]string{"foo", "bar"}, []string{""}))
	assert.Equal(t, []string{"foo", "bar"}, Join([]string{"foo", "bar"}, []string{"foo"}))
	assert.Equal(t, []string{"foo", "bar", "zzz"}, Join([]string{"foo", "bar"}, []string{"zzz"}))
	assert.Equal(t, []string{"foo", "bar", " "}, Join([]string{"foo", "bar"}, []string{" "}))
	assert.Equal(t, []string{"foo", "bar", "645656"}, Join([]string{"foo", "bar"}, []string{"645656"}))
	assert.Equal(t, []string{"foo", "bar ", "foo ", "baz", "bar"}, Join([]string{"foo", "bar ", "foo ", "baz"}, []string{"bar"}))
	assert.Equal(t, []string{"foo", "bar", "foo ", "baz", "bar "}, Join([]string{"foo", "bar", "foo ", "baz"}, []string{"bar "}))
	assert.Equal(t, []string{"bar", "baz", "foo", "bar ", "foo "}, Join([]string{"bar", "baz"}, []string{"foo", "bar ", "foo ", "baz"}))
	assert.Equal(t, []string{"bar", "foo", "foo ", "baz"}, Join([]string{"bar"}, []string{"foo", "bar", "foo ", "baz"}))
}

package list

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExcludes(t *testing.T) {
	t.Run("True", func(t *testing.T) {
		assert.True(t, Excludes([]string{"foo", "bar"}, "baz"))
		assert.True(t, Excludes([]string{"foo", "bar"}, "zzz"))
		assert.True(t, Excludes([]string{"foo", "bar"}, " "))
		assert.True(t, Excludes([]string{"foo", "bar"}, "645656"))
		assert.True(t, Excludes([]string{"foo", "bar ", "foo ", "baz"}, "bar"))
		assert.True(t, Excludes([]string{"foo", "bar", "foo ", "baz"}, "bar "))
	})
	t.Run("False", func(t *testing.T) {
		assert.False(t, Excludes([]string{"foo", "bar"}, "foo"))
		assert.False(t, Excludes([]string{"foo", "bar"}, "bar"))
	})
	t.Run("Empty", func(t *testing.T) {
		assert.False(t, Excludes(nil, ""))
		assert.False(t, Excludes(nil, "foo"))
		assert.False(t, Excludes([]string{}, ""))
		assert.False(t, Excludes([]string{}, "foo"))
		assert.False(t, Excludes([]string{""}, ""))
		assert.False(t, Excludes([]string{"foo", "bar"}, ""))
	})
	t.Run("Wildcard", func(t *testing.T) {
		assert.False(t, Excludes(nil, "*"))
		assert.False(t, Excludes(nil, "* "))
		assert.False(t, Excludes([]string{}, "*"))
		assert.False(t, Excludes([]string{"foo", "*"}, "baz"))
		assert.False(t, Excludes([]string{"foo", "*"}, "foo"))
		assert.False(t, Excludes([]string{""}, "*"))
		assert.False(t, Excludes([]string{"foo", "bar"}, "*"))
	})
}

func TestExcludesAny(t *testing.T) {
	t.Run("True", func(t *testing.T) {
		assert.False(t, ExcludesAny(List{"foo", "bar"}, List{"bar"}))
		assert.False(t, ExcludesAny([]string{"foo", "bAr"}, List{"bAr"}))
		assert.False(t, ExcludesAny([]string{"foo", "bar ", "foo ", "baz"}, List{"foo"}))
		assert.False(t, ExcludesAny([]string{"foo", "bar ", "foo ", "baz"}, List{"foo "}))
		assert.False(t, ExcludesAny([]string{"foo", "bar ", "foo ", "baz"}, List{"bar "}))
	})
	t.Run("False", func(t *testing.T) {
		assert.True(t, ExcludesAny([]string{"foo", "bar"}, List{""}))
		assert.True(t, ExcludesAny([]string{"foo", "bar"}, List{"bAr"}))
		assert.True(t, ExcludesAny([]string{"foo", "bar"}, List{"baz"}))
	})
	t.Run("Empty", func(t *testing.T) {
		assert.False(t, ExcludesAny(nil, nil))
		assert.False(t, ExcludesAny(nil, List{"foo"}))
		assert.False(t, ExcludesAny([]string{}, []string{}))
		assert.False(t, ExcludesAny([]string{}, []string{"foo"}))
		assert.False(t, ExcludesAny(List{}, List{}))
		assert.False(t, ExcludesAny(List{}, List{"foo"}))
		assert.False(t, ExcludesAny([]string{""}, List{}))
		assert.False(t, ExcludesAny([]string{}, List{""}))
		assert.False(t, ExcludesAny([]string{""}, List{""}))
		assert.True(t, ExcludesAny([]string{"foo", "bar"}, List{""}))
	})
	t.Run("Wildcard", func(t *testing.T) {
		assert.False(t, ExcludesAny(nil, List{"*"}))
		assert.False(t, ExcludesAny(nil, List{"* "}))
		assert.False(t, ExcludesAny([]string{}, List{"*"}))
		assert.True(t, ExcludesAny([]string{"foo", "*"}, List{"baz"}))
		assert.False(t, ExcludesAny([]string{"foo", "*"}, List{"foo"}))
		assert.False(t, ExcludesAny([]string{""}, List{"*"}))
		assert.False(t, ExcludesAny([]string{"foo", "bar"}, List{"*"}))
	})
}

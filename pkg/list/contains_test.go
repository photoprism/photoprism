package list

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContains(t *testing.T) {
	t.Run("True", func(t *testing.T) {
		assert.True(t, Contains([]string{"foo", "bar"}, "bar"))
		assert.True(t, Contains([]string{"foo", "bAr"}, "bAr"))
		assert.True(t, Contains([]string{"foo", "bar ", "foo ", "baz"}, "foo"))
		assert.True(t, Contains([]string{"foo", "bar ", "foo ", "baz"}, "foo "))
		assert.True(t, Contains([]string{"foo", "bar ", "foo ", "baz"}, "bar "))
	})
	t.Run("False", func(t *testing.T) {
		assert.False(t, Contains([]string{"foo", "bar"}, ""))
		assert.False(t, Contains([]string{"foo", "bar"}, "bAr"))
		assert.False(t, Contains([]string{"foo", "bar"}, "baz"))
	})
	t.Run("Empty", func(t *testing.T) {
		assert.False(t, Contains(nil, ""))
		assert.False(t, Contains(nil, "foo"))
		assert.False(t, Contains([]string{}, ""))
		assert.False(t, Contains([]string{}, "foo"))
		assert.False(t, Contains([]string{""}, ""))
		assert.False(t, Contains([]string{"foo", "bar"}, ""))
	})
	t.Run("Wildcard", func(t *testing.T) {
		assert.False(t, Contains(nil, "*"))
		assert.False(t, Contains(nil, "* "))
		assert.False(t, Contains([]string{}, "*"))
		assert.True(t, Contains([]string{"foo", "*"}, "baz"))
		assert.True(t, Contains([]string{"foo", "*"}, "foo"))
		assert.True(t, Contains([]string{""}, "*"))
		assert.True(t, Contains([]string{"foo", "bar"}, "*"))
	})
}

func TestContainsAny(t *testing.T) {
	t.Run("True", func(t *testing.T) {
		assert.True(t, ContainsAny(List{"foo", "bar"}, List{"bar"}))
		assert.True(t, ContainsAny([]string{"foo", "bAr"}, List{"bAr"}))
		assert.True(t, ContainsAny([]string{"foo", "bar ", "foo ", "baz"}, List{"foo"}))
		assert.True(t, ContainsAny([]string{"foo", "bar ", "foo ", "baz"}, List{"foo "}))
		assert.True(t, ContainsAny([]string{"foo", "bar ", "foo ", "baz"}, List{"bar "}))
	})
	t.Run("False", func(t *testing.T) {
		assert.False(t, ContainsAny([]string{"foo", "bar"}, List{""}))
		assert.False(t, ContainsAny([]string{"foo", "bar"}, List{"bAr"}))
		assert.False(t, ContainsAny([]string{"foo", "bar"}, List{"baz"}))
	})
	t.Run("Empty", func(t *testing.T) {
		assert.False(t, ContainsAny(nil, nil))
		assert.False(t, ContainsAny(nil, List{"foo"}))
		assert.False(t, ContainsAny([]string{}, []string{}))
		assert.False(t, ContainsAny([]string{}, []string{"foo"}))
		assert.False(t, ContainsAny(List{}, List{}))
		assert.False(t, ContainsAny(List{}, List{"foo"}))
		assert.False(t, ContainsAny([]string{""}, List{}))
		assert.False(t, ContainsAny([]string{}, List{""}))
		assert.True(t, ContainsAny([]string{""}, List{""}))
		assert.False(t, ContainsAny([]string{"foo", "bar"}, List{""}))
	})
	t.Run("Wildcard", func(t *testing.T) {
		assert.False(t, ContainsAny(nil, List{"*"}))
		assert.False(t, ContainsAny(nil, List{"* "}))
		assert.False(t, ContainsAny([]string{}, List{"*"}))
		assert.False(t, ContainsAny([]string{"foo", "*"}, List{"baz"}))
		assert.True(t, ContainsAny([]string{"foo", "*"}, List{"foo"}))
		assert.True(t, ContainsAny([]string{""}, List{"*"}))
		assert.True(t, ContainsAny([]string{"foo", "bar"}, List{"*"}))
	})
}

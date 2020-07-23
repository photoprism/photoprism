package form

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSelection_Empty(t *testing.T) {
	t.Run("not empty photos", func(t *testing.T) {
		sel := Selection{Photos: []string{"foo", "bar"}, Albums: []string{}, Labels: []string{}, Files: []string{}, Places: []string{}}
		assert.Equal(t, false, sel.Empty())
	})
	t.Run("not empty albums", func(t *testing.T) {
		sel := Selection{Photos: []string{}, Albums: []string{"foo", "bar"}, Labels: []string{}, Files: []string{}, Places: []string{}}
		assert.Equal(t, false, sel.Empty())
	})
	t.Run("not empty labels", func(t *testing.T) {
		sel := Selection{Photos: []string{}, Albums: []string{}, Labels: []string{"foo", "bar"}, Files: []string{}, Places: []string{}}
		assert.Equal(t, false, sel.Empty())
	})
	t.Run("not empty files", func(t *testing.T) {
		sel := Selection{Photos: []string{}, Albums: []string{}, Labels: []string{}, Files: []string{"foo", "bar"}, Places: []string{}}
		assert.Equal(t, false, sel.Empty())
	})
	t.Run("not empty places", func(t *testing.T) {
		sel := Selection{Photos: []string{}, Albums: []string{}, Labels: []string{}, Files: []string{}, Places: []string{"foo", "bar"}}
		assert.Equal(t, false, sel.Empty())
	})
	t.Run("empty", func(t *testing.T) {
		sel := Selection{Photos: []string{}, Albums: []string{}, Labels: []string{}}
		assert.Equal(t, true, sel.Empty())
	})
}

func TestSelection_All(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		sel := Selection{Photos: []string{"p123", "p456"}, Albums: []string{"a123"}, Labels: []string{"l123", "l456", "l789"}, Files: []string{"f567", "f111"}, Places: []string{"p568"}}
		assert.Equal(t, []string{"p123", "p456", "a123", "l123", "l456", "l789", "p568"}, sel.All())
	})
}

func TestSelection_String(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		sel := Selection{Photos: []string{"p123", "p456"}, Albums: []string{"a123"}, Labels: []string{"l123", "l456", "l789"}, Files: []string{"f567", "f111"}, Places: []string{"p568"}}
		assert.Equal(t, "p123, p456, a123, l123, l456, l789, p568", sel.String())
	})
}

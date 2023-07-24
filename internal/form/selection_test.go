package form

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
	t.Run("not empty subjects", func(t *testing.T) {
		sel := Selection{Photos: []string{}, Albums: []string{}, Labels: []string{}, Files: []string{}, Places: []string{}, Subjects: []string{"jqzkpo13j8ngpgv4", "jqzkq8j10hj39sxp"}}
		assert.Equal(t, false, sel.Empty())
		assert.Equal(t, []string{"jqzkpo13j8ngpgv4", "jqzkq8j10hj39sxp"}, sel.Subjects)
	})
	t.Run("empty", func(t *testing.T) {
		sel := Selection{Photos: []string{}, Albums: []string{}, Labels: []string{}}
		assert.Equal(t, true, sel.Empty())
	})
}

func TestSelection_Get(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		sel := Selection{Photos: []string{"p123", "p456"}, Albums: []string{"a123"}, Labels: []string{"l123", "l456", "l789"}, Files: []string{"f567", "f111"}, Places: []string{"p568"}, Subjects: []string{"jqzkpo13j8ngpgv4"}}
		assert.Equal(t, []string{"p123", "p456", "a123", "l123", "l456", "l789", "p568", "jqzkpo13j8ngpgv4"}, sel.Get())
	})
}

func TestSelection_String(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		sel := Selection{Photos: []string{"p123", "p456"}, Albums: []string{"a123"}, Labels: []string{"l123", "l456", "l789"}, Files: []string{"f567", "f111"}, Places: []string{"p568"}}
		assert.Equal(t, "p123, p456, a123, l123, l456, l789, p568", sel.String())
	})
}

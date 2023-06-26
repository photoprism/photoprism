package clean

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSearchString(t *testing.T) {
	t.Run("Replace", func(t *testing.T) {
		q := SearchString("table spoon & usa | img% json OR BILL!\n")
		assert.Equal(t, "table spoon & usa | img* json OR BILL!", q)
	})
	t.Run("AndOr", func(t *testing.T) {
		q := SearchString("Jens AND Mander and me Or Kitty AND ")
		assert.Equal(t, "Jens AND Mander and me Or Kitty AND ", q)
	})
	t.Run("FlowersInThePark", func(t *testing.T) {
		q := SearchString(" Flowers in the Park ")
		assert.Equal(t, " Flowers in the Park ", q)
	})
	t.Run("Empty", func(t *testing.T) {
		q := SearchString("")
		assert.Equal(t, "", q)
	})
}

func TestSearchQuery(t *testing.T) {
	t.Run("Replace", func(t *testing.T) {
		q := SearchQuery("table spoon & usa | img% json OR BILL!\n")
		assert.Equal(t, "table spoon & usa | img* json|BILL!", q)
	})
	t.Run("AndOr", func(t *testing.T) {
		q := SearchQuery("Jens AND Mander and me Or Kitty AND ")
		assert.Equal(t, "Jens&Mander&me|Kitty&", q)
	})
	t.Run("FlowersInThePark", func(t *testing.T) {
		q := SearchQuery(" Flowers in the Park ")
		assert.Equal(t, "Flowers&the Park", q)
	})
	t.Run("Empty", func(t *testing.T) {
		q := SearchQuery("")
		assert.Equal(t, "", q)
	})
}

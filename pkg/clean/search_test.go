package clean

import (
	"strings"
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

func BenchmarkSearchQuery_Complex(b *testing.B) {
	s := "Jens AND Mander and me Or Kitty WITH flowers IN the park AT noon | img% json OR BILL!\n"
	b.ReportAllocs()
	for b.Loop() {
		_ = SearchQuery(s)
	}
}

func BenchmarkSearchQuery_Short(b *testing.B) {
	s := "cat and dog"
	b.ReportAllocs()
	for b.Loop() {
		_ = SearchQuery(s)
	}
}

func BenchmarkSearchQuery_LongNoOps(b *testing.B) {
	// No tokens to replace, primarily tests normalization + trim.
	s := strings.Repeat("alpha beta gamma ", 50)
	b.ReportAllocs()
	for b.Loop() {
		_ = SearchQuery(s)
	}
}

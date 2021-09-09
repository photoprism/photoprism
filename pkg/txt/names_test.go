package txt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUniqueNames(t *testing.T) {
	t.Run("ManyNames", func(t *testing.T) {
		result := UniqueNames([]string{"lazy", "jpg", "Brown", "apple", "brown", "new-york", "JPG"})
		assert.Equal(t, []string{"lazy", "jpg", "Brown", "apple", "brown", "new-york", "JPG"}, result)
	})
	t.Run("OneNames", func(t *testing.T) {
		result := UniqueNames([]string{"foo bar"})
		assert.Equal(t, []string{"foo bar"}, result)
	})
	t.Run("None", func(t *testing.T) {
		result := UniqueNames(nil)
		assert.Equal(t, []string{}, result)
	})
}

func TestJoinNames(t *testing.T) {
	t.Run("NoName", func(t *testing.T) {
		result := JoinNames([]string{})
		assert.Equal(t, "", result)
	})
	t.Run("OneName", func(t *testing.T) {
		result := JoinNames([]string{"Jens Mander"})
		assert.Equal(t, "Jens Mander", result)
	})
	t.Run("TwoNames", func(t *testing.T) {
		result := JoinNames([]string{"Jens Mander", "Name 2"})
		assert.Equal(t, "Jens Mander & Name 2", result)
	})
	t.Run("ThreeNames", func(t *testing.T) {
		result := JoinNames([]string{"Jens Mander", "Name 2", "Name 3"})
		assert.Equal(t, "Jens Mander, Name 2 & Name 3", result)
	})
	t.Run("ManyNames", func(t *testing.T) {
		result := JoinNames([]string{"Jens Mander", "Name 2", "Name 3", "Name 4"})
		assert.Equal(t, "Jens Mander, Name 2, Name 3 & Name 4", result)
	})
	t.Run("Partners", func(t *testing.T) {
		result := JoinNames([]string{"Jens Mander", "Jane Mander"})
		assert.Equal(t, "Jens & Jane Mander", result)
	})
	t.Run("Family", func(t *testing.T) {
		result := JoinNames([]string{"Anna Mander", "Jens Mander", "Jane Mander"})
		assert.Equal(t, "Anna, Jens & Jane Mander", result)
	})
	t.Run("ShortFamilyName", func(t *testing.T) {
		result := JoinNames([]string{"Anna M", "Jens M", "Jane M"})
		assert.Equal(t, "Anna M, Jens M & Jane M", result)
	})
	t.Run("NoFamily", func(t *testing.T) {
		result := JoinNames([]string{"Anna Mander", "Jane Mander", "Bill Gates"})
		assert.Equal(t, "Anna Mander, Jane Mander & Bill Gates", result)
	})
}

func TestNameKeywords(t *testing.T) {
	t.Run("BillGates", func(t *testing.T) {
		result := NameKeywords("William Henry Gates III", "Windows Guru")
		assert.Equal(t, []string{"william", "henry", "gates", "iii", "windows", "guru"}, result)
	})
}

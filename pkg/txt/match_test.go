package txt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		assert.False(t, New(""))
	})
	t.Run("EnNew", func(t *testing.T) {
		assert.True(t, New(EnNew))
	})
	t.Run("Spaces", func(t *testing.T) {
		assert.True(t, New("     new "))
	})
	t.Run("Uppercase", func(t *testing.T) {
		assert.True(t, New("NEW"))
	})
	t.Run("Lowercase", func(t *testing.T) {
		assert.True(t, New("new"))
	})
	t.Run("True", func(t *testing.T) {
		assert.True(t, New("New"))
	})
	t.Run("False", func(t *testing.T) {
		assert.False(t, New("non"))
	})
}

func TestBool(t *testing.T) {
	t.Run("NotEmpty", func(t *testing.T) {
		assert.True(t, Bool("Browse your life in pictures"))
	})
	t.Run("Oui", func(t *testing.T) {
		assert.True(t, Bool("oui"))
	})
	t.Run("Non", func(t *testing.T) {
		assert.False(t, Bool("non"))
	})
	t.Run("Ja", func(t *testing.T) {
		assert.True(t, Bool("ja"))
	})
	t.Run("True", func(t *testing.T) {
		assert.True(t, Bool("true"))
	})
	t.Run("Yes", func(t *testing.T) {
		assert.True(t, Bool("yes"))
	})
	t.Run("No", func(t *testing.T) {
		assert.False(t, Bool("no"))
	})
	t.Run("False", func(t *testing.T) {
		assert.False(t, Bool("false"))
	})
	t.Run("Empty", func(t *testing.T) {
		assert.False(t, Bool(""))
	})
	t.Run("UppercaseNo", func(t *testing.T) {
		assert.False(t, Bool("NO"))
	})
}

func TestYes(t *testing.T) {
	t.Run("NotEmpty", func(t *testing.T) {
		assert.False(t, Yes("Browse your life in pictures"))
	})
	t.Run("Oui", func(t *testing.T) {
		assert.True(t, Yes("oui"))
		assert.True(t, Yes("OUI"))
	})
	t.Run("Non", func(t *testing.T) {
		assert.False(t, Yes("non"))
	})
	t.Run("Ja", func(t *testing.T) {
		assert.True(t, Yes("ja"))
	})
	t.Run("True", func(t *testing.T) {
		assert.True(t, Yes("true"))
	})
	t.Run("Yes", func(t *testing.T) {
		assert.True(t, Yes("yes"))
	})
	t.Run("No", func(t *testing.T) {
		assert.False(t, Yes("no"))
	})
	t.Run("False", func(t *testing.T) {
		assert.False(t, Yes("false"))
	})
	t.Run("Exclude", func(t *testing.T) {
		assert.False(t, Yes("exclude"))
	})
	t.Run("Include", func(t *testing.T) {
		assert.True(t, Yes("include"))
	})
	t.Run("Unknown", func(t *testing.T) {
		assert.False(t, Yes("unknown"))
	})
	t.Run("Please", func(t *testing.T) {
		assert.True(t, Yes("please"))
		assert.True(t, Yes("pLeAsE"))
	})
	t.Run("Positive", func(t *testing.T) {
		assert.True(t, Yes("positive"))
	})
	t.Run("Empty", func(t *testing.T) {
		assert.False(t, Yes(""))
	})
	t.Run("Space", func(t *testing.T) {
		assert.False(t, Yes("Yes Please"))
	})
	t.Run("One", func(t *testing.T) {
		assert.True(t, Yes("1"))
	})
	t.Run("Zero", func(t *testing.T) {
		assert.False(t, Yes("0"))
	})
	t.Run("tak", func(t *testing.T) {
		assert.True(t, Yes("так"))
		assert.True(t, Yes("ТАК"))
	})
	t.Run("russian", func(t *testing.T) {
		assert.True(t, Yes("да"))
		assert.True(t, Yes("Да"))
	})
	t.Run("TabSeparatedPhrase", func(t *testing.T) {
		assert.False(t, Yes("yes\tplease"))
	})
	t.Run("NonBreakingSpace", func(t *testing.T) {
		assert.False(t, Yes("yes\u00a0please"))
	})
}

func TestNo(t *testing.T) {
	t.Run("NotEmpty", func(t *testing.T) {
		assert.False(t, No("Browse your life in pictures"))
	})
	t.Run("Oui", func(t *testing.T) {
		assert.False(t, No("oui"))
		assert.False(t, No("OUI"))
	})
	t.Run("Non", func(t *testing.T) {
		assert.True(t, No("non"))
	})
	t.Run("Ja", func(t *testing.T) {
		assert.False(t, No("ja"))
	})
	t.Run("True", func(t *testing.T) {
		assert.False(t, No("true"))
	})
	t.Run("Yes", func(t *testing.T) {
		assert.False(t, No("yes"))
	})
	t.Run("No", func(t *testing.T) {
		assert.True(t, No("no"))
	})
	t.Run("False", func(t *testing.T) {
		assert.True(t, No("false"))
	})
	t.Run("Exclude", func(t *testing.T) {
		assert.True(t, No("exclude"))
	})
	t.Run("Include", func(t *testing.T) {
		assert.False(t, No("include"))
	})
	t.Run("Unknown", func(t *testing.T) {
		assert.True(t, No("unknown"))
	})
	t.Run("Please", func(t *testing.T) {
		assert.False(t, No("please"))
	})
	t.Run("Positive", func(t *testing.T) {
		assert.False(t, No("positive"))
	})
	t.Run("Empty", func(t *testing.T) {
		assert.False(t, No(""))
	})
	t.Run("Space", func(t *testing.T) {
		assert.False(t, No("No Thanks"))
	})
	t.Run("One", func(t *testing.T) {
		assert.False(t, No("1"))
	})
	t.Run("Zero", func(t *testing.T) {
		assert.True(t, No("0"))
	})
	t.Run("HiAccent", func(t *testing.T) {
		assert.True(t, No("ні"))
		assert.True(t, No("НІ"))
	})
	t.Run("Hi", func(t *testing.T) {
		assert.False(t, No("Hi"))
	})
	t.Run("Zadny", func(t *testing.T) {
		assert.True(t, No("žádný"))
		assert.True(t, No("ŽÁDNÝ"))
	})
	t.Run("Nao", func(t *testing.T) {
		assert.True(t, No("não"))
		assert.True(t, No("NÃO"))
	})
	t.Run("Het", func(t *testing.T) {
		assert.True(t, No("нет"))
		assert.True(t, No("НЕТ"))
	})
	t.Run("Ingen", func(t *testing.T) {
		assert.True(t, No("ingen"))
	})
	t.Run("Nee", func(t *testing.T) {
		assert.True(t, No("nee"))
	})
	t.Run("Nein", func(t *testing.T) {
		assert.True(t, No("nein"))
	})
	t.Run("TabSeparatedPhrase", func(t *testing.T) {
		assert.False(t, No("no\tthanks"))
	})
	t.Run("NonBreakingSpace", func(t *testing.T) {
		assert.True(t, No("нет\u00a0"))
	})
}

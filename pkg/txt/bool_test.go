package txt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBool(t *testing.T) {
	t.Run("not empty", func(t *testing.T) {
		assert.True(t, Bool("Browse your life in pictures"))
	})
	t.Run("oui", func(t *testing.T) {
		assert.True(t, Bool("oui"))
	})
	t.Run("non", func(t *testing.T) {
		assert.False(t, Bool("non"))
	})
	t.Run("ja", func(t *testing.T) {
		assert.True(t, Bool("ja"))
	})
	t.Run("true", func(t *testing.T) {
		assert.True(t, Bool("true"))
	})
	t.Run("yes", func(t *testing.T) {
		assert.True(t, Bool("yes"))
	})
	t.Run("no", func(t *testing.T) {
		assert.False(t, Bool("no"))
	})
	t.Run("false", func(t *testing.T) {
		assert.False(t, Bool("false"))
	})
	t.Run("empty", func(t *testing.T) {
		assert.False(t, Bool(""))
	})
}

func TestYes(t *testing.T) {
	t.Run("not empty", func(t *testing.T) {
		assert.False(t, Yes("Browse your life in pictures"))
	})
	t.Run("oui", func(t *testing.T) {
		assert.True(t, Yes("oui"))
		assert.True(t, Yes("OUI"))
	})
	t.Run("non", func(t *testing.T) {
		assert.False(t, Yes("non"))
	})
	t.Run("ja", func(t *testing.T) {
		assert.True(t, Yes("ja"))
	})
	t.Run("true", func(t *testing.T) {
		assert.True(t, Yes("true"))
	})
	t.Run("yes", func(t *testing.T) {
		assert.True(t, Yes("yes"))
	})
	t.Run("no", func(t *testing.T) {
		assert.False(t, Yes("no"))
	})
	t.Run("false", func(t *testing.T) {
		assert.False(t, Yes("false"))
	})
	t.Run("exclude", func(t *testing.T) {
		assert.False(t, Yes("exclude"))
	})
	t.Run("include", func(t *testing.T) {
		assert.True(t, Yes("include"))
	})
	t.Run("unknown", func(t *testing.T) {
		assert.False(t, Yes("unknown"))
	})
	t.Run("please", func(t *testing.T) {
		assert.True(t, Yes("please"))
		assert.True(t, Yes("pLeAsE"))
	})
	t.Run("positive", func(t *testing.T) {
		assert.True(t, Yes("positive"))
	})
	t.Run("empty", func(t *testing.T) {
		assert.False(t, Yes(""))
	})
	t.Run("space", func(t *testing.T) {
		assert.False(t, Yes("Yes Please"))
	})
	t.Run("one", func(t *testing.T) {
		assert.True(t, Yes("1"))
	})
	t.Run("zero", func(t *testing.T) {
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
}

func TestNo(t *testing.T) {
	t.Run("not empty", func(t *testing.T) {
		assert.False(t, No("Browse your life in pictures"))
	})
	t.Run("oui", func(t *testing.T) {
		assert.False(t, No("oui"))
		assert.False(t, No("OUI"))
	})
	t.Run("non", func(t *testing.T) {
		assert.True(t, No("non"))
	})
	t.Run("ja", func(t *testing.T) {
		assert.False(t, No("ja"))
	})
	t.Run("true", func(t *testing.T) {
		assert.False(t, No("true"))
	})
	t.Run("yes", func(t *testing.T) {
		assert.False(t, No("yes"))
	})
	t.Run("no", func(t *testing.T) {
		assert.True(t, No("no"))
	})
	t.Run("false", func(t *testing.T) {
		assert.True(t, No("false"))
	})
	t.Run("exclude", func(t *testing.T) {
		assert.True(t, No("exclude"))
	})
	t.Run("include", func(t *testing.T) {
		assert.False(t, No("include"))
	})
	t.Run("unknown", func(t *testing.T) {
		assert.True(t, No("unknown"))
	})
	t.Run("please", func(t *testing.T) {
		assert.False(t, No("please"))
	})
	t.Run("positive", func(t *testing.T) {
		assert.False(t, No("positive"))
	})
	t.Run("empty", func(t *testing.T) {
		assert.False(t, No(""))
	})
	t.Run("space", func(t *testing.T) {
		assert.False(t, No("No Thanks"))
	})
	t.Run("one", func(t *testing.T) {
		assert.False(t, No("1"))
	})
	t.Run("zero", func(t *testing.T) {
		assert.True(t, No("0"))
	})
	t.Run("hi accent", func(t *testing.T) {
		assert.True(t, No("ні"))
		assert.True(t, No("НІ"))
	})
	t.Run("hi", func(t *testing.T) {
		assert.False(t, No("Hi"))
	})
	t.Run("zadny", func(t *testing.T) {
		assert.True(t, No("žádný"))
		assert.True(t, No("ŽÁDNÝ"))
	})
	t.Run("nao", func(t *testing.T) {
		assert.True(t, No("não"))
		assert.True(t, No("NÃO"))
	})
	t.Run("het", func(t *testing.T) {
		assert.True(t, No("нет"))
		assert.True(t, No("НЕТ"))
	})
	t.Run("ingen", func(t *testing.T) {
		assert.True(t, No("ingen"))
	})
	t.Run("nee", func(t *testing.T) {
		assert.True(t, No("nee"))
	})
	t.Run("nein", func(t *testing.T) {
		assert.True(t, No("nein"))
	})
}

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

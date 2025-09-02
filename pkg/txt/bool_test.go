package txt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBool(t *testing.T) {
	t.Run("not empty", func(t *testing.T) {
		assert.Equal(t, true, Bool("Browse your life in pictures"))
	})
	t.Run("oui", func(t *testing.T) {
		assert.Equal(t, true, Bool("oui"))
	})
	t.Run("non", func(t *testing.T) {
		assert.Equal(t, false, Bool("non"))
	})
	t.Run("ja", func(t *testing.T) {
		assert.Equal(t, true, Bool("ja"))
	})
	t.Run("true", func(t *testing.T) {
		assert.Equal(t, true, Bool("true"))
	})
	t.Run("yes", func(t *testing.T) {
		assert.Equal(t, true, Bool("yes"))
	})
	t.Run("no", func(t *testing.T) {
		assert.Equal(t, false, Bool("no"))
	})
	t.Run("false", func(t *testing.T) {
		assert.Equal(t, false, Bool("false"))
	})
	t.Run("empty", func(t *testing.T) {
		assert.Equal(t, false, Bool(""))
	})
}

func TestYes(t *testing.T) {
	t.Run("not empty", func(t *testing.T) {
		assert.Equal(t, false, Yes("Browse your life in pictures"))
	})
	t.Run("oui", func(t *testing.T) {
		assert.Equal(t, true, Yes("oui"))
	})
	t.Run("non", func(t *testing.T) {
		assert.Equal(t, false, Yes("non"))
	})
	t.Run("ja", func(t *testing.T) {
		assert.Equal(t, true, Yes("ja"))
	})
	t.Run("true", func(t *testing.T) {
		assert.Equal(t, true, Yes("true"))
	})
	t.Run("yes", func(t *testing.T) {
		assert.Equal(t, true, Yes("yes"))
	})
	t.Run("no", func(t *testing.T) {
		assert.Equal(t, false, Yes("no"))
	})
	t.Run("false", func(t *testing.T) {
		assert.Equal(t, false, Yes("false"))
	})
	t.Run("exclude", func(t *testing.T) {
		assert.Equal(t, false, Yes("exclude"))
	})
	t.Run("include", func(t *testing.T) {
		assert.Equal(t, true, Yes("include"))
	})
	t.Run("unknown", func(t *testing.T) {
		assert.Equal(t, false, Yes("unknown"))
	})
	t.Run("please", func(t *testing.T) {
		assert.Equal(t, true, Yes("please"))
	})
	t.Run("positive", func(t *testing.T) {
		assert.Equal(t, true, Yes("positive"))
	})
	t.Run("empty", func(t *testing.T) {
		assert.Equal(t, false, Yes(""))
	})
	t.Run("space", func(t *testing.T) {
		assert.Equal(t, false, Yes("Yes Please"))
	})
	t.Run("one", func(t *testing.T) {
		assert.Equal(t, true, Yes("1"))
	})
	t.Run("zero", func(t *testing.T) {
		assert.Equal(t, false, Yes("0"))
	})
	t.Run("tak", func(t *testing.T) {
		assert.Equal(t, true, Yes("так"))
	})
}

func TestNo(t *testing.T) {
	t.Run("not empty", func(t *testing.T) {
		assert.Equal(t, false, No("Browse your life in pictures"))
	})
	t.Run("oui", func(t *testing.T) {
		assert.Equal(t, false, No("oui"))
	})
	t.Run("non", func(t *testing.T) {
		assert.Equal(t, true, No("non"))
	})
	t.Run("ja", func(t *testing.T) {
		assert.Equal(t, false, No("ja"))
	})
	t.Run("true", func(t *testing.T) {
		assert.Equal(t, false, No("true"))
	})
	t.Run("yes", func(t *testing.T) {
		assert.Equal(t, false, No("yes"))
	})
	t.Run("no", func(t *testing.T) {
		assert.Equal(t, true, No("no"))
	})
	t.Run("false", func(t *testing.T) {
		assert.Equal(t, true, No("false"))
	})
	t.Run("exclude", func(t *testing.T) {
		assert.Equal(t, true, No("exclude"))
	})
	t.Run("include", func(t *testing.T) {
		assert.Equal(t, false, No("include"))
	})
	t.Run("unknown", func(t *testing.T) {
		assert.Equal(t, true, No("unknown"))
	})
	t.Run("please", func(t *testing.T) {
		assert.Equal(t, false, No("please"))
	})
	t.Run("positive", func(t *testing.T) {
		assert.Equal(t, false, No("positive"))
	})
	t.Run("empty", func(t *testing.T) {
		assert.Equal(t, false, No(""))
	})
	t.Run("space", func(t *testing.T) {
		assert.Equal(t, false, No("No Thanks"))
	})
	t.Run("one", func(t *testing.T) {
		assert.Equal(t, false, No("1"))
	})
	t.Run("zero", func(t *testing.T) {
		assert.Equal(t, true, No("0"))
	})
	t.Run("hi accent", func(t *testing.T) {
		assert.Equal(t, true, No("ні"))
	})
	t.Run("hi", func(t *testing.T) {
		assert.Equal(t, false, No("Hi"))
	})
	t.Run("zadny", func(t *testing.T) {
		assert.Equal(t, true, No("žádný"))
	})
	t.Run("nao", func(t *testing.T) {
		assert.Equal(t, true, No("não"))
	})
	t.Run("het", func(t *testing.T) {
		assert.Equal(t, true, No("нет"))
	})
}

func TestNew(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		assert.Equal(t, false, New(""))
	})
	t.Run("EnNew", func(t *testing.T) {
		assert.Equal(t, true, New(EnNew))
	})
	t.Run("Spaces", func(t *testing.T) {
		assert.Equal(t, true, New("     new "))
	})
	t.Run("Uppercase", func(t *testing.T) {
		assert.Equal(t, true, New("NEW"))
	})
	t.Run("Lowercase", func(t *testing.T) {
		assert.Equal(t, true, New("new"))
	})
	t.Run("True", func(t *testing.T) {
		assert.Equal(t, true, New("New"))
	})
	t.Run("False", func(t *testing.T) {
		assert.Equal(t, false, New("non"))
	})
}

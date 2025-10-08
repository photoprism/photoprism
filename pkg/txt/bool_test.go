package txt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBool(t *testing.T) {
	t.Run("NotEmpty", func(t *testing.T) {
		assert.Equal(t, true, Bool("Browse your life in pictures"))
	})
	t.Run("Oui", func(t *testing.T) {
		assert.Equal(t, true, Bool("oui"))
	})
	t.Run("Non", func(t *testing.T) {
		assert.Equal(t, false, Bool("non"))
	})
	t.Run("Ja", func(t *testing.T) {
		assert.Equal(t, true, Bool("ja"))
	})
	t.Run("True", func(t *testing.T) {
		assert.Equal(t, true, Bool("true"))
	})
	t.Run("Yes", func(t *testing.T) {
		assert.Equal(t, true, Bool("yes"))
	})
	t.Run("No", func(t *testing.T) {
		assert.Equal(t, false, Bool("no"))
	})
	t.Run("False", func(t *testing.T) {
		assert.Equal(t, false, Bool("false"))
	})
	t.Run("Empty", func(t *testing.T) {
		assert.Equal(t, false, Bool(""))
	})
}

func TestYes(t *testing.T) {
	t.Run("NotEmpty", func(t *testing.T) {
		assert.Equal(t, false, Yes("Browse your life in pictures"))
	})
	t.Run("Oui", func(t *testing.T) {
		assert.Equal(t, true, Yes("oui"))
	})
	t.Run("Non", func(t *testing.T) {
		assert.Equal(t, false, Yes("non"))
	})
	t.Run("Ja", func(t *testing.T) {
		assert.Equal(t, true, Yes("ja"))
	})
	t.Run("True", func(t *testing.T) {
		assert.Equal(t, true, Yes("true"))
	})
	t.Run("Yes", func(t *testing.T) {
		assert.Equal(t, true, Yes("yes"))
	})
	t.Run("No", func(t *testing.T) {
		assert.Equal(t, false, Yes("no"))
	})
	t.Run("False", func(t *testing.T) {
		assert.Equal(t, false, Yes("false"))
	})
	t.Run("Exclude", func(t *testing.T) {
		assert.Equal(t, false, Yes("exclude"))
	})
	t.Run("Include", func(t *testing.T) {
		assert.Equal(t, true, Yes("include"))
	})
	t.Run("Unknown", func(t *testing.T) {
		assert.Equal(t, false, Yes("unknown"))
	})
	t.Run("Please", func(t *testing.T) {
		assert.Equal(t, true, Yes("please"))
	})
	t.Run("Positive", func(t *testing.T) {
		assert.Equal(t, true, Yes("positive"))
	})
	t.Run("Empty", func(t *testing.T) {
		assert.Equal(t, false, Yes(""))
	})
}

func TestNo(t *testing.T) {
	t.Run("NotEmpty", func(t *testing.T) {
		assert.Equal(t, false, No("Browse your life in pictures"))
	})
	t.Run("Oui", func(t *testing.T) {
		assert.Equal(t, false, No("oui"))
	})
	t.Run("Non", func(t *testing.T) {
		assert.Equal(t, true, No("non"))
	})
	t.Run("Ja", func(t *testing.T) {
		assert.Equal(t, false, No("ja"))
	})
	t.Run("True", func(t *testing.T) {
		assert.Equal(t, false, No("true"))
	})
	t.Run("Yes", func(t *testing.T) {
		assert.Equal(t, false, No("yes"))
	})
	t.Run("No", func(t *testing.T) {
		assert.Equal(t, true, No("no"))
	})
	t.Run("False", func(t *testing.T) {
		assert.Equal(t, true, No("false"))
	})
	t.Run("Exclude", func(t *testing.T) {
		assert.Equal(t, true, No("exclude"))
	})
	t.Run("Include", func(t *testing.T) {
		assert.Equal(t, false, No("include"))
	})
	t.Run("Unknown", func(t *testing.T) {
		assert.Equal(t, true, No("unknown"))
	})
	t.Run("Please", func(t *testing.T) {
		assert.Equal(t, false, No("please"))
	})
	t.Run("Positive", func(t *testing.T) {
		assert.Equal(t, false, No("positive"))
	})
	t.Run("Empty", func(t *testing.T) {
		assert.Equal(t, false, No(""))
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

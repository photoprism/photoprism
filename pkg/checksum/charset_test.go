package checksum

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestChar(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		data := ""
		expected := "a"

		result := string(Char([]byte(data)))

		t.Logf("Char(%s): result %s, expected %s", data, result, expected)

		assert.Equal(t, expected, result)
	})
	t.Run("One", func(t *testing.T) {
		data := "1"
		expected := "R"

		result := string(Char([]byte(data)))

		t.Logf("Char(%s): result %s, expected %s", data, result, expected)

		assert.Equal(t, expected, result)
	})
	t.Run("Serial", func(t *testing.T) {
		data := "zr2g80wvjmm1zwzg"
		expected := "G"

		result := string(Char([]byte(data)))

		t.Logf("Char(%s): result %s, expected %s", data, result, expected)

		assert.Equal(t, expected, result)
	})
	t.Run("HelloWorld", func(t *testing.T) {
		data := "Hello World!"
		expected := "X"

		result := string(Char([]byte(data)))

		t.Logf("Char(%s): result %s, expected %s", data, result, expected)

		assert.Equal(t, expected, result)
	})
}

func TestBase36(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		data := ""
		expected := "a"

		result := string(Base36([]byte(data)))

		t.Logf("Base36(%s): result %s, expected %s", data, result, expected)

		assert.Equal(t, expected, result)
	})
	t.Run("One", func(t *testing.T) {
		data := "1"
		expected := "l"

		result := string(Base36([]byte(data)))

		t.Logf("Base36(%s): result %s, expected %s", data, result, expected)

		assert.Equal(t, expected, result)
	})
	t.Run("Serial", func(t *testing.T) {
		data := "zr2g80wvjmm1zwzg"
		expected := "u"

		result := string(Base36([]byte(data)))

		t.Logf("Base36(%s): result %s, expected %s", data, result, expected)

		assert.Equal(t, expected, result)
	})
	t.Run("HelloWorld", func(t *testing.T) {
		data := "Hello World!"
		expected := "x"

		result := string(Base36([]byte(data)))

		t.Logf("Base36(%s): result %s, expected %s", data, result, expected)

		assert.Equal(t, expected, result)
	})
}

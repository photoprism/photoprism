package checksum

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCrc32(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		data := ""
		expected := "00000000"

		crc := Crc32([]byte(data))
		result := fmt.Sprintf("%08x", crc)

		t.Logf("Crc32(%s): result %s, expected %s", data, result, expected)

		assert.Equal(t, expected, result)
	})
	t.Run("One", func(t *testing.T) {
		data := "1"
		expected := "83dcefb7"

		crc := Crc32([]byte(data))
		result := fmt.Sprintf("%08x", crc)

		t.Logf("Crc32(%s): result %s, expected %s", data, result, expected)

		assert.Equal(t, expected, result)
	})
	t.Run("Serial", func(t *testing.T) {
		data := "zr2g80wvjmm1zwzg"
		expected := "2db41d54"

		crc := Crc32([]byte(data))
		result := fmt.Sprintf("%08x", crc)

		t.Logf("Crc32(%s): result %s, expected %s", data, result, expected)

		assert.Equal(t, expected, result)
	})
	t.Run("HelloWorld", func(t *testing.T) {
		data := "Hello World!"
		expected := "1c291ca3"

		crc := Crc32([]byte(data))
		result := fmt.Sprintf("%08x", crc)

		t.Logf("Crc32(%s): result %s, expected %s", data, result, expected)

		assert.Equal(t, expected, result)
	})
}

func TestSerial(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		data := ""
		expected := "00000000"

		result := Serial([]byte(data))

		t.Logf("Serial(%s): result %s, expected %s", data, result, expected)

		assert.Equal(t, expected, result)
	})
	t.Run("One", func(t *testing.T) {
		data := "1"
		expected := "90f599e3"

		result := Serial([]byte(data))

		t.Logf("Serial(%s): result %s, expected %s", data, result, expected)

		assert.Equal(t, expected, result)
	})
	t.Run("Serial", func(t *testing.T) {
		data := "zr2g80wvjmm1zwzg"
		expected := "c7dcdb1c"

		result := Serial([]byte(data))

		t.Logf("Serial(%s): result %s, expected %s", data, result, expected)

		assert.Equal(t, expected, result)
	})
	t.Run("HelloWorld", func(t *testing.T) {
		data := "Hello World!"
		expected := "fe6cf1dc"

		result := Serial([]byte(data))

		t.Logf("Serial(%s): result %s, expected %s", data, result, expected)

		assert.Equal(t, expected, result)
	})
}

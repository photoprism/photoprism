package checksum

import (
	"crypto/rand"
	"fmt"
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDigit(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		data := ""
		expected := "0"

		result := Digit([]byte(data))

		t.Logf("Digit(%s): result %d, expected %s", data, result, expected)

		assert.Equal(t, expected, fmt.Sprintf("%d", result))
	})
	t.Run("One", func(t *testing.T) {
		data := "1"
		expected := "3"

		result := Digit([]byte(data))

		t.Logf("Digit(%s): result %d, expected %s", data, result, expected)

		assert.Equal(t, expected, fmt.Sprintf("%d", result))
	})
	t.Run("Serial", func(t *testing.T) {
		data := "zr2g80wvjmm1zwzg"
		expected := "8"

		result := Digit([]byte(data))

		t.Logf("Digit(%s): result %d, expected %s", data, result, expected)

		assert.Equal(t, expected, fmt.Sprintf("%d", result))
	})
	t.Run("HelloWorld", func(t *testing.T) {
		data := "Hello World!"
		expected := "5"

		result := Digit([]byte(data))

		t.Logf("Digit(%s): result %d, expected %s", data, result, expected)

		assert.Equal(t, expected, fmt.Sprintf("%d", result))
	})
	t.Run("Rand", func(t *testing.T) {
		for i := 0; i < 100; i++ {
			b := make([]byte, 24)

			if _, err := rand.Read(b); err != nil {
				log.Fatal(err)
			}

			result := Digit(b)

			if result < 0 {
				t.Fatal("result < 0")
			} else if result > 9 {
				t.Fatal("result > 9")
			}
		}
	})
}

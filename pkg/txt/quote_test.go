package txt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQuote(t *testing.T) {
	t.Run("The quick brown fox.", func(t *testing.T) {
		assert.Equal(t, "“The quick brown fox.”", Quote("The quick brown fox."))
	})
	t.Run("filename.txt", func(t *testing.T) {
		assert.Equal(t, "filename.txt", Quote("filename.txt"))
	})
}

package txt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQuote(t *testing.T) {
	t.Run("TheQuickBrownFox", func(t *testing.T) {
		assert.Equal(t, "“The quick brown fox.”", Quote("The quick brown fox."))
	})
	t.Run("FilenameTxt", func(t *testing.T) {
		assert.Equal(t, "filename.txt", Quote("filename.txt"))
	})
	t.Run("EmptyString", func(t *testing.T) {
		assert.Equal(t, "“”", Quote(""))
	})
}

func TestQuoteLower(t *testing.T) {
	t.Run("TheQuickBrownFox", func(t *testing.T) {
		assert.Equal(t, "“the quick brown fox.”", QuoteLower("The quick brown fox."))
	})
	t.Run("FilenameTxt", func(t *testing.T) {
		assert.Equal(t, "filename.txt", QuoteLower("filename.txt"))
	})
	t.Run("EmptyString", func(t *testing.T) {
		assert.Equal(t, "“”", QuoteLower(""))
	})
}

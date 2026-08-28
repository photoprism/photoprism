package face

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBlockEmbeddings(t *testing.T) {
	t.Cleanup(UnblockEmbeddings)

	t.Run("Blocked", func(t *testing.T) {
		BlockEmbeddings("12 marker(s) use facenet, but this instance is configured for sface")

		assert.True(t, EmbeddingsBlocked())
		assert.Contains(t, EmbeddingsBlockedReason(), "configured for sface")
	})
	t.Run("Unblocked", func(t *testing.T) {
		BlockEmbeddings("blocked")
		UnblockEmbeddings()

		assert.False(t, EmbeddingsBlocked())
		assert.Equal(t, "", EmbeddingsBlockedReason())
	})
	t.Run("EmptyReasonDoesNotBlock", func(t *testing.T) {
		BlockEmbeddings("")

		assert.False(t, EmbeddingsBlocked())
	})
}

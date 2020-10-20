package form

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFeedback_Empty(t *testing.T) {
	t.Run("true", func(t *testing.T) {
		feedback := Feedback{}
		assert.True(t, feedback.Empty())
	})
	t.Run("false", func(t *testing.T) {
		feedback := Feedback{Message: "I found a bug", Category: "Bug Report", UserEmail: "test@test.com"}
		assert.False(t, feedback.Empty())
	})
}

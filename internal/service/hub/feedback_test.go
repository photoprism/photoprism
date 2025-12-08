package hub

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/form"
)

func TestNewFeedback(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		feedback := NewFeedback("xxx", "zqkunt22r0bewti9", "test", "test")
		assert.Equal(t, "xxx", feedback.ClientVersion)
		assert.Equal(t, "zqkunt22r0bewti9", feedback.ClientSerial)
	})
}

func TestSendFeedback(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		c := NewConfig("test", "testdata/new.yml", "zqkunt22r0bewti9", "test", "PhotoPrism/Test", "test")

		feedback := Feedback{
			Category:      "Bug Report",
			Subject:       "",
			Message:       "I found a new bug",
			UserName:      "Test User",
			UserEmail:     "test@example.com",
			UserAgent:     "",
			ApiKey:        "123456",
			ClientVersion: "test",
			ClientOS:      "linux",
			ClientArch:    "amd64",
			ClientCPU:     2,
		}

		feedbackForm, formErr := form.NewFeedback(feedback)

		if formErr != nil {
			t.Fatal(formErr)
		}

		sendErr := c.SendFeedback(feedbackForm)
		assert.Error(t, sendErr)

		if Disabled() {
			assert.EqualError(t, sendErr, "unable to send feedback (service disabled)")
		}
	})
}

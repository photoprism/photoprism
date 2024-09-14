package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSendFeedback(t *testing.T) {
	t.Run("PublicMode", func(t *testing.T) {
		app, router, _ := NewApiTest()
		SendFeedback(router)
		r := PerformRequestWithBody(app, "POST", "/api/v1/feedback", `{"Subject": "Send feedback from unit test", "Message": "Test message"}`)
		assert.Equal(t, 403, r.Code)
	})
}

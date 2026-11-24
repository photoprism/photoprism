package hub

import (
	"errors"
	"fmt"
	"net/url"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/form"
)

func restoreBaseURL(t *testing.T) func() {
	t.Helper()

	previous := GetServiceURL("")
	wasDisabled := Disabled()

	return func() {
		if wasDisabled {
			Disable()
			return
		}

		SetBaseURL(previous)
	}
}

func TestGetServiceURL(t *testing.T) {
	cleanup := restoreBaseURL(t)
	t.Cleanup(cleanup)

	SetBaseURL(ProdBaseURL)

	assert.Equal(t, ProdBaseURL, GetServiceURL(""))
	assert.Equal(t, ProdBaseURL+"/demo", GetServiceURL("demo"))

	Disable()

	assert.Empty(t, GetServiceURL("demo"))
}

func TestGetFeedbackServiceURL(t *testing.T) {
	cleanup := restoreBaseURL(t)
	t.Cleanup(cleanup)

	SetBaseURL(ProdBaseURL)

	assert.Empty(t, GetFeedbackServiceURL(""))
	assert.Equal(t, ProdBaseURL+"/demo/feedback", GetFeedbackServiceURL("demo"))

	Disable()

	assert.Empty(t, GetFeedbackServiceURL("demo"))
}

func TestGetServiceHost(t *testing.T) {
	cleanup := restoreBaseURL(t)
	t.Cleanup(cleanup)

	SetBaseURL(ProdBaseURL)

	assert.Equal(t, "my.photoprism.app", GetServiceHost())

	Disable()

	assert.Empty(t, GetServiceHost())
}

func TestSetBaseURLRejectsHTTP(t *testing.T) {
	cleanup := restoreBaseURL(t)
	t.Cleanup(cleanup)

	SetBaseURL(ProdBaseURL)
	SetBaseURL("http://example.com/v1/hello")

	assert.Equal(t, ProdBaseURL, GetServiceURL(""))
}

func TestValidateServiceURL(t *testing.T) {
	t.Run("ValidHttps", func(t *testing.T) {
		assert.NoError(t, ValidateServiceURL("https://my.photoprism.app/v1/hello/demo"))
	})
	t.Run("RejectInsecureScheme", func(t *testing.T) {
		assert.Error(t, ValidateServiceURL("http://my.photoprism.app/v1/hello/demo"))
	})
	t.Run("RejectMissingHost", func(t *testing.T) {
		assert.Error(t, ValidateServiceURL("https:///v1/hello"))
	})
}

func TestApplyTestConfig(t *testing.T) {
	t.Run("DisableByDefault", func(t *testing.T) {
		cleanup := restoreBaseURL(t)
		t.Cleanup(cleanup)

		t.Setenv("PHOTOPRISM_TEST_HUB", "")
		SetBaseURL(ProdBaseURL)

		ApplyTestConfig()

		assert.True(t, Disabled())
	})
	t.Run("EnableTest", func(t *testing.T) {
		cleanup := restoreBaseURL(t)
		t.Cleanup(cleanup)

		t.Setenv("PHOTOPRISM_TEST_HUB", "test")
		Disable()

		ApplyTestConfig()

		assert.False(t, Disabled())
		assert.Equal(t, TestBaseURL, GetServiceURL(""))
	})
	t.Run("EnableProd", func(t *testing.T) {
		cleanup := restoreBaseURL(t)
		t.Cleanup(cleanup)

		t.Setenv("PHOTOPRISM_TEST_HUB", "prod")
		Disable()

		ApplyTestConfig()

		assert.False(t, Disabled())
		assert.Equal(t, ProdBaseURL, GetServiceURL(""))
	})
}

func TestRequestError(t *testing.T) {
	t.Cleanup(restoreBaseURL(t))
	SetBaseURL("https://hub.example.com/v1/hello")

	const key = "b32e9ccdc90eb7c0f6f1b9fbc82b8a2b0e993304"

	t.Run("URLError", func(t *testing.T) {
		cause := errors.New("connection refused")
		err := requestError(&url.Error{Op: "Put", URL: "https://hub.example.com/v1/hello/" + key, Err: cause})
		assert.EqualError(t, err, "request to hub.example.com failed (connection refused)")
		assert.NotContains(t, err.Error(), "/v1/hello")
		assert.ErrorIs(t, err, cause)
	})
	t.Run("URLErrorWithoutCause", func(t *testing.T) {
		err := requestError(&url.Error{Op: "Put", URL: "https://hub.example.com/v1/hello/" + key})
		assert.EqualError(t, err, "request to hub.example.com failed")
	})
	t.Run("Other", func(t *testing.T) {
		other := errors.New("other error")
		assert.Equal(t, other, requestError(other))
	})
	t.Run("Nil", func(t *testing.T) {
		assert.NoError(t, requestError(nil))
	})
	t.Run("TypedNil", func(t *testing.T) {
		var urlErr *url.Error
		err := fmt.Errorf("wrapped: %w", urlErr)
		assert.Equal(t, err, requestError(err))
	})
}

func TestRequestError_Requests(t *testing.T) {
	// A request to an endpoint that refuses connections fails with an error that names the host only.
	t.Cleanup(restoreBaseURL(t))
	SetBaseURL("https://127.0.0.1:1/v1/hello")

	const key = "b32e9ccdc90eb7c0f6f1b9fbc82b8a2b0e993304"
	c := NewConfig("test", filepath.Join(t.TempDir(), "hub.yml"), "zqkunt22r0bewti9", "test", "PhotoPrism/Test", "test")
	c.Key = key

	t.Run("ReSync", func(t *testing.T) {
		err := c.ReSync("")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "request to 127.0.0.1:1 failed")
		assert.NotContains(t, err.Error(), "/v1/hello")
	})
	t.Run("SendFeedback", func(t *testing.T) {
		err := c.SendFeedback(form.Feedback{Category: "help", Message: "test", UserEmail: "test@example.com"})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "request to 127.0.0.1:1 failed")
		assert.NotContains(t, err.Error(), "/v1/hello")
	})
	t.Run("InvalidURL", func(t *testing.T) {
		// An endpoint URL that cannot be parsed is reported without the URL as well.
		SetBaseURL("https://hub example.com/v1/hello")

		err := c.ReSync("")
		require.Error(t, err)
		assert.NotContains(t, err.Error(), "/v1/hello")
		assert.Contains(t, err.Error(), "service request failed")

		err = c.SendFeedback(form.Feedback{Category: "help", Message: "test", UserEmail: "test@example.com"})
		require.Error(t, err)
		assert.NotContains(t, err.Error(), "/v1/hello")
	})
}

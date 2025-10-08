package hub

import (
	"testing"

	"github.com/stretchr/testify/assert"
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

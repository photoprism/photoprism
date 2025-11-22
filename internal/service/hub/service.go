package hub

import (
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/photoprism/photoprism/pkg/clean"
)

// Default service base URLs for testing and production.
const (
	ProdBaseURL = "https://my.photoprism.app/v1/hello"
	TestBaseURL = "https://hub-int.photoprism.app/v1/hello"
)

// baseURL specifies the service endpoint URL.
var baseURL = ProdBaseURL

// GetServiceURL returns the currently configured Hub endpoint, optionally
// appending the provided API key. An empty string is returned when Hub
// requests are disabled.
func GetServiceURL(key string) string {
	if baseURL == "" {
		return ""
	}

	if key == "" {
		return baseURL
	}

	return fmt.Sprintf(baseURL+"/%s", key)
}

// GetFeedbackServiceURL builds the feedback endpoint corresponding to the
// supplied API key. A disabled Hub service results in an empty string.
func GetFeedbackServiceURL(key string) string {
	if key == "" {
		return ""
	}

	u := GetServiceURL(key)

	if u == "" {
		return ""
	}

	return u + "/feedback"
}

// GetServiceHost extracts the Hub host name from the active base URL, or
// returns an empty string when the service is disabled or invalid.
func GetServiceHost() string {
	s := GetServiceURL("")

	if s == "" {
		return ""
	}

	u, err := url.Parse(s)

	if err != nil {
		log.Warn(err)
		return ""
	}

	return u.Host
}

// SetBaseURL updates the Hub endpoint, ignoring inputs that are not HTTPS or
// identical to the current value. Changes are logged so integration tests and
// developers can trace the active target.
func SetBaseURL(u string) {
	// Return if it is not an HTTPS URL.
	if !strings.HasPrefix(u, "https://") {
		return
	}

	// Return if URL has not changed.
	if u == baseURL {
		return
	}

	// Set new service endpoint URL.
	switch u {
	case TestBaseURL:
		log.Debug("config: enabled hub test service endpoint")
	case ProdBaseURL:
		log.Debug("config: enabled hub production service endpoint")
	default:
		log.Debugf("config: changed hub service endpoint to %s", clean.Log(u))
	}

	baseURL = u
}

// Disabled reports whether outbound Hub requests have been switched off.
func Disabled() bool {
	return baseURL == ""
}

// Disable clears the Hub endpoint so no network calls are attempted.
func Disable() {
	// Return if already disabled.
	if Disabled() {
		return
	}

	// Remove configured endpoint URL to disable service.
	baseURL = ""
	log.Debugf("config: disabled hub service requests")
}

// ApplyTestConfig reads PHOTOPRISM_TEST_HUB and switches the Hub endpoint to a
// matching environment ("test", "prod"), disabling requests by default so
// automated tests stay hermetic.
func ApplyTestConfig() {
	switch os.Getenv("PHOTOPRISM_TEST_HUB") {
	case "true", "test", "int":
		SetBaseURL(TestBaseURL)
	case "prod":
		SetBaseURL(ProdBaseURL)
	default:
		Disable()
	}
}

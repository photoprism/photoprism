package vision

import (
	"net/http"
	"sync"

	"github.com/photoprism/photoprism/pkg/clean"
)

// serviceFailures holds the last failure status logged per engine and model until a request succeeds.
var serviceFailures sync.Map

// serviceFailureKey returns the key under which a failure of the engine and model is recorded.
func serviceFailureKey(engine, model string) string {
	return engine + "\x00" + model
}

// warnServiceFailure logs a failed request once per engine, model, and status, and repeats at debug level.
// The hint for a missing model depends on the request format, as the engine may use another API.
func warnServiceFailure(engine string, format ApiFormat, model string, status int) {
	if prev, loaded := serviceFailures.Swap(serviceFailureKey(engine, model), status); loaded && prev == status {
		log.Debugf("vision: %s request for model %s failed again (status %d)", clean.Log(engine), clean.Log(model), status)
		return
	}

	switch {
	case status == http.StatusNotFound && format == ApiFormatOpenAI:
		log.Warnf("vision: %s model %s is unavailable (status %d), check the model name and the service uri", clean.Log(engine), clean.Log(model), status)
	case status == http.StatusNotFound || status == http.StatusGone:
		log.Warnf("vision: %s model %s is unavailable (status %d), it may have been retired or renamed", clean.Log(engine), clean.Log(model), status)
	default:
		log.Warnf("vision: %s request for model %s failed (status %d)", clean.Log(engine), clean.Log(model), status)
	}
}

// clearServiceFailure re-arms the failure warning for the engine and model after a request succeeded.
func clearServiceFailure(engine, model string) {
	serviceFailures.Delete(serviceFailureKey(engine, model))
}

// serviceStatusError returns the error for a response with a status of 300 or more, warning from 400 once
// per model and status. A status below 300 re-arms the warning; a redirect keeps its state.
func serviceStatusError(engine string, format ApiFormat, model string, status int) error {
	if status < http.StatusMultipleChoices {
		clearServiceFailure(engine, model)
		return nil
	}

	if status >= http.StatusBadRequest {
		warnServiceFailure(engine, format, model, status)
	}

	return serviceError(format, status)
}

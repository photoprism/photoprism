package onnx

import (
	"os"
)

// TelemetryEnv is the environment variable ONNX Runtime reads when it creates its environment.
const TelemetryEnv = "ORT_DISABLE_TELEMETRY"

// init disables ONNX Runtime telemetry before any native code runs, as Go then updates the
// C environment without racing threads that might read it.
func init() {
	_ = disableTelemetryEnv()
}

// disableTelemetryEnv sets ORT_DISABLE_TELEMETRY to 1, replacing any other value, so the runtime
// never starts its uploader or writes a device identifier.
func disableTelemetryEnv() error {
	if os.Getenv(TelemetryEnv) == "1" {
		return nil
	}

	return os.Setenv(TelemetryEnv, "1")
}

package onnx

import (
	"errors"
	"fmt"
	"path/filepath"

	onnxruntime "github.com/yalue/onnxruntime_go"

	"github.com/photoprism/photoprism/pkg/clean"
)

// SessionConfig carries the options a model is loaded with, together with the execution
// provider actually in force. Create one per model and release it with Destroy.
type SessionConfig struct {
	Options  *onnxruntime.SessionOptions
	Provider Provider
	settings SessionSettings
}

// NewSessionConfig builds the session options for the requested provider, falling back to the
// CPU with one warning when it cannot be applied.
func NewSessionConfig(settings SessionSettings) (*SessionConfig, error) {
	opts, provider, err := NewSessionOptions(settings)

	if err != nil {
		return nil, err
	}

	return &SessionConfig{Options: opts, Provider: provider, settings: settings}, nil
}

// Destroy releases the session options. Sessions already created from them keep working and
// are closed by their own owners.
func (c *SessionConfig) Destroy() {
	if c == nil {
		return
	}

	DestroySessionOptions(c.Options)
	c.Options = nil
}

// NewSession creates an inference session and verifies a GPU one with a single warm-up
// inference on a zero tensor of the given geometry.
//
// An incomplete CUDA installation builds provider options and a session successfully and only
// fails at the first node, so a load-time check cannot see it. Verifying here costs one
// inference per model instead of failing every indexed photo, and the CUDA path pays that
// first inference anyway. A failed warm-up reloads the model CPU-only and updates the config,
// so later sessions built from it no longer ask for the provider.
func (c *SessionConfig) NewSession(modelPath string, inputNames, outputNames []string, inputShape []int64) (*onnxruntime.DynamicAdvancedSession, error) {
	session, err := onnxruntime.NewDynamicAdvancedSession(modelPath, inputNames, outputNames, c.Options)

	if err != nil {
		return nil, err
	}

	if c.Provider == ProviderCPU {
		return session, nil
	}

	if err = warmUpSession(session, inputShape, len(outputNames)); err == nil {
		return session, nil
	}

	log.Warnf("onnx: %s execution provider cannot run %s (%s), loading it on the %s",
		c.Provider, clean.Log(filepath.Base(modelPath)), err, ProviderCPU)

	DestroySession(session)

	cpuSettings := c.settings
	cpuSettings.Provider = ProviderCPU

	cpuOpts, err := newBaseSessionOptions(cpuSettings)

	if err != nil {
		return nil, err
	}

	// Only the provider differs from what was asked for; the thread counts are kept.
	DestroySessionOptions(c.Options)
	c.Options = cpuOpts
	c.Provider = ProviderCPU

	return onnxruntime.NewDynamicAdvancedSession(modelPath, inputNames, outputNames, c.Options)
}

// warmUpSession runs one inference on a zero tensor of the model's input geometry, so that a
// provider which loads but cannot execute is reported before the session is used.
func warmUpSession(session *onnxruntime.DynamicAdvancedSession, inputShape []int64, outputs int) error {
	if session == nil {
		return errors.New("no session to verify")
	}

	if len(inputShape) == 0 {
		return errors.New("no input geometry to verify")
	}

	for _, dim := range inputShape {
		if dim <= 0 {
			return fmt.Errorf("input geometry %v is not fixed", inputShape)
		}
	}

	input, err := onnxruntime.NewEmptyTensor[float32](onnxruntime.NewShape(inputShape...))

	if err != nil {
		return fmt.Errorf("create warm-up tensor: %w", err)
	}

	defer DestroyValue(input)

	// A nil output is allocated by the runtime and must be released here.
	outputValues := make([]onnxruntime.Value, outputs)

	defer func() {
		for _, value := range outputValues {
			DestroyValue(value)
		}
	}()

	return session.Run([]onnxruntime.Value{input}, outputValues)
}

// DestroySession closes an inference session, reporting a failure at debug level only, since it
// cannot affect a result the caller already holds.
func DestroySession(session *onnxruntime.DynamicAdvancedSession) {
	if session == nil {
		return
	}

	if err := session.Destroy(); err != nil {
		log.Debugf("onnx: %s (destroy session)", err)
	}
}

// DestroyValue releases a tensor or other session value, ignoring one that was never allocated.
func DestroyValue(value onnxruntime.Value) {
	if value == nil {
		return
	}

	if err := value.Destroy(); err != nil {
		log.Debugf("onnx: %s (destroy value)", err)
	}
}

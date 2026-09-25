package onnx

import (
	"errors"
	"fmt"
	"path/filepath"

	onnxruntime "github.com/yalue/onnxruntime_go"

	"github.com/photoprism/photoprism/pkg/clean"
)

// ErrGeometryUnknown reports that a session could not be verified because the caller has no
// resolved input geometry. It is not a provider failure and must not trigger a fall back.
var ErrGeometryUnknown = errors.New("input geometry is not resolved")

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

// WithFallback runs a load step against the options in force and, when it fails while a GPU
// provider is applied, reloads CPU-only and runs the step once more.
//
// Every step that opens a session pays the provider's cost and can fail for the provider's
// reasons, so the metadata read goes through here as well as the inference session. Without
// that, a GPU too busy to open a session for two seconds fails the whole model load, and the
// caller drops the model it already had rather than continuing on the CPU.
func (c *SessionConfig) WithFallback(modelPath string, load func(opts *onnxruntime.SessionOptions) error) error {
	err := load(c.Options)

	if err == nil || c.Provider == ProviderCPU {
		return err
	}

	if !c.fallBackToCPU(modelPath, err) {
		return err
	}

	return load(c.Options)
}

// fallBackToCPU rebuilds the options CPU-only, keeping the thread counts, and reports whether
// the caller may retry. It is the single place that announces losing a provider.
func (c *SessionConfig) fallBackToCPU(modelPath string, reason error) bool {
	cpuSettings := c.settings
	cpuSettings.Provider = ProviderCPU

	cpuOpts, err := newBaseSessionOptions(cpuSettings)

	if err != nil {
		return false
	}

	log.Warnf("onnx: %s execution provider cannot load %s (%s), using the %s instead",
		c.Provider, clean.Log(filepath.Base(modelPath)), providerError(reason), ProviderCPU)

	DestroySessionOptions(c.Options)
	c.Options = cpuOpts
	c.Provider = ProviderCPU

	return true
}

// NewSession creates an inference session and verifies a GPU one with a single warm-up
// inference on a zero tensor of the given geometry.
//
// An incomplete CUDA installation builds provider options and a session successfully and only
// fails at the first node, so a load-time check cannot see it. Verifying here costs one
// inference per model instead of failing every indexed photo, and the CUDA path pays that
// first inference anyway. Geometry the caller could not resolve is reported at debug level and
// leaves the session in place, because an unverifiable model is not evidence against the GPU.
func (c *SessionConfig) NewSession(modelPath string, inputNames, outputNames []string, inputShape []int64) (*onnxruntime.DynamicAdvancedSession, error) {
	var session *onnxruntime.DynamicAdvancedSession

	err := c.WithFallback(modelPath, func(opts *onnxruntime.SessionOptions) error {
		created, createErr := onnxruntime.NewDynamicAdvancedSession(modelPath, inputNames, outputNames, opts)

		if createErr != nil {
			return createErr
		}

		if c.Provider != ProviderCPU {
			switch warmUpErr := warmUpSession(created, inputShape, len(outputNames)); {
			case errors.Is(warmUpErr, ErrGeometryUnknown):
				log.Debugf("onnx: %s session for %s was not verified (%s)",
					c.Provider, clean.Log(filepath.Base(modelPath)), warmUpErr)
			case warmUpErr != nil:
				DestroySession(created)
				return warmUpErr
			}
		}

		session = created

		return nil
	})

	if err != nil {
		return nil, err
	}

	return session, nil
}

// warmUpSession runs one inference on a zero tensor of the model's input geometry, so that a
// provider which loads but cannot execute is reported before the session is used.
func warmUpSession(session *onnxruntime.DynamicAdvancedSession, inputShape []int64, outputs int) error {
	if session == nil {
		return errors.New("no session to verify")
	}

	if len(inputShape) == 0 {
		return fmt.Errorf("%w: no geometry given", ErrGeometryUnknown)
	}

	for _, dim := range inputShape {
		if dim <= 0 {
			return fmt.Errorf("%w: %v", ErrGeometryUnknown, inputShape)
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

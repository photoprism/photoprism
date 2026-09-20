package onnx

import (
	"fmt"
	"strings"

	onnxruntime "github.com/yalue/onnxruntime_go"
)

// Provider identifies the ONNX Runtime execution provider that runs a session's graph.
type Provider string

const (
	// ProviderCPU runs inference on the CPU and is available in every runtime build.
	ProviderCPU Provider = "cpu"
	// ProviderCUDA runs inference on an NVIDIA GPU through the CUDA Execution Provider.
	ProviderCUDA Provider = "cuda"
)

// DefaultProvider is the execution provider used when none is configured.
const DefaultProvider = ProviderCPU

// Providers lists the execution providers that may be configured.
var Providers = []Provider{ProviderCPU, ProviderCUDA}

// appendProviderVar applies a non-default execution provider, and is a variable so that the
// fall back can be tested on a machine where the provider does work.
var appendProviderVar = appendCUDAProvider

// String returns the provider name as it is written in the configuration and in logs.
func (p Provider) String() string {
	return string(p)
}

// ProviderUsageString lists the configurable execution providers for CLI help, built from the
// registered list so a new provider cannot be offered in one place and missing from the other.
func ProviderUsageString() string {
	names := make([]string, 0, len(Providers))

	for _, p := range Providers {
		names = append(names, p.String())
	}

	return strings.Join(names, ", ")
}

// ParseProvider resolves a configured value to a supported provider and reports whether the
// value was recognized. An empty value is recognized, since that is the default rather than a
// mistake; anything else unknown resolves to the default so an unusable setting cannot stop
// inference.
func ParseProvider(s string) (Provider, bool) {
	switch Provider(strings.ToLower(strings.TrimSpace(s))) {
	case "":
		return DefaultProvider, true
	case ProviderCPU:
		return ProviderCPU, true
	case ProviderCUDA:
		return ProviderCUDA, true
	default:
		return DefaultProvider, false
	}
}

// SessionSettings describe how an inference session is built: the threads it may use, and the
// execution provider it should run on. A thread count of zero leaves the runtime's own default.
type SessionSettings struct {
	Provider       Provider
	IntraOpThreads int
	InterOpThreads int
}

// NewSessionOptions builds session options for the requested provider and reports the provider
// that was actually applied, so a caller can log which one a model loaded with.
//
// A provider that cannot be applied is not an error: one warning names the reason and CPU-only
// options are returned, because an instance that loses its GPU should keep indexing. Callers
// must Destroy the options they receive.
func NewSessionOptions(settings SessionSettings) (*onnxruntime.SessionOptions, Provider, error) {
	opts, err := newBaseSessionOptions(settings)

	if err != nil {
		return nil, DefaultProvider, err
	}

	if settings.Provider != ProviderCUDA {
		return opts, ProviderCPU, nil
	}

	if err = appendProviderVar(opts); err == nil {
		return opts, ProviderCUDA, nil
	}

	// The runtime prints its own error line for a missing device before returning this error,
	// so the warning has to say plainly that the fall back is not fatal.
	log.Warnf("onnx: %s execution provider is unavailable (%s), running inference on the %s", ProviderCUDA, err, ProviderCPU)

	// Options that failed to take a provider are discarded rather than reused, so what the
	// caller receives was built for the CPU from the start.
	DestroySessionOptions(opts)

	if opts, err = newBaseSessionOptions(settings); err != nil {
		return nil, DefaultProvider, err
	}

	return opts, ProviderCPU, nil
}

// newBaseSessionOptions builds the session options every provider shares. Thread settings are
// applied in both cases, because they are what the CPU fall back runs on and the CUDA provider
// uses CPU kernels for any operator it does not implement.
func newBaseSessionOptions(settings SessionSettings) (*onnxruntime.SessionOptions, error) {
	opts, err := onnxruntime.NewSessionOptions()

	if err != nil {
		return nil, fmt.Errorf("create session options: %w", err)
	}

	if settings.IntraOpThreads > 0 {
		if err = opts.SetIntraOpNumThreads(settings.IntraOpThreads); err != nil {
			DestroySessionOptions(opts)
			return nil, fmt.Errorf("configure intra-op threads: %w", err)
		}
	}

	if settings.InterOpThreads > 0 {
		if err = opts.SetInterOpNumThreads(settings.InterOpThreads); err != nil {
			DestroySessionOptions(opts)
			return nil, fmt.Errorf("configure inter-op threads: %w", err)
		}
	}

	if err = opts.SetGraphOptimizationLevel(onnxruntime.GraphOptimizationLevelEnableAll); err != nil {
		DestroySessionOptions(opts)
		return nil, fmt.Errorf("optimize session graph: %w", err)
	}

	return opts, nil
}

// appendCUDAProvider adds the CUDA execution provider to the given session options.
//
// The binding offers no discovery call, so availability is established by building the provider
// options and treating any error as unavailable. Updating them is the step that loads the
// provider library, so a missing library and an invisible device both surface here.
func appendCUDAProvider(opts *onnxruntime.SessionOptions) error {
	cudaOpts, err := onnxruntime.NewCUDAProviderOptions()

	if err != nil {
		return err
	}

	defer func() {
		if destroyErr := cudaOpts.Destroy(); destroyErr != nil {
			log.Debugf("onnx: %s (destroy cuda provider options)", destroyErr)
		}
	}()

	if err = cudaOpts.Update(map[string]string{"device_id": "0"}); err != nil {
		return err
	}

	return opts.AppendExecutionProviderCUDA(cudaOpts)
}

// DestroySessionOptions releases session options, reporting a failure at debug level only,
// since it cannot affect the result a caller already holds.
func DestroySessionOptions(opts *onnxruntime.SessionOptions) {
	if opts == nil {
		return
	}

	if err := opts.Destroy(); err != nil {
		log.Debugf("onnx: %s (destroy session options)", err)
	}
}

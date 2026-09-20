package onnx

import (
	"fmt"
	"regexp"
	"strings"

	onnxruntime "github.com/yalue/onnxruntime_go"

	"github.com/photoprism/photoprism/pkg/txt"
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

// providerErrorLen bounds a runtime error in a log line.
const providerErrorLen = 200

var (
	// sourceLocation matches a C/C++ source reference such as "/src/foo/bar.cc:62".
	sourceLocation = regexp.MustCompile(`\S*\.(?:cc|h|cpp|hpp):\d+\s*`)
	// errorMarker matches the runtime's own status marker, e.g. "[ONNXRuntimeError] : 1 : FAIL : ".
	errorMarker = regexp.MustCompile(`\[ONNXRuntimeError\]\s*:\s*\d+\s*:\s*[A-Z_]+\s*:\s*`)
)

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

	switch settings.Provider {
	case "", ProviderCPU:
		// An unset provider is the default, not an unimplemented one.
		return opts, ProviderCPU, nil
	case ProviderCUDA:
		if err = appendProviderVar(opts); err == nil {
			return opts, ProviderCUDA, nil
		}

		// The runtime prints its own error line for a missing device before returning this
		// error, so the warning has to say plainly that the fall back is not fatal.
		log.Warnf("onnx: %s execution provider is unavailable (%s), running inference on the %s",
			ProviderCUDA, providerError(err), ProviderCPU)
	default:
		// A provider that parses but has no implementation here would otherwise run on the CPU
		// with a log line that reads as though it had been honored.
		log.Warnf("onnx: %s execution provider is not implemented, running inference on the %s",
			settings.Provider, ProviderCPU)

		return opts, ProviderCPU, nil
	}

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
//
// TF32 is switched off because the runtime enables it by default on Ampere and later, which
// rounds the mantissa and moves an embedding far more than kernel order does. Embeddings are
// persisted and compared by distance, so both providers must compute the same graph in FP32.
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

	if err = cudaOpts.Update(cudaProviderOptions()); err != nil {
		return err
	}

	return opts.AppendExecutionProviderCUDA(cudaOpts)
}

// cudaProviderOptions returns the options the CUDA provider is configured with. It is separate
// so that the values guarding persisted embeddings can be asserted without a device.
func cudaProviderOptions() map[string]string {
	return map[string]string{
		"device_id": "0",
		// Off because the runtime enables it by default on Ampere and later; see above.
		"use_tf32": "0",
	}
}

// providerError renders a runtime error for a log line.
//
// An ONNX Runtime error wraps the one sentence an operator acts on in a build path, a mangled
// C++ signature, and a tail naming the host - so simply shortening it keeps the noise and drops
// the reason. The span from the first source location through the error marker is removed, the
// tail with it, and only then is the remainder bounded.
func providerError(err error) string {
	if err == nil {
		return "no error"
	}

	s := err.Error()

	for _, tail := range []string{" ; GPU=", " ; hostname=", " ; file="} {
		if i := strings.Index(s, tail); i > 0 {
			s = s[:i]
		}
	}

	if loc := sourceLocation.FindStringIndex(s); loc != nil {
		if marker := errorMarker.FindStringIndex(s[loc[0]:]); marker != nil {
			s = s[:loc[0]] + s[loc[0]+marker[1]:]
		}
	}

	s = sourceLocation.ReplaceAllString(s, "")

	return txt.Shorten(strings.Join(strings.Fields(s), " "), providerErrorLen, txt.Ellipsis)
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

package onnx

import (
	"errors"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	onnxruntime "github.com/yalue/onnxruntime_go"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// captureProviderLog redirects the package logger for the duration of the test and returns its
// entries.
func captureProviderLog(t *testing.T) *test.Hook {
	t.Helper()

	orig := log
	logger, hook := test.NewNullLogger()
	logger.SetLevel(logrus.TraceLevel)
	log = logger

	t.Cleanup(func() { log = orig })

	return hook
}

// stubAppendProvider replaces the provider append for the duration of the test and clears the
// remembered failures before and after, so no test inherits another's outcome.
func stubAppendProvider(t *testing.T, fn func(*onnxruntime.SessionOptions) error) {
	t.Helper()

	orig := appendProviderVar
	resetProviderFailures()
	appendProviderVar = fn

	t.Cleanup(func() {
		appendProviderVar = orig
		resetProviderFailures()
	})
}

// resetProviderFailures forgets every provider failure remembered by this process.
func resetProviderFailures() {
	providerMu.Lock()
	defer providerMu.Unlock()

	clear(providerFailures)
}

// requireSessionRuntime skips a test when the ONNX Runtime cannot be loaded, which is the case
// in build environments that did not run "make dep".
func requireSessionRuntime(t *testing.T) {
	t.Helper()

	if err := EnsureRuntime(""); err != nil {
		t.Skipf("onnx: skipping, %s", err)
	}
}

func TestProvider_String(t *testing.T) {
	t.Run("CPU", func(t *testing.T) {
		assert.Equal(t, "cpu", ProviderCPU.String())
	})
	t.Run("CUDA", func(t *testing.T) {
		assert.Equal(t, "cuda", ProviderCUDA.String())
	})
	t.Run("Default", func(t *testing.T) {
		assert.Equal(t, ProviderCPU, DefaultProvider)
	})
}

func TestProviderUsageString(t *testing.T) {
	t.Run("ListsEveryProvider", func(t *testing.T) {
		usage := ProviderUsageString()
		assert.Equal(t, "cpu, cuda", usage)

		// Built from the registered list, so help text cannot drift from what parses.
		for _, p := range Providers {
			assert.Contains(t, usage, p.String())

			_, ok := ParseProvider(p.String())
			assert.True(t, ok)
		}
	})
}

func TestParseProvider(t *testing.T) {
	t.Run("CPU", func(t *testing.T) {
		provider, ok := ParseProvider("cpu")
		assert.True(t, ok)
		assert.Equal(t, ProviderCPU, provider)
	})
	t.Run("CUDA", func(t *testing.T) {
		provider, ok := ParseProvider("cuda")
		assert.True(t, ok)
		assert.Equal(t, ProviderCUDA, provider)
	})
	t.Run("Empty", func(t *testing.T) {
		// An unset option is the default, not a mistake, so it must not be reported as one.
		provider, ok := ParseProvider("")
		assert.True(t, ok)
		assert.Equal(t, ProviderCPU, provider)
	})
	t.Run("Whitespace", func(t *testing.T) {
		provider, ok := ParseProvider("  CUDA  ")
		assert.True(t, ok)
		assert.Equal(t, ProviderCUDA, provider)
	})
	t.Run("Unknown", func(t *testing.T) {
		provider, ok := ParseProvider("rocm")
		assert.False(t, ok)
		assert.Equal(t, ProviderCPU, provider)
	})
}

func TestNewSessionOptions(t *testing.T) {
	t.Run("CPU", func(t *testing.T) {
		requireSessionRuntime(t)
		hook := captureProviderLog(t)

		opts, applied, err := NewSessionOptions(SessionSettings{Provider: ProviderCPU, IntraOpThreads: 2, InterOpThreads: 1})
		require.NoError(t, err)
		require.NotNil(t, opts)
		t.Cleanup(func() { DestroySessionOptions(opts) })
		assert.Equal(t, ProviderCPU, applied)
		assert.Empty(t, hook.AllEntries())
	})
	t.Run("DefaultThreads", func(t *testing.T) {
		// Zero leaves the runtime's own defaults, so an unconfigured caller still gets options.
		requireSessionRuntime(t)

		opts, applied, err := NewSessionOptions(SessionSettings{})
		require.NoError(t, err)
		require.NotNil(t, opts)
		t.Cleanup(func() { DestroySessionOptions(opts) })
		assert.Equal(t, ProviderCPU, applied)
	})
	t.Run("CUDAFallsBack", func(t *testing.T) {
		// Proves the fall back on every machine, including one where the provider works.
		requireSessionRuntime(t)
		hook := captureProviderLog(t)

		calls := 0
		stubAppendProvider(t, func(*onnxruntime.SessionOptions) error {
			calls++
			return errors.New("no CUDA-capable device is detected")
		})

		for range 2 {
			opts, applied, err := NewSessionOptions(SessionSettings{Provider: ProviderCUDA, IntraOpThreads: 2, InterOpThreads: 1})
			require.NoError(t, err)
			require.NotNil(t, opts)
			t.Cleanup(func() { DestroySessionOptions(opts) })
			assert.Equal(t, ProviderCPU, applied)
		}

		// Each model load calls this, so the failure is attempted and announced once per process.
		assert.Equal(t, 1, calls)

		var warnings []*logrus.Entry

		for _, entry := range hook.AllEntries() {
			if entry.Level == logrus.WarnLevel {
				warnings = append(warnings, entry)
			}
		}

		require.Len(t, warnings, 1)
		assert.Contains(t, warnings[0].Message, "cuda")
		assert.Contains(t, warnings[0].Message, "no CUDA-capable device is detected")
		// The runtime prints its own red error line first, so ours must say it is not fatal.
		assert.Contains(t, warnings[0].Message, "running inference on the cpu")

		opts, applied, err := NewSessionOptions(SessionSettings{Provider: ProviderCPU, IntraOpThreads: 1})
		require.NoError(t, err)
		t.Cleanup(func() { DestroySessionOptions(opts) })
		assert.Equal(t, ProviderCPU, applied)
	})
	t.Run("UnimplementedProviderWarns", func(t *testing.T) {
		// Adding a value to Providers without an implementation must not look like success.
		requireSessionRuntime(t)
		hook := captureProviderLog(t)

		opts, applied, err := NewSessionOptions(SessionSettings{Provider: Provider("coreml"), IntraOpThreads: 1})
		require.NoError(t, err)
		require.NotNil(t, opts)
		t.Cleanup(func() { DestroySessionOptions(opts) })
		assert.Equal(t, ProviderCPU, applied)

		entry := hook.LastEntry()
		require.NotNil(t, entry)
		assert.Equal(t, logrus.WarnLevel, entry.Level)
		assert.Contains(t, entry.Message, "coreml")
		assert.Contains(t, entry.Message, "not implemented")
	})
	t.Run("UnsetProviderIsDefault", func(t *testing.T) {
		// The zero value is an unset option, not an unimplemented provider.
		requireSessionRuntime(t)
		hook := captureProviderLog(t)

		opts, applied, err := NewSessionOptions(SessionSettings{IntraOpThreads: 1})
		require.NoError(t, err)
		t.Cleanup(func() { DestroySessionOptions(opts) })
		assert.Equal(t, ProviderCPU, applied)
		assert.Empty(t, hook.AllEntries())
	})
	t.Run("CUDAApplied", func(t *testing.T) {
		requireSessionRuntime(t)

		calls := 0
		stubAppendProvider(t, func(*onnxruntime.SessionOptions) error {
			calls++
			return nil
		})

		// A provider that was applied is applied again for every model.
		for range 2 {
			opts, applied, err := NewSessionOptions(SessionSettings{Provider: ProviderCUDA, IntraOpThreads: 2})
			require.NoError(t, err)
			require.NotNil(t, opts)
			t.Cleanup(func() { DestroySessionOptions(opts) })
			assert.Equal(t, ProviderCUDA, applied)
		}

		assert.Equal(t, 2, calls)
	})
}

func TestAppendProvider(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		hook := captureProviderLog(t)

		calls := 0
		stubAppendProvider(t, func(*onnxruntime.SessionOptions) error {
			calls++
			return nil
		})

		assert.NoError(t, appendProvider(ProviderCUDA, nil))
		assert.NoError(t, appendProvider(ProviderCUDA, nil))
		assert.Equal(t, 2, calls)
		assert.Empty(t, hook.AllEntries())
	})
	t.Run("FailureRemembered", func(t *testing.T) {
		hook := captureProviderLog(t)

		calls := 0
		stubAppendProvider(t, func(*onnxruntime.SessionOptions) error {
			calls++
			return errors.New("libcuda.so.1: cannot open shared object file")
		})

		first := appendProvider(ProviderCUDA, nil)
		require.Error(t, first)
		second := appendProvider(ProviderCUDA, nil)
		assert.Equal(t, first, second)
		assert.Equal(t, 1, calls)

		entries := hook.AllEntries()
		require.Len(t, entries, 2)
		assert.Equal(t, logrus.WarnLevel, entries[0].Level)
		assert.Contains(t, entries[0].Message, "libcuda.so.1")
		assert.Equal(t, logrus.DebugLevel, entries[1].Level)
	})
	t.Run("ConcurrentFirstCalls", func(t *testing.T) {
		// Models load concurrently, so the first failure must still be attempted and warned once.
		hook := captureProviderLog(t)

		var calls atomic.Int32
		stubAppendProvider(t, func(*onnxruntime.SessionOptions) error {
			calls.Add(1)
			time.Sleep(5 * time.Millisecond)
			return errors.New("unavailable")
		})

		var wg sync.WaitGroup

		for range 8 {
			wg.Go(func() { assert.Error(t, appendProvider(ProviderCUDA, nil)) })
		}

		wg.Wait()

		warnings := 0

		for _, entry := range hook.AllEntries() {
			if entry.Level == logrus.WarnLevel {
				warnings++
			}
		}

		assert.Equal(t, int32(1), calls.Load())
		assert.Equal(t, 1, warnings)
	})
	t.Run("PerProvider", func(t *testing.T) {
		captureProviderLog(t)
		stubAppendProvider(t, func(*onnxruntime.SessionOptions) error { return errors.New("unavailable") })

		require.Error(t, appendProvider(ProviderCUDA, nil))

		// A failure is remembered for the provider that failed, not for every provider.
		appendProviderVar = func(*onnxruntime.SessionOptions) error { return nil }
		assert.NoError(t, appendProvider(Provider("coreml"), nil))
		assert.Error(t, appendProvider(ProviderCUDA, nil))
	})
}

func TestNewBaseSessionOptions(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		requireSessionRuntime(t)

		opts, err := newBaseSessionOptions(SessionSettings{IntraOpThreads: 2, InterOpThreads: 1})
		require.NoError(t, err)
		require.NotNil(t, opts)
		DestroySessionOptions(opts)
	})
	t.Run("InvalidThreads", func(t *testing.T) {
		// A negative count is left to the runtime rather than passed on, so no error is
		// expected here; the options must still be usable.
		requireSessionRuntime(t)

		opts, err := newBaseSessionOptions(SessionSettings{IntraOpThreads: -4, InterOpThreads: -1})
		require.NoError(t, err)
		require.NotNil(t, opts)
		DestroySessionOptions(opts)
	})
}

func TestCUDAProviderOptions(t *testing.T) {
	t.Run("DisablesTF32", func(t *testing.T) {
		// The runtime enables TF32 by default on Ampere and later, which moves a persisted
		// embedding three orders of magnitude further than kernel order does. Asserted on the
		// map rather than through a session, so the guard is pinned on a machine with no GPU.
		assert.Equal(t, "0", cudaProviderOptions()["use_tf32"])
	})
	t.Run("SelectsFirstDevice", func(t *testing.T) {
		assert.Equal(t, "0", cudaProviderOptions()["device_id"])
	})
}

func TestAppendCUDAProvider(t *testing.T) {
	t.Run("ReportsAvailability", func(t *testing.T) {
		// Both outcomes are valid: a CPU-only build or an invisible device reports the provider
		// as unavailable, and a GPU build with a device appends it. Each branch is asserted, so
		// the test carries weight on either kind of machine.
		requireSessionRuntime(t)

		opts, err := newBaseSessionOptions(SessionSettings{IntraOpThreads: 1})
		require.NoError(t, err)
		t.Cleanup(func() { DestroySessionOptions(opts) })

		if err = appendCUDAProvider(opts); err != nil {
			// The reason has to name the provider, or the fall-back warning is unreadable.
			t.Logf("cuda provider unavailable: %s", err)
			assert.Contains(t, strings.ToLower(err.Error()), "cuda")

			return
		}

		// The options took the provider, so they must still build a usable session.
		inputName, outputName := embeddingGraph(t)
		session, err := onnxruntime.NewDynamicAdvancedSession(embeddingModelPath,
			[]string{inputName}, []string{outputName}, opts)
		require.NoError(t, err)
		require.NotNil(t, session)
		DestroySession(session)
	})
}

func TestProviderError(t *testing.T) {
	t.Run("Nil", func(t *testing.T) {
		assert.Equal(t, "no error", providerError(nil))
	})
	t.Run("DropsDiagnosticTail", func(t *testing.T) {
		// The runtime appends its build paths and the host name, which must not reach a log.
		err := errors.New("CUDA failure 100: no CUDA-capable device is detected ; GPU=-1 ; hostname=abc123 ; file=/onnxruntime_src/x.cc")
		out := providerError(err)
		assert.Contains(t, out, "no CUDA-capable device is detected")
		assert.NotContains(t, out, "hostname")
		assert.NotContains(t, out, "onnxruntime_src")
	})
	t.Run("KeepsTheReasonNotThePrefix", func(t *testing.T) {
		// Verbatim from the runtime when no device is visible. The reason sits at the far end,
		// so shortening from the front would keep only the build path and the C++ signature.
		err := errors.New(`/onnxruntime_src/onnxruntime/core/providers/cuda/cuda_execution_provider_info.cc:62 ` +
			`static onnxruntime::CUDAExecutionProviderInfo onnxruntime::CUDAExecutionProviderInfo::FromProviderOptions` +
			`(const onnxruntime::ProviderOptions&) [ONNXRuntimeError] : 1 : FAIL : provider_options_utils.h:187 Parse ` +
			`Failed to parse provider option "device_id": CUDA failure 100: no CUDA-capable device is detected ; ` +
			`GPU=-1 ; hostname=65a6e914d819 ; file=/onnxruntime_src/x.cc ; line=69 ; expr=cudaGetDeviceCount(&num);`)

		out := providerError(err)
		assert.Contains(t, out, "CUDA failure 100: no CUDA-capable device is detected")
		assert.NotContains(t, out, "onnxruntime_src")
		assert.NotContains(t, out, "ONNXRuntimeError")
		assert.NotContains(t, out, "hostname")
		assert.LessOrEqual(t, len([]rune(out)), providerErrorLen)
	})
	t.Run("KeepsThePrefixWhenItCarriesTheReason", func(t *testing.T) {
		// Verbatim from the runtime when cuDNN is absent. Here the useful text is in front of
		// the source location, so the removal must not swallow it.
		err := errors.New(`Error running network: Non-zero status code returned while running Conv node. ` +
			`Name:'conv_1_conv2d' Status Message: /onnxruntime_src/onnxruntime/core/providers/cuda/cuda_kernel.h:272 ` +
			`static cudnnContext* onnxruntime::cuda::CudaKernel::RequireCudnnHandle(cudnnHandle_t) ` +
			`[ONNXRuntimeError] : 9 : NOT_IMPLEMENTED : cuDNN is unavailable or disabled for CUDA Execution Provider`)

		out := providerError(err)
		assert.Contains(t, out, "running Conv node")
		assert.Contains(t, out, "cuDNN is unavailable or disabled")
		assert.NotContains(t, out, "onnxruntime_src")
		assert.NotContains(t, out, "cudnnContext")
	})
	t.Run("Bounded", func(t *testing.T) {
		out := providerError(errors.New(strings.Repeat("x", 4000)))
		assert.LessOrEqual(t, len([]rune(out)), providerErrorLen)
	})
	t.Run("CollapsesWhitespace", func(t *testing.T) {
		assert.Equal(t, "a b c", providerError(errors.New("a\n  b\tc")))
	})
}

func TestDestroySessionOptions(t *testing.T) {
	t.Run("Nil", func(t *testing.T) {
		assert.NotPanics(t, func() { DestroySessionOptions(nil) })
	})
	t.Run("Success", func(t *testing.T) {
		requireSessionRuntime(t)
		hook := captureProviderLog(t)

		opts, err := newBaseSessionOptions(SessionSettings{IntraOpThreads: 1})
		require.NoError(t, err)

		DestroySessionOptions(opts)
		assert.Empty(t, hook.AllEntries())
	})
}

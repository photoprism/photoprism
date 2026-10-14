package onnx

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	onnxruntime "github.com/yalue/onnxruntime_go"
)

var (
	runtimeOnce    sync.Once
	runtimeInitErr error
	runtimePath    string
	executableVar  = os.Executable

	setLibraryPathVar     = onnxruntime.SetSharedLibraryPath
	initEnvironmentVar    = onnxruntime.InitializeEnvironment
	disableTelemetryVar   = onnxruntime.DisableTelemetry
	destroyEnvironmentVar = onnxruntime.DestroyEnvironment
)

// EnsureRuntime loads the ONNX Runtime shared library and initializes the global environment with
// telemetry disabled, and must succeed before any model is inspected or loaded.
//
// The binding requests the exact C API version of its vendored headers, so it fails against an
// older library; the candidate errors are kept in the message to tell missing from mismatched.
func EnsureRuntime(libraryPath string) error {
	runtimeOnce.Do(func() {
		runtimePath, runtimeInitErr = loadRuntime(SharedLibraryCandidates(libraryPath))
	})

	return runtimeInitErr
}

// loadRuntime initializes the environment from the first loadable candidate with telemetry disabled,
// and fails closed if it cannot be: the variable must be set before loading, as it is read only once.
func loadRuntime(candidates []string) (string, error) {
	if err := disableTelemetryEnv(); err != nil {
		return "", fmt.Errorf("failed to disable ONNX runtime telemetry: %w", err)
	}

	var errs []string

	for _, candidate := range candidates {
		setLibraryPathVar(candidate)

		if err := initEnvironmentVar(); err != nil {
			// Collect errors so we can surface meaningful diagnostics when all options fail.
			errs = append(errs, fmt.Sprintf("%s (%v)", candidate, err))
			continue
		}

		if err := disableTelemetryVar(); err != nil {
			_ = destroyEnvironmentVar()
			return "", fmt.Errorf("failed to disable ONNX runtime telemetry: %w", err)
		}

		return candidate, nil
	}

	if len(errs) == 0 {
		return "", errors.New("no ONNX runtime library candidates")
	}

	return "", fmt.Errorf("failed to load ONNX runtime: %s", strings.Join(errs, "; "))
}

// RuntimeLibraryPath returns the shared-library candidate used to initialize ONNX Runtime.
func RuntimeLibraryPath() string {
	return runtimePath
}

// SharedLibraryCandidates lists the library paths to try when loading the ONNX Runtime,
// starting with an explicitly configured one.
func SharedLibraryCandidates(explicit string) []string {
	appendUnique := func(list []string, seen map[string]struct{}, values ...string) []string {
		for _, value := range values {
			if value == "" {
				continue
			}
			if _, ok := seen[value]; ok {
				continue
			}
			list = append(list, value)
			seen[value] = struct{}{}
		}
		return list
	}

	seen := make(map[string]struct{})
	candidates := make([]string, 0, 8)
	candidates = appendUnique(candidates, seen, explicit)
	candidates = appendUnique(candidates, seen,
		"libonnxruntime.so",
		"libonnxruntime.so.1",
		"onnxruntime.so",
	)

	if exePath, err := executableVar(); err == nil {
		exeDir := filepath.Dir(exePath)
		rootDir := filepath.Dir(exeDir)

		candidates = appendUnique(candidates, seen,
			filepath.Join(exeDir, "libonnxruntime.so"),
			filepath.Join(exeDir, "lib", "libonnxruntime.so"),
		)

		if rootDir != "" && rootDir != "." && rootDir != exeDir {
			candidates = appendUnique(candidates, seen, filepath.Join(rootDir, "lib", "libonnxruntime.so"))
		}
	}

	return candidates
}

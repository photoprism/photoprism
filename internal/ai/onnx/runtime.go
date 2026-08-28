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
	executableVar  = os.Executable
)

// EnsureRuntime loads the ONNX Runtime shared library and initializes the global environment,
// and must succeed before any model is inspected or loaded.
//
// The binding requests the exact C API version of its vendored headers, so it fails against an
// older library; the candidate errors are kept in the message to tell missing from mismatched.
func EnsureRuntime(libraryPath string) error {
	runtimeOnce.Do(func() {
		candidates := SharedLibraryCandidates(libraryPath)
		var errs []string

		for _, candidate := range candidates {
			onnxruntime.SetSharedLibraryPath(candidate)

			if err := onnxruntime.InitializeEnvironment(); err != nil {
				// Collect errors so we can surface meaningful diagnostics when all options fail.
				errs = append(errs, fmt.Sprintf("%s (%v)", candidate, err))
				continue
			}

			// Successfully initialized; stop retrying.
			runtimeInitErr = nil
			return
		}

		if len(errs) == 0 {
			runtimeInitErr = errors.New("no ONNX runtime library candidates")
			return
		}

		runtimeInitErr = fmt.Errorf("failed to load ONNX runtime: %s", strings.Join(errs, "; "))
	})

	return runtimeInitErr
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

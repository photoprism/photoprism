package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/photoprism/photoprism/pkg/clean"
)

// videoRemuxPath identifies a directory entry through its resolved parent, including missing parents.
// The final component is not resolved because publication replaces that entry, not its target.
func videoRemuxPath(name string) (string, error) {
	if name == "" {
		return "", fmt.Errorf("remux: file name is empty")
	}
	absolute, err := filepath.Abs(name)
	if err != nil {
		return "", err
	}
	parent, tail := filepath.Dir(absolute), filepath.Base(absolute)
	for {
		resolved, resolveErr := filepath.EvalSymlinks(parent)
		if resolveErr == nil {
			return filepath.Join(resolved, tail), nil
		}
		if !os.IsNotExist(resolveErr) {
			return "", resolveErr
		}
		if _, statErr := os.Lstat(parent); !os.IsNotExist(statErr) {
			return "", resolveErr
		}
		next := filepath.Dir(parent)
		if next == parent {
			return "", resolveErr
		}
		tail = filepath.Join(filepath.Base(parent), tail)
		parent = next
	}
}

// videoValidateRemuxPlans validates the entire selection before a batch can publish output.
func videoValidateRemuxPlans(plans []videoRemuxPlan, inputs []string) error {
	sources := make(map[string]map[string]string, len(inputs))
	for _, input := range inputs {
		entry, err := videoRemuxPath(input)
		if err != nil {
			return fmt.Errorf("remux: resolve input: %w", err)
		}
		target, err := filepath.EvalSymlinks(entry)
		if err != nil {
			return fmt.Errorf("remux: resolve input: %w", err)
		}
		for _, key := range []string{entry, target} {
			if sources[key] == nil {
				sources[key] = make(map[string]string)
			}
			sources[key][entry] = input
		}
	}
	destinations := make(map[string]string, len(plans))
	for _, plan := range plans {
		source, err := videoRemuxPath(plan.SrcPath)
		if err != nil {
			return fmt.Errorf("remux: resolve source: %w", err)
		}
		destination, err := videoRemuxPath(plan.DestPath)
		if err != nil {
			return fmt.Errorf("remux: resolve output: %w", err)
		}
		if previous, exists := destinations[destination]; exists {
			return fmt.Errorf("remux: output %s is selected more than once by %s and %s", clean.Log(plan.DestPath), clean.Log(previous), clean.Log(plan.SrcPath))
		}
		destinations[destination] = plan.SrcPath
		for input, name := range sources[destination] {
			if input != source {
				return fmt.Errorf("remux: output %s from %s is another selected input %s", clean.Log(plan.DestPath), clean.Log(plan.SrcPath), clean.Log(name))
			}
		}
	}
	return nil
}

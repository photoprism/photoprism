package onnx

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
)

// VerifyChecksum reports whether modelPath holds the artifact this description was written for.
//
// Names collide across publishers - two unrelated models have shipped as glintr100.onnx - so
// confirming by name would apply one model's preprocessing to another's weights, which fails
// quietly. A description with no recorded checksum is accepted, so custom models keep working.
func (m *ModelInfo) VerifyChecksum(modelPath string) error {
	if m == nil || m.SHA256 == "" {
		return nil
	}

	actual := fs.Sha256(modelPath)

	if actual == "" {
		return fmt.Errorf("failed to read %s", clean.Log(filepath.Base(modelPath)))
	}

	if !strings.EqualFold(actual, m.SHA256) {
		return fmt.Errorf("%s has checksum %s, expected %s", clean.Log(filepath.Base(modelPath)), clean.Log(actual), clean.Log(m.SHA256))
	}

	return nil
}

// VerifyGraph reports whether an inspected graph contradicts this description. A recorded
// value disagreeing with the graph is a configuration error to surface rather than a
// difference to reconcile, because one of the two is describing a different model.
// Dimensions the graph leaves dynamic are not compared.
func (m *ModelInfo) VerifyGraph(graph *ModelInfo) error {
	if m == nil || graph == nil {
		return nil
	}

	if m.Input != nil && graph.Input != nil {
		if err := verifyAxis("input width", m.Input.Width, graph.Input.Width); err != nil {
			return err
		}

		if err := verifyAxis("input height", m.Input.Height, graph.Input.Height); err != nil {
			return err
		}

		if m.Input.Layout != LayoutUndefined && graph.Input.Layout != LayoutUndefined && m.Input.Layout != graph.Input.Layout {
			return fmt.Errorf("model has %s input layout, expected %s", graph.Input.Layout, m.Input.Layout)
		}
	}

	if m.Output != nil && graph.Output != nil {
		if err := verifyAxis("output width", m.Output.Width, graph.Output.Width); err != nil {
			return err
		}
	}

	return nil
}

// verifyAxis compares a recorded dimension with the one read from the graph, ignoring
// dimensions that either side leaves unspecified.
func verifyAxis(name string, expected, actual int) error {
	if expected <= 0 || actual <= 0 || expected == actual {
		return nil
	}

	return fmt.Errorf("model has %s %d, expected %d", name, actual, expected)
}

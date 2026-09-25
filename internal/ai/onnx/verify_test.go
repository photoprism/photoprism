package onnx

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/pkg/fs"
)

func TestModelInfo_VerifyChecksum(t *testing.T) {
	fileName := filepath.Join(t.TempDir(), "model.onnx")
	require.NoError(t, os.WriteFile(fileName, []byte("photoprism"), fs.ModeFile))
	sum := fs.Sha256(fileName)
	require.NotEmpty(t, sum)

	t.Run("Success", func(t *testing.T) {
		require.NoError(t, (&ModelInfo{SHA256: sum}).VerifyChecksum(fileName))
	})
	t.Run("CaseInsensitive", func(t *testing.T) {
		require.NoError(t, (&ModelInfo{SHA256: strings.ToUpper(sum)}).VerifyChecksum(fileName))
	})
	t.Run("Mismatch", func(t *testing.T) {
		err := (&ModelInfo{SHA256: "0000000000000000000000000000000000000000000000000000000000000000"}).VerifyChecksum(fileName)
		require.Error(t, err)
		assert.Contains(t, err.Error(), sum)
	})
	t.Run("NoChecksumRecorded", func(t *testing.T) {
		// Custom models supplied through PHOTOPRISM_MODELS_PATH have no registry entry, so
		// an empty checksum must stay acceptable.
		require.NoError(t, (&ModelInfo{}).VerifyChecksum(fileName))
	})
	t.Run("MissingFile", func(t *testing.T) {
		require.Error(t, (&ModelInfo{SHA256: sum}).VerifyChecksum(filepath.Join(t.TempDir(), "missing.onnx")))
	})
	t.Run("Nil", func(t *testing.T) {
		var m *ModelInfo
		require.NoError(t, m.VerifyChecksum(fileName))
	})
}

func TestModelInfo_VerifyGraph(t *testing.T) {
	registered := &ModelInfo{
		Input:  &Input{Width: 112, Height: 112, Layout: LayoutNCHW},
		Output: &Output{Width: 128, Count: 1},
	}

	t.Run("Success", func(t *testing.T) {
		graph := &ModelInfo{
			Input:  &Input{Width: 112, Height: 112, Layout: LayoutNCHW},
			Output: &Output{Width: 128},
		}
		require.NoError(t, registered.VerifyGraph(graph))
	})
	t.Run("DynamicAxisIgnored", func(t *testing.T) {
		graph := &ModelInfo{Input: &Input{}, Output: &Output{}}
		require.NoError(t, registered.VerifyGraph(graph))
	})
	t.Run("WidthMismatch", func(t *testing.T) {
		graph := &ModelInfo{Input: &Input{Width: 384, Height: 112}}
		err := registered.VerifyGraph(graph)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "input width 384")
	})
	t.Run("HeightMismatch", func(t *testing.T) {
		graph := &ModelInfo{Input: &Input{Width: 112, Height: 384}}
		err := registered.VerifyGraph(graph)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "input height 384")
	})
	t.Run("LayoutMismatch", func(t *testing.T) {
		graph := &ModelInfo{Input: &Input{Width: 112, Height: 112, Layout: LayoutNHWC}}
		err := registered.VerifyGraph(graph)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "input layout")
	})
	t.Run("OutputWidthMismatch", func(t *testing.T) {
		graph := &ModelInfo{Output: &Output{Width: 512}}
		err := registered.VerifyGraph(graph)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "output width 512")
	})
	t.Run("OutputCountMismatch", func(t *testing.T) {
		graph := &ModelInfo{Output: &Output{Width: 128, Count: 2}}
		err := registered.VerifyGraph(graph)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "output count 2")
	})
	t.Run("NilGraph", func(t *testing.T) {
		require.NoError(t, registered.VerifyGraph(nil))
	})
	t.Run("NilReceiver", func(t *testing.T) {
		var m *ModelInfo
		require.NoError(t, m.VerifyGraph(registered))
	})
}

func TestVerifyAxis(t *testing.T) {
	t.Run("Equal", func(t *testing.T) {
		require.NoError(t, verifyAxis("input width", 112, 112))
	})
	t.Run("ExpectedUnset", func(t *testing.T) {
		require.NoError(t, verifyAxis("input width", 0, 112))
	})
	t.Run("ActualDynamic", func(t *testing.T) {
		require.NoError(t, verifyAxis("input width", 112, 0))
	})
	t.Run("Mismatch", func(t *testing.T) {
		require.Error(t, verifyAxis("input width", 112, 384))
	})
}

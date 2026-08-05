package face

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/pkg/fs/fastwalk"
)

var modelPath, _ = filepath.Abs("../../../assets/models/facenet")
var detectorModelPath, _ = filepath.Abs("../../../assets/models/scrfd/" + DefaultONNXModelFilename)

func TestNet(t *testing.T) {
	prev := UseEngine(nil)
	t.Cleanup(func() {
		current := UseEngine(prev)
		if current != nil {
			_ = current.Close()
		}
	})

	err := ConfigureEngine(EngineSettings{
		Name: EngineONNX,
		ONNX: ONNXOptions{
			ModelPath: detectorModelPath,
			Threads:   1,
		},
	})
	if err != nil {
		// Only a missing model may skip this test, as a detector that fails to
		// initialize despite being present indicates a broken ONNX Runtime or a
		// binding that does not match the installed library version.
		if _, statErr := os.Stat(detectorModelPath); statErr != nil {
			t.Skipf("faces: skipping detector-dependent test, %s is not available", DefaultONNXModelFilename)
		}

		t.Fatalf("faces: failed to initialize detector: %s", err)
	}
	require.Equal(t, EngineONNX, ActiveEngineName())

	faceNet := NewModel(modelPath, "testdata/cache", 160, nil, false)
	detectedFiles := 0
	embeddedFaces := 0

	if err := fastwalk.Walk("testdata", func(fileName string, info os.FileMode) error {
		if info.IsDir() || filepath.Base(filepath.Dir(fileName)) != "testdata" {
			return nil
		}

		t.Run(fileName, func(t *testing.T) {
			baseName := filepath.Base(fileName)

			faces, err := faceNet.Detect(fileName, 20, false, -1)

			if err != nil {
				t.Fatal(err)
			}

			if len(faces) > 0 {
				detectedFiles++
			}

			for i, f := range faces {
				if len(f.Embeddings) == 0 {
					continue
				}

				embeddedFaces++
				magnitude := f.Embeddings[0].Magnitude()
				assert.InDeltaf(t, 1.0, magnitude, 0.02, "embedding %d in %s should stay normalized", i, baseName)
			}
		})

		return nil
	}); err != nil {
		t.Fatal(err)
	}

	assert.Greater(t, detectedFiles, 0)
	assert.Greater(t, embeddedFaces, 0)
}

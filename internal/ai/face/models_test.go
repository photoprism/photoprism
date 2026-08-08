package face

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNormalizeModelName(t *testing.T) {
	t.Run("Lowercase", func(t *testing.T) {
		assert.Equal(t, "sface", NormalizeModelName("SFace"))
	})
	t.Run("Hyphens", func(t *testing.T) {
		assert.Equal(t, "arcface_r50", NormalizeModelName("ArcFace-R50"))
	})
	t.Run("Whitespace", func(t *testing.T) {
		assert.Equal(t, "facenet", NormalizeModelName("  facenet\n"))
	})
	t.Run("Empty", func(t *testing.T) {
		assert.Equal(t, "", NormalizeModelName(""))
	})
}

func TestModelsComparable(t *testing.T) {
	t.Run("SameModel", func(t *testing.T) {
		assert.True(t, ModelsComparable(ModelSFace, ModelSFace))
	})
	t.Run("LegacyFaceNet", func(t *testing.T) {
		assert.True(t, ModelsComparable("", ModelFaceNet))
	})
	t.Run("LegacyArcFace", func(t *testing.T) {
		assert.False(t, ModelsComparable("", ModelArcFaceR50))
	})
	t.Run("DifferentModel", func(t *testing.T) {
		assert.False(t, ModelsComparable(ModelFaceNet, ModelArcFaceR50))
	})
	t.Run("MissingCurrent", func(t *testing.T) {
		assert.False(t, ModelsComparable(ModelFaceNet, ""))
	})
	t.Run("LegacyWithoutCurrent", func(t *testing.T) {
		assert.True(t, ModelsComparable("", ""))
	})
}

func TestParseModelName(t *testing.T) {
	t.Run("Auto", func(t *testing.T) {
		assert.Equal(t, ModelAuto, ParseModelName("Auto"))
	})
	t.Run("None", func(t *testing.T) {
		assert.Equal(t, ModelNone, ParseModelName("none"))
	})
	t.Run("Registered", func(t *testing.T) {
		assert.Equal(t, ModelSFace, ParseModelName("SFace"))
	})
	t.Run("HyphenatedAlias", func(t *testing.T) {
		assert.Equal(t, ModelArcFaceR50, ParseModelName("arcface-r50"))
	})
	t.Run("Unknown", func(t *testing.T) {
		assert.Equal(t, ModelAuto, ParseModelName("dlib"))
	})
	t.Run("Empty", func(t *testing.T) {
		assert.Equal(t, ModelAuto, ParseModelName(""))
	})
}

func TestKnownModelName(t *testing.T) {
	t.Run("Auto", func(t *testing.T) {
		assert.True(t, KnownModelName("auto"))
	})
	t.Run("None", func(t *testing.T) {
		assert.True(t, KnownModelName("NONE"))
	})
	t.Run("Registered", func(t *testing.T) {
		assert.True(t, KnownModelName("arcface_mbf"))
	})
	t.Run("Unknown", func(t *testing.T) {
		assert.False(t, KnownModelName("dlib"))
	})
	t.Run("Empty", func(t *testing.T) {
		assert.False(t, KnownModelName(""))
	})
}

func TestFindEmbeddingModel(t *testing.T) {
	t.Run("FaceNet", func(t *testing.T) {
		m := FindEmbeddingModel(ModelFaceNet)
		require.NotNil(t, m)
		assert.Equal(t, RuntimeTensorFlow, m.Runtime)
		assert.Equal(t, 512, m.Dims)
		assert.Equal(t, AlignBox, m.Alignment)
	})
	t.Run("SFace", func(t *testing.T) {
		m := FindEmbeddingModel("SFace")
		require.NotNil(t, m)
		assert.Equal(t, RuntimeONNX, m.Runtime)
		assert.Equal(t, 128, m.Dims)
		assert.Equal(t, 112, m.Width)
		assert.Equal(t, AlignArcFace5, m.Alignment)
	})
	t.Run("AuraFace", func(t *testing.T) {
		m := FindEmbeddingModel("AuraFace")
		require.NotNil(t, m)
		assert.Equal(t, RuntimeONNX, m.Runtime)
		assert.Equal(t, 512, m.Dims)
		assert.Equal(t, 112, m.Width)
		assert.Equal(t, AlignArcFace5, m.Alignment)
		assert.Equal(t, LicenseApache2, m.License)
	})
	t.Run("Unknown", func(t *testing.T) {
		assert.Nil(t, FindEmbeddingModel("dlib"))
	})
	t.Run("Auto", func(t *testing.T) {
		assert.Nil(t, FindEmbeddingModel(ModelAuto))
	})
}

func TestEmbeddingModelNames(t *testing.T) {
	names := EmbeddingModelNames()

	t.Run("Sorted", func(t *testing.T) {
		assert.Equal(t, []ModelName{ModelArcFaceMBF, ModelArcFaceR50, ModelAuraFace, ModelFaceNet, ModelSFace}, names)
	})
	t.Run("ExcludesAliases", func(t *testing.T) {
		assert.NotContains(t, names, ModelAuto)
		assert.NotContains(t, names, ModelNone)
	})
}

func TestModelUsageString(t *testing.T) {
	t.Run("Aliases", func(t *testing.T) {
		assert.True(t, strings.HasPrefix(ModelUsageString(), "auto, none, "))
	})
	t.Run("AllModels", func(t *testing.T) {
		for _, name := range EmbeddingModelNames() {
			assert.Contains(t, ModelUsageString(), name)
		}
	})
}

func TestEmbeddingModels(t *testing.T) {
	for name, m := range EmbeddingModels {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, name, m.Name)
			assert.NotEmpty(t, m.Dir)
			assert.Positive(t, m.Dims)
			assert.Positive(t, m.Width)
			assert.Positive(t, m.Height)
			assert.Positive(t, m.Scale)
			assert.NotEmpty(t, m.License)
			assert.Contains(t, []EmbeddingRuntime{RuntimeTensorFlow, RuntimeONNX}, m.Runtime)
			assert.Contains(t, []CropAlignment{AlignBox, AlignArcFace5}, m.Alignment)

			// A model without calibrated thresholds would silently fall back to the
			// FaceNet-tuned values and discard most of its matches.
			assert.Positive(t, m.ClusterDist)
			assert.Positive(t, m.ClusterRadius)
			assert.Positive(t, m.MatchDist)
			assert.Less(t, m.ClusterRadius, m.ClusterDist)

			// ONNX models are single files, TensorFlow models are SavedModel directories.
			if m.Runtime == RuntimeONNX {
				assert.NotEmpty(t, m.FileName)
			} else {
				assert.Empty(t, m.FileName)
			}
		})
	}
}

func TestEmbeddingModelThresholds(t *testing.T) {
	t.Run("FaceNetKeepsShippedValues", func(t *testing.T) {
		// Changing these would alter matching for every library that upgrades.
		m := FindEmbeddingModel(ModelFaceNet)
		require.NotNil(t, m)
		assert.Equal(t, ClusterDistDefault, m.ClusterDist)
		assert.Equal(t, ClusterRadiusDefault, m.ClusterRadius)
		assert.Equal(t, MatchDistDefault, m.MatchDist)
	})
	t.Run("CalibratedModelsDifferFromFaceNet", func(t *testing.T) {
		for _, name := range []ModelName{ModelSFace, ModelAuraFace, ModelArcFaceR50, ModelArcFaceMBF} {
			m := FindEmbeddingModel(name)
			require.NotNil(t, m, name)
			assert.NotEqual(t, ClusterDistDefault, m.ClusterDist, name)
			assert.NotEqual(t, MatchDistDefault, m.MatchDist, name)
		}
	})
}

func TestAutoModelPreference(t *testing.T) {
	t.Run("PrefersFaceNet", func(t *testing.T) {
		require.NotEmpty(t, AutoModelPreference)
		assert.Equal(t, ModelFaceNet, AutoModelPreference[0])
	})
	t.Run("Registered", func(t *testing.T) {
		for _, name := range AutoModelPreference {
			assert.NotNil(t, FindEmbeddingModel(name), name)
		}
	})
	t.Run("Complete", func(t *testing.T) {
		assert.Len(t, AutoModelPreference, len(EmbeddingModels))
	})
}

func TestEmbeddingModel_FilePath(t *testing.T) {
	t.Run("ONNXFile", func(t *testing.T) {
		m := FindEmbeddingModel(ModelSFace)
		require.NotNil(t, m)
		assert.Equal(t, filepath.Join("/models", "sface", "face_recognition_sface_2021dec.onnx"), m.FilePath("/models"))
	})
	t.Run("SavedModelDir", func(t *testing.T) {
		m := FindEmbeddingModel(ModelFaceNet)
		require.NotNil(t, m)
		assert.Equal(t, filepath.Join("/models", "facenet"), m.FilePath("/models"))
	})
	t.Run("EmptyModelsPath", func(t *testing.T) {
		assert.Equal(t, "", FindEmbeddingModel(ModelSFace).FilePath(""))
	})
	t.Run("NilModel", func(t *testing.T) {
		var m *EmbeddingModel
		assert.Equal(t, "", m.FilePath("/models"))
	})
}

func TestEmbeddingModel_Installed(t *testing.T) {
	modelsPath := t.TempDir()

	t.Run("MissingFile", func(t *testing.T) {
		assert.False(t, FindEmbeddingModel(ModelSFace).Installed(modelsPath))
	})
	t.Run("MissingDir", func(t *testing.T) {
		assert.False(t, FindEmbeddingModel(ModelFaceNet).Installed(modelsPath))
	})
	t.Run("ExistingFile", func(t *testing.T) {
		m := FindEmbeddingModel(ModelSFace)
		require.NoError(t, os.MkdirAll(filepath.Join(modelsPath, m.Dir), 0o750))
		require.NoError(t, os.WriteFile(m.FilePath(modelsPath), []byte("onnx"), 0o600))
		assert.True(t, m.Installed(modelsPath))
	})
	t.Run("ExistingDir", func(t *testing.T) {
		m := FindEmbeddingModel(ModelFaceNet)
		require.NoError(t, os.MkdirAll(m.FilePath(modelsPath), 0o750))
		assert.True(t, m.Installed(modelsPath))
	})
	t.Run("NilModel", func(t *testing.T) {
		var m *EmbeddingModel
		assert.False(t, m.Installed(modelsPath))
	})
}

func TestEmbeddingModel_Aligned(t *testing.T) {
	t.Run("ArcFace5", func(t *testing.T) {
		assert.True(t, FindEmbeddingModel(ModelSFace).Aligned())
	})
	t.Run("Box", func(t *testing.T) {
		assert.False(t, FindEmbeddingModel(ModelFaceNet).Aligned())
	})
	t.Run("NilModel", func(t *testing.T) {
		var m *EmbeddingModel
		assert.False(t, m.Aligned())
	})
}

func TestEmbeddingModel_License(t *testing.T) {
	t.Run("Apache2", func(t *testing.T) {
		assert.Equal(t, LicenseApache2, FindEmbeddingModel(ModelSFace).License)
	})
	t.Run("ResearchOnly", func(t *testing.T) {
		assert.Equal(t, LicenseResearchOnly, FindEmbeddingModel(ModelArcFaceR50).License)
	})
	t.Run("Unknown", func(t *testing.T) {
		assert.Equal(t, LicenseUnknown, FindEmbeddingModel(ModelFaceNet).License)
	})
}

func TestEmbeddingModel_String(t *testing.T) {
	t.Run("Name", func(t *testing.T) {
		assert.Equal(t, ModelArcFaceR50, FindEmbeddingModel(ModelArcFaceR50).String())
	})
	t.Run("NilModel", func(t *testing.T) {
		var m *EmbeddingModel
		assert.Equal(t, ModelNone, m.String())
	})
}

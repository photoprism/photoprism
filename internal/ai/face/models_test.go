package face

import (
	"math"
	"os"
	"path/filepath"
	"regexp"
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

func TestSameEmbeddingSpace(t *testing.T) {
	t.Run("BothBlank", func(t *testing.T) {
		assert.True(t, SameEmbeddingSpace("", ""))
	})
	t.Run("SameModel", func(t *testing.T) {
		assert.True(t, SameEmbeddingSpace(ModelSFace, ModelSFace))
	})
	// A library mid-upgrade holds both forms of FaceNet, so neither order may be rejected.
	t.Run("BlankAndFaceNet", func(t *testing.T) {
		assert.True(t, SameEmbeddingSpace("", ModelFaceNet))
	})
	t.Run("FaceNetAndBlank", func(t *testing.T) {
		assert.True(t, SameEmbeddingSpace(ModelFaceNet, ""))
	})
	t.Run("BlankAndSFace", func(t *testing.T) {
		assert.False(t, SameEmbeddingSpace("", ModelSFace))
	})
	t.Run("SFaceAndBlank", func(t *testing.T) {
		assert.False(t, SameEmbeddingSpace(ModelSFace, ""))
	})
	t.Run("DifferentModels", func(t *testing.T) {
		assert.False(t, SameEmbeddingSpace(ModelFaceNet, ModelArcFaceR50))
	})
}

func TestParseModelName(t *testing.T) {
	t.Run("Detect", func(t *testing.T) {
		assert.Equal(t, ModelDetect, ParseModelName("Detect"))
	})
	t.Run("AutoIsDetect", func(t *testing.T) {
		assert.Equal(t, ModelDetect, ParseModelName("Auto"))
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
		assert.Equal(t, ModelDetect, ParseModelName("dlib"))
	})
	t.Run("Empty", func(t *testing.T) {
		assert.Equal(t, ModelDetect, ParseModelName(""))
	})
}

func TestKnownModelName(t *testing.T) {
	t.Run("Detect", func(t *testing.T) {
		assert.True(t, KnownModelName("detect"))
	})
	t.Run("Auto", func(t *testing.T) {
		assert.True(t, KnownModelName("auto"))
	})
	t.Run("Empty", func(t *testing.T) {
		assert.True(t, KnownModelName(""))
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
		width, height := m.InputSize()
		assert.Equal(t, ArcFaceTemplateSize, width)
		assert.Equal(t, ArcFaceTemplateSize, height)
		assert.Equal(t, AlignArcFace5, m.Alignment)
	})
	t.Run("AuraFace", func(t *testing.T) {
		m := FindEmbeddingModel("AuraFace")
		require.NotNil(t, m)
		assert.Equal(t, RuntimeONNX, m.Runtime)
		assert.Equal(t, 512, m.Dims)
		width, _ := m.InputSize()
		assert.Equal(t, ArcFaceTemplateSize, width)
		assert.Equal(t, AlignArcFace5, m.Alignment)
		assert.Equal(t, LicenseApache2, m.WeightLicense())
	})
	t.Run("Unknown", func(t *testing.T) {
		assert.Nil(t, FindEmbeddingModel("dlib"))
	})
	t.Run("Detect", func(t *testing.T) {
		assert.Nil(t, FindEmbeddingModel(ModelDetect))
	})
}

func TestEmbeddingModelNames(t *testing.T) {
	names := EmbeddingModelNames()

	t.Run("Sorted", func(t *testing.T) {
		assert.Equal(t, []ModelName{ModelArcFaceMBF, ModelArcFaceR50, ModelAuraFace, ModelFaceNet, ModelSFace}, names)
	})
	t.Run("ExcludesAliases", func(t *testing.T) {
		assert.NotContains(t, names, ModelDetect)
		assert.NotContains(t, names, ModelAuto)
		assert.NotContains(t, names, ModelNone)
	})
}

func TestModelUsageString(t *testing.T) {
	t.Run("Aliases", func(t *testing.T) {
		assert.True(t, strings.HasPrefix(ModelUsageString(), "detect, none, "))
	})
	t.Run("PermissiveModels", func(t *testing.T) {
		for _, name := range EmbeddingModelNames() {
			if FindEmbeddingModel(name).LicenseGated() {
				continue
			}

			assert.Contains(t, ModelUsageString(), name)
		}
	})
	t.Run("OmitsGatedModels", func(t *testing.T) {
		// Help text reads as an offer, and these weights may not be used until their terms
		// have been accepted, so they are named where that acceptance is asked for instead.
		assert.NotContains(t, ModelUsageString(), ModelArcFaceR50)
		assert.NotContains(t, ModelUsageString(), ModelArcFaceMBF)
	})
}

func TestEmbeddingModels(t *testing.T) {
	for name, m := range EmbeddingModels {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, name, m.Name)
			assert.NotEmpty(t, m.Dir)
			assert.Positive(t, m.Dims)
			assert.NotEmpty(t, m.WeightLicense())
			assert.Contains(t, []EmbeddingRuntime{RuntimeTensorFlow, RuntimeONNX}, m.Runtime)
			assert.Contains(t, []CropAlignment{AlignBox, AlignArcFace5}, m.Alignment)

			// A model without calibrated thresholds would silently fall back to the
			// FaceNet-tuned values and discard most of its matches.
			assert.Positive(t, m.ClusterDist)
			assert.Positive(t, m.ClusterRadius)
			assert.Positive(t, m.MatchDist)
			assert.Less(t, m.ClusterRadius, m.ClusterDist)

			// The collision floor and its slack follow the width of the model's distance
			// scale, so leaving them at the FaceNet values would be the same trap the
			// per-model thresholds exist to close.
			assert.Positive(t, m.CollisionDist)
			assert.Positive(t, m.Epsilon)
			assert.Less(t, m.Epsilon, m.CollisionDist)
			assert.Less(t, m.CollisionDist, m.MatchDist)
			scale := m.ClusterDist / ClusterDistDefault
			assert.InDelta(t, roundTo3(scale*CollisionDistDefault), m.CollisionDist, 1e-9)
			assert.InDelta(t, roundTo3(scale*EpsilonDefault), m.Epsilon, 1e-9)

			// ONNX models are single files described by the shared model info, while
			// TensorFlow models are SavedModel directories that have none.
			if m.Runtime == RuntimeONNX {
				require.NotNil(t, m.ONNX)
				assert.NotEmpty(t, m.ONNX.File)
				assert.NotEmpty(t, m.ONNX.SHA256)
				require.NotNil(t, m.ONNX.Input)
				assert.Positive(t, m.ONNX.Input.Width)
				assert.Positive(t, m.ONNX.Input.Height)
				assert.True(t, m.ONNX.Input.ColorOrder.Valid())
				assert.False(t, m.ONNX.Input.Normalization.IsZero())
				assert.Equal(t, m.Dims, m.ONNX.OutputWidth())
			} else {
				assert.Nil(t, m.ONNX)
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
	t.Run("AcceptDistStaysBelowCeiling", func(t *testing.T) {
		// A model calibrated above the ceiling would be silently capped at runtime, so
		// adding one is a decision to make here rather than discover in production.
		for name, m := range EmbeddingModels {
			assert.Less(t, m.ClusterRadius+m.MatchDist, float64(AcceptDistMax), name)
		}
	})
	t.Run("CalibratedValuesStayConfigurable", func(t *testing.T) {
		// ConfigDistMax bounds what an operator may set, so a model calibrated above it
		// would carry values that are refused as out of range when they are set by hand.
		for name, m := range EmbeddingModels {
			assert.LessOrEqual(t, m.ClusterRadius+m.MatchDist, float64(ConfigDistMax), name)
			assert.LessOrEqual(t, m.ClusterDist, float64(ConfigDistMax), name)
		}
	})
	t.Run("ClusterDistStaysWithinAcceptance", func(t *testing.T) {
		// A face close enough to seed a cluster must be close enough for that cluster to
		// accept it again, or the migration relinks a marker the matcher then refuses and
		// cannot renew. Holding this per model is what lets the outlier rule stay a plain
		// ClusterDist comparison instead of carrying the acceptance bound as well.
		for name, m := range EmbeddingModels {
			assert.LessOrEqual(t, m.ClusterDist, m.ClusterRadius+m.MatchDist, name)
		}
	})
}

func TestAutoModelPreference(t *testing.T) {
	t.Run("PrefersSFace", func(t *testing.T) {
		// Only libraries without face vectors follow this order, so the first entry is
		// what a fresh install starts with rather than what an upgrade switches to.
		require.NotEmpty(t, AutoModelPreference)
		assert.Equal(t, ModelSFace, AutoModelPreference[0])
	})
	t.Run("BundledModelsFirst", func(t *testing.T) {
		// Models that are not shipped must not precede one that is, or "auto" would
		// resolve to nothing on an installation that never ran their install script.
		assert.Equal(t, []ModelName{ModelSFace, ModelFaceNet}, AutoModelPreference[:2])
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

func TestEmbeddingModel_InputSize(t *testing.T) {
	t.Run("ONNXModel", func(t *testing.T) {
		width, height := FindEmbeddingModel(ModelSFace).InputSize()
		assert.Equal(t, ArcFaceTemplateSize, width)
		assert.Equal(t, ArcFaceTemplateSize, height)
	})
	t.Run("SavedModel", func(t *testing.T) {
		// TensorFlow entries carry no ONNX description, so their crop size comes from
		// CropSize rather than from the registry.
		width, height := FindEmbeddingModel(ModelFaceNet).InputSize()
		assert.Equal(t, 0, width)
		assert.Equal(t, 0, height)
	})
	t.Run("NilModel", func(t *testing.T) {
		var m *EmbeddingModel
		width, height := m.InputSize()
		assert.Equal(t, 0, width)
		assert.Equal(t, 0, height)
	})
}

func TestEmbeddingModel_WeightLicense(t *testing.T) {
	t.Run("Apache2", func(t *testing.T) {
		assert.Equal(t, LicenseApache2, FindEmbeddingModel(ModelSFace).WeightLicense())
	})
	t.Run("ResearchOnly", func(t *testing.T) {
		assert.Equal(t, LicenseNonFree, FindEmbeddingModel(ModelArcFaceR50).WeightLicense())
	})
	t.Run("Unknown", func(t *testing.T) {
		assert.Equal(t, LicenseUnknown, FindEmbeddingModel(ModelFaceNet).WeightLicense())
	})
	t.Run("NilModel", func(t *testing.T) {
		var m *EmbeddingModel
		assert.Empty(t, m.WeightLicense())
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

// readModelScript returns the contents of an installation script, or skips the test when
// the repository scripts are not available, as in a production image.
func readModelScript(t *testing.T, fileName string) string {
	t.Helper()

	path := filepath.Join("..", "..", "..", "scripts", fileName)
	data, err := os.ReadFile(path) //nolint:gosec // G304: fixed repository path.

	if err != nil {
		t.Skipf("faces: skipping, %s is not available", fileName)
	}

	return string(data)
}

// TestEmbeddingModelChecksums keeps the registry and the installers pinned to the same
// artifact. Bundled models are installed from the shared registry, while the ArcFace
// weights keep a dedicated script because they are license-gated and never mirrored.
func TestEmbeddingModelChecksums(t *testing.T) {
	t.Run("Registry", func(t *testing.T) {
		// Registry fields: name|url|fallback|sha256|type|dir|file
		data := readModelScript(t, filepath.Join("dist", "download-models.sh"))

		for _, name := range []ModelName{ModelSFace, ModelAuraFace} {
			t.Run(name, func(t *testing.T) {
				m := FindEmbeddingModel(name)
				require.NotNil(t, m)
				require.NotNil(t, m.ONNX)

				entry := regexp.MustCompile(`(?m)^` + regexp.QuoteMeta(name) + `\|.*$`).FindString(data)
				require.NotEmpty(t, entry, "download-models.sh must list %s", name)

				fields := strings.Split(entry, "|")
				require.Len(t, fields, 7, "%s must have all registry fields", name)
				assert.Equal(t, m.ONNX.SHA256, fields[3])
				assert.Equal(t, m.ONNX.File, fields[6])
				assert.Equal(t, m.Dir, fields[5])
			})
		}
	})
	t.Run("ArcFace", func(t *testing.T) {
		data := readModelScript(t, filepath.Join("dist", "download-arcface.sh"))

		for name, variable := range map[ModelName]string{
			ModelArcFaceR50: "R50_SHA256",
			ModelArcFaceMBF: "MBF_SHA256",
		} {
			t.Run(name, func(t *testing.T) {
				m := FindEmbeddingModel(name)
				require.NotNil(t, m)
				require.NotNil(t, m.ONNX)

				matches := regexp.MustCompile(variable + `="([0-9a-f]{64})"`).FindStringSubmatch(data)
				require.Len(t, matches, 2, "download-arcface.sh must declare %s", variable)
				assert.Equal(t, matches[1], m.ONNX.SHA256)
			})
		}
	})
}

// roundTo3 rounds a derived threshold the way the registry records it.
func roundTo3(value float64) float64 {
	return math.Round(value*1000) / 1000
}

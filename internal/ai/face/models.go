package face

import (
	"path/filepath"
	"sort"
	"strings"

	"github.com/photoprism/photoprism/internal/ai/onnx"
	"github.com/photoprism/photoprism/pkg/fs"
)

// EmbeddingRuntime identifies the inference runtime that generates face embeddings.
type EmbeddingRuntime = string

const (
	// RuntimeTensorFlow generates embeddings through the TensorFlow bindings.
	RuntimeTensorFlow EmbeddingRuntime = "tensorflow"
	// RuntimeONNX generates embeddings through ONNX Runtime.
	RuntimeONNX EmbeddingRuntime = "onnx"
)

// CropAlignment identifies how face crops are prepared before embedding inference.
type CropAlignment = string

const (
	// AlignBox scales the detected face bounding box without using landmarks.
	AlignBox CropAlignment = "box"
	// AlignArcFace5 warps the five detected landmarks onto the standard ArcFace template.
	AlignArcFace5 CropAlignment = "arcface5"
)

// Licenses that apply to the pretrained weights of the supported embedding models.
const (
	// LicenseApache2 allows redistribution under the Apache License 2.0.
	LicenseApache2 = "Apache-2.0"
	// LicenseResearchOnly restricts the weights to non-commercial research unless licensed separately.
	LicenseResearchOnly = "research-only"
	// LicenseUnknown marks weights whose provenance has not been verified.
	LicenseUnknown = "unknown"
)

// ModelName identifies a face embedding model.
type ModelName = string

const (
	// ModelAuto resolves to the first model in AutoModelPreference that is installed.
	ModelAuto ModelName = "auto"
	// ModelNone disables embedding generation so only face regions are detected.
	ModelNone ModelName = "none"
	// ModelFaceNet is the TensorFlow FaceNet model PhotoPrism has shipped since 2021.
	ModelFaceNet ModelName = "facenet"
	// ModelSFace is the OpenCV Zoo SFace model.
	ModelSFace ModelName = "sface"
	// ModelAuraFace is the fal AuraFace v1 model, an Apache-2.0 ArcFace ResNet-100.
	ModelAuraFace ModelName = "auraface"
	// ModelArcFaceR50 is the InsightFace ArcFace WebFace600K ResNet-50 model.
	ModelArcFaceR50 ModelName = "arcface_r50"
	// ModelArcFaceMBF is the InsightFace ArcFace WebFace600K MobileFaceNet model.
	ModelArcFaceMBF ModelName = "arcface_mbf"
)

// EmbeddingModel describes what a face embedding model expects from the pipeline.
//
// ONNX carries the artifact and the preprocessing contract in the structure that every
// subsystem running an ONNX model shares; it is nil for TensorFlow models. What stays
// here is what only face embedding needs: the alignment mode, the embedding length, and
// the distance thresholds.
//
// ClusterDist, ClusterRadius, MatchDist, CollisionDist, and Epsilon are all expressed in
// that model's own distance scale. Distances are not comparable across models, so a single
// global set of thresholds silently discards most matches for every model except the one
// it was tuned for.
type EmbeddingModel struct {
	Name          ModelName
	Runtime       EmbeddingRuntime
	Dir           string
	Dims          int
	Alignment     CropAlignment
	ONNX          *onnx.ModelInfo
	ClusterDist   float64
	ClusterRadius float64
	MatchDist     float64
	CollisionDist float64
	Epsilon       float64
}

// EmbeddingModels lists the supported face embedding models by name.
//
// The ArcFace entries are recognized so operators can benchmark them, but their
// weights are not installed by "make dep" because InsightFace publishes them for
// non-commercial research only. AuraFace is Apache-2.0 and may be redistributed,
// but its graph is 261 MB, so it is an opt-in download rather than a bundled model.
//
// The thresholds come from TestCalibrateFaceThresholds, which translates the error
// budget of the shipped FaceNet configuration into each model's distance scale. The
// matching pair is taken at a tenth of that budget, where every ONNX model still beats
// FaceNet's current true accept rate, and rounded down so the measured budget is never
// exceeded. FaceNet keeps the values PhotoPrism has shipped, because changing them
// would alter matching for every existing library on upgrade.
//
// CollisionDist and Epsilon are not measured separately. They are a floor below which two
// vectors count as indistinguishable and the slack added to that check, so they follow the
// width of the model's distance scale rather than an error budget:
//
//	value = FaceNet value * (model ClusterDist / ClusterDistDefault), rounded to three decimals
var EmbeddingModels = map[ModelName]*EmbeddingModel{
	ModelFaceNet: {
		Name:          ModelFaceNet,
		Runtime:       RuntimeTensorFlow,
		Dir:           "facenet",
		Dims:          512,
		Alignment:     AlignBox,
		ClusterDist:   ClusterDistDefault,
		ClusterRadius: ClusterRadiusDefault,
		MatchDist:     MatchDistDefault,
		CollisionDist: CollisionDistDefault,
		Epsilon:       EpsilonDefault,
	},
	ModelSFace: {
		Name:      ModelSFace,
		Runtime:   RuntimeONNX,
		Dir:       "sface",
		Dims:      128,
		Alignment: AlignArcFace5,
		ONNX: &onnx.ModelInfo{
			File:    "face_recognition_sface_2021dec.onnx",
			SHA256:  "0ba9fbfa01b5270c96627c4ef784da859931e02f04419c829e83484087c34e79",
			License: LicenseApache2,
			Input:   alignedCropInput(onnx.Uniform(0, 1)),
			Output:  &onnx.Output{Width: 128},
		},
		ClusterDist:   0.91,
		ClusterRadius: 0.67,
		MatchDist:     0.39,
		CollisionDist: 0.071,
		Epsilon:       0.014,
	},
	ModelAuraFace: {
		Name:      ModelAuraFace,
		Runtime:   RuntimeONNX,
		Dir:       "auraface",
		Dims:      512,
		Alignment: AlignArcFace5,
		ONNX: &onnx.ModelInfo{
			File:    "auraface_v1_glintr100.onnx",
			SHA256:  "a7933ea5330113b01c9b60351d8f4c33003f145d8470ac5f0e52ee2effe25c60",
			License: LicenseApache2,
			Input:   alignedCropInput(onnx.Uniform(127.5, 127.5)),
			Output:  &onnx.Output{Width: 512},
		},
		ClusterDist:   0.98,
		ClusterRadius: 0.76,
		MatchDist:     0.35,
		CollisionDist: 0.077,
		Epsilon:       0.015,
	},
	ModelArcFaceR50: {
		Name:      ModelArcFaceR50,
		Runtime:   RuntimeONNX,
		Dir:       "arcface",
		Dims:      512,
		Alignment: AlignArcFace5,
		ONNX: &onnx.ModelInfo{
			File:    "w600k_r50.onnx",
			SHA256:  "4c06341c33c2ca1f86781dab0e829f88ad5b64be9fba56e56bc9ebdefc619e43",
			License: LicenseResearchOnly,
			Input:   alignedCropInput(onnx.Uniform(127.5, 127.5)),
			Output:  &onnx.Output{Width: 512},
		},
		ClusterDist:   1.07,
		ClusterRadius: 0.67,
		MatchDist:     0.55,
		CollisionDist: 0.084,
		Epsilon:       0.017,
	},
	ModelArcFaceMBF: {
		Name:      ModelArcFaceMBF,
		Runtime:   RuntimeONNX,
		Dir:       "arcface",
		Dims:      512,
		Alignment: AlignArcFace5,
		ONNX: &onnx.ModelInfo{
			File:    "w600k_mbf.onnx",
			SHA256:  "9cc6e4a75f0e2bf0b1aed94578f144d15175f357bdc05e815e5c4a02b319eb4f",
			License: LicenseResearchOnly,
			Input:   alignedCropInput(onnx.Uniform(127.5, 127.5)),
			Output:  &onnx.Output{Width: 512},
		},
		ClusterDist:   1.03,
		ClusterRadius: 0.64,
		MatchDist:     0.49,
		CollisionDist: 0.080,
		Epsilon:       0.016,
	},
}

// alignedCropInput returns the input description shared by every model that consumes the
// standard 112x112 five-point aligned crop (AlignArcFace5), which differ from one another
// only in normalization. It describes the crop, not the ArcFace weights: SFace and AuraFace
// take the same geometry.
func alignedCropInput(normalization onnx.Normalization) *onnx.Input {
	return &onnx.Input{
		Width:         ArcFaceTemplateSize,
		Height:        ArcFaceTemplateSize,
		Layout:        onnx.LayoutNCHW,
		ColorOrder:    onnx.RGB,
		Normalization: normalization,
		Resize:        onnx.Resize{Mode: onnx.ResizeCenterCrop},
	}
}

// AutoModelPreference lists embedding models in the order that ModelAuto prefers them.
//
// This list only decides what a library with no face vectors starts out with. An
// existing library keeps whatever model produced its vectors, because resolving "auto"
// away from it would leave every stored cluster incomparable - see Config.FaceModel.
var AutoModelPreference = []ModelName{ModelSFace, ModelFaceNet, ModelAuraFace, ModelArcFaceR50, ModelArcFaceMBF}

// NormalizeModelName lowercases a model name and accepts hyphens in place of underscores.
func NormalizeModelName(s string) ModelName {
	return strings.ReplaceAll(strings.ToLower(strings.TrimSpace(s)), "-", "_")
}

// ModelsComparable reports whether vectors recorded under stored may be compared with current.
// Legacy vectors may compare with FaceNet or each other because it was the only bundled model.
func ModelsComparable(stored, current ModelName) bool {
	stored = NormalizeModelName(stored)
	current = NormalizeModelName(current)

	if current == "" {
		return stored == ""
	}

	return stored == current || stored == "" && current == ModelFaceNet
}

// SameEmbeddingSpace reports whether two stored vectors may be compared with each other.
// ModelsComparable treats its second argument as the reference, which is wrong when both
// names come from storage, so a legacy blank has to mean FaceNet on either side.
func SameEmbeddingSpace(a, b ModelName) bool {
	return ModelsComparable(a, b) || ModelsComparable(b, a)
}

// ParseModelName returns the supported model name matching s, or ModelAuto when unknown.
func ParseModelName(s string) ModelName {
	name := NormalizeModelName(s)

	switch name {
	case ModelAuto, ModelNone:
		return name
	}

	if _, found := EmbeddingModels[name]; found {
		return name
	}

	return ModelAuto
}

// KnownModelName reports whether s names a supported embedding model, "auto", or "none".
func KnownModelName(s string) bool {
	name := NormalizeModelName(s)

	switch name {
	case ModelAuto, ModelNone:
		return true
	}

	_, found := EmbeddingModels[name]

	return found
}

// FindEmbeddingModel returns the model registered under the specified name, or nil when unknown.
func FindEmbeddingModel(name ModelName) *EmbeddingModel {
	return EmbeddingModels[NormalizeModelName(name)]
}

// EmbeddingModelNames returns the names of all supported embedding models in alphabetical order.
func EmbeddingModelNames() []ModelName {
	names := make([]ModelName, 0, len(EmbeddingModels))

	for name := range EmbeddingModels {
		names = append(names, name)
	}

	sort.Strings(names)

	return names
}

// ModelUsageString lists the accepted FACE_MODEL values for use in CLI help text.
// It is generated from the registry so the help can never advertise a model that
// has been renamed or removed.
func ModelUsageString() string {
	return strings.Join(append([]ModelName{ModelAuto, ModelNone}, EmbeddingModelNames()...), ", ")
}

// FilePath returns the absolute model path within the specified models directory.
// TensorFlow SavedModel bundles are loaded from their directory, so entries without
// an ONNX artifact resolve to the directory itself.
func (m *EmbeddingModel) FilePath(modelsPath string) string {
	if m == nil || modelsPath == "" {
		return ""
	}

	if m.ONNX == nil {
		return filepath.Join(modelsPath, m.Dir)
	}

	return m.ONNX.FilePath(filepath.Join(modelsPath, m.Dir))
}

// Installed reports whether the model files exist in the specified models directory.
func (m *EmbeddingModel) Installed(modelsPath string) bool {
	if m == nil {
		return false
	}

	path := m.FilePath(modelsPath)

	if path == "" {
		return false
	}

	if m.ONNX == nil {
		return fs.PathExists(path)
	}

	return fs.FileExists(path)
}

// WeightLicense returns the license of the pretrained weights. FaceNet is the only
// TensorFlow entry and its weight provenance has never been verified, so it reports
// unknown rather than claiming a license it may not have.
func (m *EmbeddingModel) WeightLicense() string {
	switch {
	case m == nil:
		return ""
	case m.ONNX != nil && m.ONNX.License != "":
		return m.ONNX.License
	default:
		return LicenseUnknown
	}
}

// InputSize returns the crop width and height the model expects, or zero for models
// whose input geometry is not described in the registry.
func (m *EmbeddingModel) InputSize() (width, height int) {
	if m == nil {
		return 0, 0
	}

	return m.ONNX.InputSize()
}

// Aligned reports whether the model requires landmark-aligned crops.
func (m *EmbeddingModel) Aligned() bool {
	return m != nil && m.Alignment == AlignArcFace5
}

// String returns the model name, or "none" for nil receivers.
func (m *EmbeddingModel) String() string {
	if m == nil {
		return ModelNone
	}

	return m.Name
}

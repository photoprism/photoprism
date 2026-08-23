package face

import (
	"path/filepath"

	"github.com/photoprism/photoprism/internal/ai/onnx"
	"github.com/photoprism/photoprism/pkg/fs"
)

// DecodeKind names the way a detector's output tensors are turned into faces. It is what
// actually differs between detectors: the preprocessing and the graph geometry are read from
// the artifact, but no two detection heads agree on how to read their own predictions.
type DecodeKind string

const (
	// DecodeSCRFD is anchor-based with two priors per cell and one score output per stride.
	DecodeSCRFD DecodeKind = "scrfd"
	// DecodeYuNet is anchor-free with one prior per cell, scoring as sqrt(cls x obj).
	DecodeYuNet DecodeKind = "yunet"
)

// Detector describes a face detection model: where its weights live, how an image must be
// prepared for it, and which decoder reads its output.
type Detector struct {
	Name   DetectorName
	Dir    string
	Decode DecodeKind
	ONNX   *onnx.ModelInfo
}

// DetectorName identifies a detection model.
type DetectorName = string

const (
	// DetectorYuNet is the permissively licensed default.
	DetectorYuNet DetectorName = "yunet"
	// DetectorSCRFD is the InsightFace detector, kept selectable for comparison.
	DetectorSCRFD DetectorName = "scrfd"
)

// Detectors registers the detection models this build can run.
//
// YuNet is listed first because that is the order Config resolution walks: it is MIT at the
// weights layer as well as the code layer, which SCRFD is not, and it measures within noise of
// SCRFD on ordinary libraries while finding more faces in group photographs.
var Detectors = []*Detector{
	{
		Name:   DetectorYuNet,
		Dir:    "yunet",
		Decode: DecodeYuNet,
		ONNX: &onnx.ModelInfo{
			File:    "face_detection_yunet_2026may.onnx",
			SHA256:  "ebafce4e3c118d6554634be5c27ab333b4c047a9a8c3faf1d7cf93101c22f0f0",
			License: "MIT",
			Input: &onnx.Input{
				Width:      640,
				Height:     640,
				Layout:     onnx.LayoutNCHW,
				ColorOrder: onnx.BGR,
				// YuNet consumes raw 0-255 values: OpenCV hands FaceDetectorYN a blob built
				// with the default scale factor and no mean, and the graph expects that.
				Normalization: onnx.Uniform(0, 1),
				Resize:        onnx.Resize{Mode: onnx.ResizePad},
			},
		},
	},
	{
		Name:   DetectorSCRFD,
		Dir:    "scrfd",
		Decode: DecodeSCRFD,
		ONNX: &onnx.ModelInfo{
			File:   DefaultONNXModelFilename,
			SHA256: "ae72185653e279aa2056b288662a19ec3519ced5426d2adeffbe058a86369a24",
			Input: &onnx.Input{
				Width:         640,
				Height:        640,
				Layout:        onnx.LayoutNCHW,
				ColorOrder:    onnx.RGB,
				Normalization: onnx.Uniform(127.5, 128),
				Resize:        onnx.Resize{Mode: onnx.ResizePad},
			},
		},
	},
}

// FindDetector returns the registered detector with the given name, or nil.
func FindDetector(name DetectorName) *Detector {
	for _, d := range Detectors {
		if d.Name == name {
			return d
		}
	}

	return nil
}

// DetectorForFile returns the detector whose artifact the path names.
//
// Matching is by file name rather than by checksum so that an operator who re-exported a model
// still gets its preprocessing: the alternative is applying another detector's channel order and
// normalization, which produces detections that are subtly wrong rather than an error.
func DetectorForFile(fileName string) *Detector {
	base := filepath.Base(fileName)

	for _, d := range Detectors {
		if d.ONNX != nil && d.ONNX.File == base {
			return d
		}
	}

	// The SCRFD installer also produced a differently named fixed-shape export.
	if base == "scrfd_500m_bnkps_shape640x640.onnx" {
		return FindDetector(DetectorSCRFD)
	}

	return nil
}

// Installed reports whether this detector's weights are present under the models path.
func (d *Detector) Installed(modelsPath string) bool {
	return d != nil && d.ONNX != nil && fs.FileExists(d.Path(modelsPath))
}

// Path returns the absolute path to this detector's weights.
func (d *Detector) Path(modelsPath string) string {
	if d == nil || d.ONNX == nil {
		return ""
	}

	return filepath.Join(modelsPath, d.Dir, d.ONNX.File)
}

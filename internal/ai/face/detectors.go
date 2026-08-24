package face

import (
	"path/filepath"
	"slices"
	"strings"

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
//
// Legacy names the same weights were installed under before, so a copy an operator already
// holds still resolves to its own preprocessing instead of the default detector's.
type Detector struct {
	Name     DetectorName
	Dir      string
	Decode   DecodeKind
	MinScore float32
	// ClusterMinScore is the higher bar a face must clear to contribute to automatic clustering,
	// on the 0-100 scale. It follows MinScore rather than being shared, because a value that is a
	// meaningful step above one detector's cutoff is below another's and gates nothing.
	ClusterMinScore int
	ONNX            *onnx.ModelInfo
	Legacy          []string
}

// DetectorName identifies a detection model.
type DetectorName = string

const (
	// DetectorAuto derives the detector from the configured embedding model. It is derived on
	// every start rather than resolved once and recorded, which is what tells it apart from the
	// embedding model's ModelDetect.
	DetectorAuto DetectorName = "auto"
	// DetectorDetect is an accepted spelling of DetectorAuto.
	DetectorDetect DetectorName = "detect"
	// DetectorNone disables face detection.
	DetectorNone DetectorName = "none"
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
		// Calibrated on our own corpus rather than adopted from another detector: YuNet scores
		// as sqrt(cls x obj) and is not a single calibrated sigmoid, so it needs its own cutoff.
		// 0.65 is where its false positives on non-faces stop while it still finds more in a
		// group photograph than SCRFD does.
		MinScore:        0.65,
		ClusterMinScore: 70,
		ONNX: &onnx.ModelInfo{
			File:    "face_detection_yunet_2026may.onnx",
			SHA256:  "ebafce4e3c118d6554634be5c27ab333b4c047a9a8c3faf1d7cf93101c22f0f0",
			License: "MIT",
			Input: &onnx.Input{
				Width:      640,
				Height:     640,
				Layout:     onnx.LayoutNCHW,
				ColorOrder: onnx.BGR,
				// Raw 0-255 BGR: OpenCV builds the detector's blob with every default, where
				// the SFace one sets swapRB. The two are opposite and neither is inferable.
				Normalization: onnx.Uniform(0, 1),
				Resize:        onnx.Resize{Mode: onnx.ResizePad},
			},
		},
	},
	{
		Name:            DetectorSCRFD,
		Dir:             "scrfd",
		Decode:          DecodeSCRFD,
		MinScore:        0.50,
		ClusterMinScore: 60,
		ONNX: &onnx.ModelInfo{
			// The publisher's own artifact, which is where an opt-in install fetches from. Its
			// input is dynamic where our earlier re-export was fixed, and it is otherwise the
			// same weights: both produce identical boxes and landmarks at a 640 input.
			File:    "det_500m.onnx",
			SHA256:  "5e4447f50245bbd7966bd6c0fa52938c61474a04ec7def48753668a9d8b4ea3a",
			License: LicenseNonFree,
			Input: &onnx.Input{
				Width:         640,
				Height:        640,
				Layout:        onnx.LayoutNCHW,
				ColorOrder:    onnx.RGB,
				Normalization: onnx.Uniform(127.5, 128),
				Resize:        onnx.Resize{Mode: onnx.ResizePad},
			},
		},
		Legacy: []string{"scrfd.onnx", "scrfd_500m_bnkps_shape640x640.onnx"},
	},
}

// FindDetector returns the registered detector with the given name, or nil.
func FindDetector(name DetectorName) *Detector {
	name = NormalizeDetectorName(name)

	for _, d := range Detectors {
		if d.Name == name {
			return d
		}
	}

	return nil
}

// NormalizeDetectorName lowercases a detector name and accepts hyphens in place of underscores.
func NormalizeDetectorName(s string) DetectorName {
	return strings.ReplaceAll(strings.ToLower(strings.TrimSpace(s)), "-", "_")
}

// ParseDetectorName returns the registered detector name matching s, or DetectorAuto when the
// value is empty, asks for derivation, or is not recognized. Use KnownDetectorName to tell an
// unknown value apart from a request to derive one.
func ParseDetectorName(s string) DetectorName {
	switch name := NormalizeDetectorName(s); name {
	case "", DetectorAuto, DetectorDetect:
		return DetectorAuto
	case DetectorNone:
		return name
	default:
		if FindDetector(name) != nil {
			return name
		}

		return DetectorAuto
	}
}

// KnownDetectorName reports whether s names a registered detector, asks for derivation, or
// disables detection.
func KnownDetectorName(s string) bool {
	switch name := NormalizeDetectorName(s); name {
	case "", DetectorAuto, DetectorDetect, DetectorNone:
		return true
	default:
		return FindDetector(name) != nil
	}
}

// DetectorsComparable reports whether a crop recorded under stored may be treated as one the
// current detector would place, so that a marker holding it needs no new detection.
//
// Only an exact match qualifies: a blank or engine-level value names no detector, and two
// detectors place different landmarks and therefore a different crop from the same face.
// A current name that is empty or disables detection compares equal to everything, because
// there is then no detector to disagree with.
func DetectorsComparable(stored, current DetectorName) bool {
	stored = NormalizeDetectorName(stored)
	current = NormalizeDetectorName(current)

	if current == "" || current == DetectorNone {
		return true
	}

	return stored == current
}

// DetectorUsageString lists the accepted FACE_DETECTOR values for use in CLI help text.
//
// It leaves out detectors whose weights may only be used after their publisher's terms have
// been accepted, because help text is read as an offer.
func DetectorUsageString() string {
	names := []DetectorName{DetectorAuto, DetectorNone}

	for _, d := range Detectors {
		if !d.LicenseGated() {
			names = append(names, d.Name)
		}
	}

	return strings.Join(names, ", ")
}

// DefaultDetector returns the detector a build runs when nothing selects one.
// It is the first registered entry whose weights may be redistributed, so a second list
// cannot name a detector no image contains.
func DefaultDetector() *Detector {
	for _, d := range Detectors {
		if !d.LicenseGated() {
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

	for _, d := range Detectors {
		if slices.Contains(d.Legacy, base) {
			return d
		}
	}

	return nil
}

// Installed reports whether this detector's weights are present under the models path.
func (d *Detector) Installed(modelsPath string) bool {
	return d.InstalledPath(modelsPath) != ""
}

// InstalledPath returns the path the weights are actually at, or an empty string when they
// are not installed. It reports a legacy name only when the current artifact is absent, so
// an install that holds both loads the one the registry describes.
func (d *Detector) InstalledPath(modelsPath string) string {
	if path := d.Path(modelsPath); path != "" && fs.FileExists(path) {
		return path
	} else if d == nil {
		return ""
	}

	for _, name := range d.Legacy {
		if path := filepath.Join(modelsPath, d.Dir, name); fs.FileExists(path) {
			return path
		}
	}

	return ""
}

// Path returns the absolute path to the artifact the registry describes, whether or not it
// is installed, so a caller can report which detector a build would have loaded.
func (d *Detector) Path(modelsPath string) string {
	if d == nil || d.ONNX == nil {
		return ""
	}

	return filepath.Join(modelsPath, d.Dir, d.ONNX.File)
}

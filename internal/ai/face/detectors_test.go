package face

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/ai/onnx"
	"github.com/photoprism/photoprism/pkg/fs"
)

func TestFindDetector(t *testing.T) {
	assert.Equal(t, DetectorYuNet, FindDetector(DetectorYuNet).Name)
	assert.Equal(t, DecodeYuNet, FindDetector(DetectorYuNet).Decode)
	assert.Equal(t, DecodeSCRFD, FindDetector(DetectorSCRFD).Decode)
	assert.Nil(t, FindDetector("nonexistent"))
}

func TestNormalizeDetectorName(t *testing.T) {
	assert.Equal(t, DetectorYuNet, NormalizeDetectorName(" YuNet "))
	assert.Equal(t, "arc_face", NormalizeDetectorName("Arc-Face"))
	assert.Equal(t, "", NormalizeDetectorName(""))
}

func TestParseDetectorName(t *testing.T) {
	t.Run("Registered", func(t *testing.T) {
		assert.Equal(t, DetectorYuNet, ParseDetectorName("YuNet"))
		assert.Equal(t, DetectorSCRFD, ParseDetectorName(DetectorSCRFD))
	})
	t.Run("Derive", func(t *testing.T) {
		assert.Equal(t, DetectorAuto, ParseDetectorName(""))
		assert.Equal(t, DetectorAuto, ParseDetectorName(DetectorAuto))
		assert.Equal(t, DetectorAuto, ParseDetectorName(DetectorDetect), "detect is an accepted spelling")
	})
	t.Run("None", func(t *testing.T) {
		assert.Equal(t, DetectorNone, ParseDetectorName("None"))
	})
	t.Run("Unknown", func(t *testing.T) {
		// An unknown value asks for derivation, and KnownDetectorName is what tells the two
		// apart, so a typo does not silently disable detection.
		assert.Equal(t, DetectorAuto, ParseDetectorName("nonexistent"))
		assert.False(t, KnownDetectorName("nonexistent"))
	})
}

func TestKnownDetectorName(t *testing.T) {
	for _, name := range []string{"", DetectorAuto, DetectorDetect, DetectorNone, DetectorYuNet, "SCRFD"} {
		assert.True(t, KnownDetectorName(name), name)
	}
	assert.False(t, KnownDetectorName("pigo"))
}

func TestDetectorsComparable(t *testing.T) {
	t.Run("SameDetector", func(t *testing.T) {
		assert.True(t, DetectorsComparable(DetectorYuNet, DetectorYuNet))
		assert.True(t, DetectorsComparable("YuNet", DetectorYuNet))
	})
	t.Run("OtherDetector", func(t *testing.T) {
		assert.False(t, DetectorsComparable(DetectorSCRFD, DetectorYuNet))
		assert.False(t, DetectorsComparable(DetectorYuNet, DetectorSCRFD))
	})
	t.Run("NoRecordedDetector", func(t *testing.T) {
		// A blank names no detector and "onnx" names only the runtime, so neither can be shown
		// to agree with the crop the current detector would place.
		assert.False(t, DetectorsComparable("", DetectorYuNet))
		assert.False(t, DetectorsComparable(string(EngineONNX), DetectorYuNet))
	})
	t.Run("NoCurrentDetector", func(t *testing.T) {
		// Nothing is running to disagree with, so this must not report every stored crop as
		// belonging to another detector.
		assert.True(t, DetectorsComparable(DetectorSCRFD, ""))
		assert.True(t, DetectorsComparable(DetectorSCRFD, DetectorNone))
		assert.True(t, DetectorsComparable("", ""))
	})
}

func TestDetectorUsageString(t *testing.T) {
	usage := DetectorUsageString()

	assert.Equal(t, "auto, yunet, none", usage)
	// Help text is read as an offer, so weights that need their publisher's terms accepted
	// are not listed.
	assert.NotContains(t, usage, DetectorSCRFD)
}

func TestDefaultDetector(t *testing.T) {
	d := DefaultDetector()
	require.NotNil(t, d)
	assert.Equal(t, DetectorYuNet, d.Name)
	assert.False(t, d.LicenseGated(), "a build may only default to weights it may redistribute")
	assert.True(t, d.Advertise, "the default has to be one the product offers")
}

// TestDetectorDefault pins that exactly one detector is the default. It used to be whichever
// redistributable one came first, which made the registry order load-bearing without saying so.
func TestDetectorDefault(t *testing.T) {
	defaults := 0

	for _, d := range Detectors {
		if d.Default {
			defaults++
		}
	}

	assert.Equal(t, 1, defaults, "exactly one detector may be the default")
}

// TestDetectorDisplayNames pins that every registered detector has a human-readable name, so a
// report never falls back to the identifier for one that is shipped.
func TestDetectorDisplayNames(t *testing.T) {
	seen := make(map[string]DetectorName, len(Detectors))

	for _, d := range Detectors {
		assert.NotEmpty(t, d.DisplayName, d.Name)

		if other, dup := seen[d.DisplayName]; dup {
			t.Errorf("%s and %s share the display name %q", d.Name, other, d.DisplayName)
		}

		seen[d.DisplayName] = d.Name
	}

	t.Run("Registered", func(t *testing.T) {
		assert.Equal(t, FindDetector(DetectorYuNet).DisplayName, DetectorDisplayName(DetectorYuNet))
	})
	t.Run("FallsBackToTheIdentifier", func(t *testing.T) {
		assert.Equal(t, "nonexistent", DetectorDisplayName("nonexistent"))
	})
}

// TestDetectorInstallers pins the registry to the scripts that install the weights, and with it
// the guarantee a fresh checkout depends on: "make dep-models" has to install the default
// detector. Without one, detection resolves to nothing and every picture is indexed as holding
// nobody rather than failing.
func TestDetectorInstallers(t *testing.T) {
	t.Run("Default", func(t *testing.T) {
		d := DefaultDetector()
		require.NotNil(t, d)

		// Registry fields: name|url|fallback|sha256|type|dir|file
		data := readModelScript(t, filepath.Join("dist", "download-models.sh"))
		entry := regexp.MustCompile(`(?m)^` + regexp.QuoteMeta(d.Dir) + `\|.*$`).FindString(data)
		require.NotEmpty(t, entry, "download-models.sh must list %s", d.Name)

		fields := strings.Split(entry, "|")
		require.Len(t, fields, 7, "%s must have all registry fields", d.Name)
		assert.Equal(t, d.ONNX.SHA256, fields[3])
		assert.Equal(t, d.Dir, fields[5])
		assert.Equal(t, d.ONNX.File, fields[6])

		assert.Contains(t, depModelsRecipe(t), " "+d.Dir, "make dep-models must install the default detector")
	})
	t.Run("GatedWeightsStayOutOfTheBuild", func(t *testing.T) {
		recipe := depModelsRecipe(t)

		for _, d := range Detectors {
			if !d.LicenseGated() {
				continue
			}

			assert.NotContains(t, recipe, d.Dir, "make dep-models must not install %s", d.Name)
		}
	})
	t.Run("SCRFD", func(t *testing.T) {
		d := FindDetector(DetectorSCRFD)
		require.NotNil(t, d)

		data := readModelScript(t, filepath.Join("dist", "download-scrfd.sh"))

		assert.Contains(t, data, `MODEL_SHA256="`+d.ONNX.SHA256+`"`, "the installer must pin the registered checksum")
		assert.Contains(t, data, `MODEL_ENTRY="`+d.ONNX.File+`"`, "the installer must install the registered artifact")
		// Gated weights are fetched from their publisher after an explicit acceptance, so the
		// installer must state both rather than being reachable by running it.
		assert.Contains(t, data, LicenseAcceptanceVar)
		assert.Contains(t, data, "https://github.com/deepinsight/insightface/releases/")
	})
}

// depModelsRecipe returns the "dep-models" recipe from the Makefile, which is what decides
// which weights a development build and therefore a published image contains.
func depModelsRecipe(t *testing.T) string {
	t.Helper()

	data, err := os.ReadFile(filepath.Join("..", "..", "..", "Makefile")) //nolint:gosec // G304: fixed repository path.

	if err != nil {
		t.Skip("faces: skipping, the Makefile is not available")
	}

	recipe := regexp.MustCompile(`(?m)^dep-models:\n(?:\t.*\n)+`).FindString(string(data))
	require.NotEmpty(t, recipe, "the Makefile must define dep-models")

	return recipe
}

func TestDetectorForFile(t *testing.T) {
	t.Run("YuNet", func(t *testing.T) {
		d := DetectorForFile("/models/yunet/face_detection_yunet_2026may.onnx")
		require.NotNil(t, d)
		assert.Equal(t, DecodeYuNet, d.Decode)
		// Raw 0-255 BGR: applying the other detector's normalization would not fail, it would
		// quietly produce worse detections, which is why the mapping is asserted.
		assert.Equal(t, onnx.BGR, d.ONNX.Input.ColorOrder)
		assert.Equal(t, float32(0), d.ONNX.Input.Normalization.Mean[0])
	})
	t.Run("SCRFD", func(t *testing.T) {
		d := DetectorForFile("/models/scrfd/det_500m.onnx")
		require.NotNil(t, d)
		assert.Equal(t, DecodeSCRFD, d.Decode)
		assert.Equal(t, onnx.RGB, d.ONNX.Input.ColorOrder)
	})
	t.Run("LegacyNames", func(t *testing.T) {
		for _, name := range FindDetector(DetectorSCRFD).Legacy {
			d := DetectorForFile(filepath.Join("/models/scrfd", name))
			require.NotNil(t, d, name)
			assert.Equal(t, DecodeSCRFD, d.Decode, name)
		}
	})
	t.Run("Unknown", func(t *testing.T) {
		assert.Nil(t, DetectorForFile("/models/other/whatever.onnx"))
	})
}

func TestYuNetFeatDim(t *testing.T) {
	// A multiple of 32 agrees with division, and 720 is the case that does not.
	assert.Equal(t, 80, yunetFeatDim(640, 8))
	assert.Equal(t, 20, yunetFeatDim(640, 32))
	assert.Equal(t, 23, yunetFeatDim(720, 32), "the halving chain rounds up")
}

// TestYuNetEngineLive runs the shipped YuNet path end to end against the installed weights.
func TestYuNetEngineLive(t *testing.T) {
	restoreEngine(t)

	detector := FindDetector(DetectorYuNet)
	require.NotNil(t, detector)

	modelPath := detector.Path("../../../assets/models")

	if _, err := os.Stat(modelPath); err != nil {
		t.Skipf("yunet is not installed (%s)", err)
	}

	require.NoError(t, ConfigureEngine(EngineSettings{
		Name: EngineONNX,
		ONNX: ONNXOptions{ModelPath: modelPath, Threads: 2},
	}))

	faces, err := Detect("testdata/1.jpg", SizeThreshold)
	require.NoError(t, err)
	require.NotEmpty(t, faces, "yunet must find the face in the test image")

	f := faces[0]
	t.Logf("yunet: %d face(s), first score %d size %d", len(faces), f.Score, f.Size())

	// A decoder that mismapped its outputs still produces boxes, so the landmarks are what
	// prove it: they must be complete and fit the template.
	pts, ok := f.AlignPoints()
	require.True(t, ok, "all five landmarks must be present")
	assert.Less(t, pts[0][0], pts[1][0], "the first eye must be the image-left one")

	img, _, decodeErr := fs.DecodeImageFile("testdata/1.jpg")
	require.NoError(t, decodeErr)

	_, alignErr := AlignedCrop(img, &f, 112, 112)
	assert.NoError(t, alignErr, "the landmarks must fit the ArcFace template")

	// Provenance names the detector, not the runtime. Every detector runs on ONNX, so the
	// engine name would distinguish nothing and must not reach the column.
	assert.Equal(t, DetectorYuNet, f.DetectModel)
	assert.NotEqual(t, EngineONNX, f.DetectModel)
}

// TestDetectorMinScore pins the score cutoff per detector. Detectors do not score alike, so a
// value adopted from another is not calibration: YuNet scores as sqrt(cls x obj) rather than as
// one sigmoid, and at SCRFD's 0.50 it accepts flowers as faces.
func TestDetectorMinScore(t *testing.T) {
	for _, d := range Detectors {
		// Both bars are on the 0-100 scale markers and the FACE_* options use, so a value that
		// looks like the 0-1 one the decoder reports is a unit mistake rather than a low cutoff.
		assert.Positive(t, d.MinScore, d.Name)
		assert.LessOrEqual(t, d.MinScore, 100, d.Name)
		assert.Positive(t, d.ClusterScore, d.Name)
		assert.LessOrEqual(t, d.ClusterScore, 100, d.Name)
		assert.Positive(t, d.MigrateScore, d.Name)
		assert.LessOrEqual(t, d.MigrateScore, d.MinScore, d.Name)
	}

	assert.NotEqual(t, FindDetector(DetectorSCRFD).MinScore, FindDetector(DetectorYuNet).MinScore,
		"a cutoff copied from the other detector is not a calibrated one")
}

// TestDetectorMigrateScore pins the migration floor to the detector, because a migration makes the
// opposite trade to an index: a miss discards a curated marker rather than adding a false positive.
func TestDetectorMigrateScore(t *testing.T) {
	t.Run("Registered", func(t *testing.T) {
		assert.InDelta(t, float64(FindDetector(DetectorYuNet).MigrateScore), DetectorMigrateScore(DetectorYuNet), 0.5)
		assert.InDelta(t, float64(FindDetector(DetectorSCRFD).MigrateScore), DetectorMigrateScore(DetectorSCRFD), 0.5)
	})
	t.Run("Unregistered", func(t *testing.T) {
		// A name nothing registers must still yield a floor some detector enforces, or a migration
		// would run at zero and re-embed whatever the decoder emits.
		assert.Equal(t, DetectorMigrateScore(DefaultDetectorName()), DetectorMigrateScore("nonexistent"))
		assert.Positive(t, DetectorMigrateScore(DetectorNone))
	})
	t.Run("Default", func(t *testing.T) {
		assert.Equal(t, DetectorMigrateScore(DefaultDetectorName()), DefaultDetectorMigrateScore())
	})
}

// TestClusterScore pins that the clustering bar follows the detector that produced a marker, not
// the one in force. A library holds markers from more than one, and nothing recomputes a score, so
// judging an old marker by a newer detector's calibration would exclude it permanently.
func TestClusterScore(t *testing.T) {
	restore := ClusterScoreThreshold
	t.Cleanup(func() { ClusterScoreThreshold = restore })
	ClusterScoreThreshold = 0

	assert.Equal(t, FindDetector(DetectorYuNet).ClusterScore, ClusterScore(DetectorYuNet))
	assert.Equal(t, FindDetector(DetectorSCRFD).ClusterScore, ClusterScore(DetectorSCRFD))
	assert.NotEqual(t, ClusterScore(DetectorYuNet), ClusterScore(DetectorSCRFD),
		"a bar shared between detectors gates nothing for one of them")

	// Everything written before the provenance column existed keeps the shared default, so an
	// upgrade strands nothing.
	assert.Equal(t, ClusterScoreThresholdDefault, ClusterScore(""))
	assert.Equal(t, ClusterScoreThresholdDefault, ClusterScore("centerface"))
	assert.Less(t, ClusterScoreThresholdDefault, ClusterScore(DetectorSCRFD))

	// FACE_CLUSTER_SCORE outranks the per-detector bars. It was assigned by Propagate and read by
	// nothing once the bar became per detector, so the option configured nothing at all - the
	// same defect FACE_SCORE had in the detection path.
	t.Run("Configured", func(t *testing.T) {
		// Safe where taking the active detector's bar is not: it is a choice rather than a
		// calibration a marker was never scored against.
		ClusterScoreThreshold = 55
		assert.Equal(t, 55, ClusterScore(DetectorYuNet))
		assert.Equal(t, 55, ClusterScore(DetectorSCRFD))
		assert.Equal(t, 55, ClusterScore(""))
	})
	t.Run("Disabled", func(t *testing.T) {
		ClusterScoreThreshold = -1
		assert.Zero(t, ClusterScore(DetectorYuNet))
	})
}

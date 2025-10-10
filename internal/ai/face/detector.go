package face

import (
	_ "embed"
	"fmt"
	_ "image/jpeg"
	"os"
	"path/filepath"
	"runtime/debug"
	"sort"

	pigo "github.com/esimov/pigo/core"

	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
)

//go:embed cascade/facefinder
var cascadeFile []byte

//go:embed cascade/puploc
var puplocFile []byte

var (
	classifier *pigo.Pigo
	plc        *pigo.PuplocCascade
	flpcs      map[string][]*FlpCascade
)

// DefaultAngles contains the canonical detection angles in radians.
var DefaultAngles = []float64{-0.3, 0, 0.3}

// DetectionAngles holds the active detection angles configured at runtime.
var DetectionAngles = append([]float64(nil), DefaultAngles...)

func init() {
	var err error

	p := pigo.NewPigo()
	// Unpack the binary file. This will return the number of cascade trees,
	// the tree depth, the threshold and the prediction from tree's leaf nodes.
	classifier, err = p.Unpack(cascadeFile)

	if err != nil {
		log.Errorf("faces: %s", err)
	}

	pl := pigo.NewPuplocCascade()
	plc, err = pl.UnpackCascade(puplocFile)

	if err != nil {
		log.Errorf("faces: %s", err)
	}

	flpcs, err = ReadCascadeDir(pl, "cascade/lps")

	if err != nil {
		log.Errorf("faces: %s", err)
	}
}

var (
	eyeCascades   = []string{"lp46", "lp44", "lp42", "lp38", "lp312"}
	mouthCascades = []string{"lp93", "lp84", "lp82", "lp81"}
)

// pigoEngine implements DetectionEngine using the bundled Pigo cascades.
type pigoEngine struct{}

// newPigoEngine constructs a Pigo-backed DetectionEngine instance.
func newPigoEngine() *pigoEngine {
	return &pigoEngine{}
}

func (p *pigoEngine) Name() string {
	return EnginePigo
}

// Close releases resources held by the Pigo engine (none at the moment).
func (p *pigoEngine) Close() error {
	return nil
}

// pigoDetector contains Pigo face detector general settings.
type pigoDetector struct {
	minSize       int
	shiftFactor   float64
	scaleFactor   float64
	iouThreshold  float64
	perturb       int
	landmarkAngle float64
	angles        []float64
}

// Detect runs the detection algorithm over the provided source image.
func (p *pigoEngine) Detect(fileName string, findLandmarks bool, minSize int) (faces Faces, err error) {
	defer func() {
		if r := recover(); r != nil {
			log.Errorf("faces: %s (panic)\nstack: %s", r, debug.Stack())
		}
	}()

	if minSize < 20 {
		minSize = 20
	}

	angles := append([]float64(nil), DetectionAngles...)

	d := &pigoDetector{
		minSize:       minSize,
		shiftFactor:   0.1,
		scaleFactor:   1.1,
		iouThreshold:  float64(OverlapThresholdFloor) / 100,
		perturb:       63,
		landmarkAngle: 0.0,
		angles:        angles,
	}

	if !fs.FileExists(fileName) {
		return faces, fmt.Errorf("faces: file '%s' not found", clean.Log(filepath.Base(fileName)))
	}

	det, params, err := d.Detect(fileName)

	if err != nil {
		return faces, fmt.Errorf("faces: %s (detect faces)", err)
	}

	if len(det) == 0 {
		return faces, nil
	}

	faces, err = d.Faces(det, params, findLandmarks)

	if err != nil {
		return faces, fmt.Errorf("faces: %s", err)
	}

	return faces, nil
}

// Detect runs the detection algorithm over the provided source image.
func (d *pigoDetector) Detect(fileName string) (faces []pigo.Detection, params pigo.CascadeParams, err error) {
	if len(d.angles) == 0 {
		// Fallback to defaults when the detector is constructed manually (e.g. tests).
		d.angles = append([]float64(nil), DetectionAngles...)
	}

	file, err := os.Open(fileName)

	if err != nil {
		return faces, params, err
	}

	defer func() {
		if cerr := file.Close(); err == nil && cerr != nil {
			err = cerr
		}
	}()

	src, err := pigo.DecodeImage(file)

	if err != nil {
		return faces, params, err
	}

	pixels := pigo.RgbToGrayscale(src)
	cols, rows := src.Bounds().Max.X, src.Bounds().Max.Y

	var maxSize int

	if cols < 20 || rows < 20 || cols < d.minSize || rows < d.minSize {
		return faces, params, fmt.Errorf("image size %dx%d is too small", cols, rows)
	} else if cols < rows {
		maxSize = cols - 4
	} else {
		maxSize = rows - 4
	}

	imageParams := pigo.ImageParams{
		Pixels: pixels,
		Rows:   rows,
		Cols:   cols,
		Dim:    cols,
	}

	params = pigo.CascadeParams{
		MinSize:     d.minSize,
		MaxSize:     maxSize,
		ShiftFactor: d.shiftFactor,
		ScaleFactor: d.scaleFactor,
		ImageParams: imageParams,
	}

	log.Tracef("faces: image size %dx%d, face size min %d, max %d", cols, rows, params.MinSize, params.MaxSize)

	// Run the classifier over the obtained leaf nodes for each configured angle and merge the results.
	var detections []pigo.Detection
	for _, angle := range d.angles {
		result := classifier.RunCascade(params, angle)
		if len(result) == 0 {
			continue
		}

		detections = append(detections, result...)
	}

	if len(detections) == 0 {
		return detections, params, nil
	}

	// Calculate the intersection over union (IoU) of two clusters.
	faces = classifier.ClusterDetections(detections, d.iouThreshold)

	return faces, params, nil
}

// Faces adds landmark coordinates to detected faces and returns the results.
func (d *pigoDetector) Faces(det []pigo.Detection, params pigo.CascadeParams, findLandmarks bool) (results Faces, err error) {
	// Sort results by size.
	sort.Slice(det, func(i, j int) bool {
		return det[i].Scale > det[j].Scale
	})

	results = make(Faces, 0, len(det))

	for _, face := range det {
		score := face.Q
		scale := face.Scale
		requiredScore := QualityThreshold(scale)
		scaleMin := LandmarkQualityScaleMin
		scaleMax := LandmarkQualityScaleMax
		fallbackCandidate := false
		if !findLandmarks && score < requiredScore && score >= LandmarkQualityFloor && scale >= scaleMin && scale <= scaleMax && requiredScore-score <= LandmarkQualitySlack {
			fallbackCandidate = true
		}

		faceCoord := NewArea(
			"face",
			face.Row,
			face.Col,
			scale,
		)

		var eyesCoords []Area
		var landmarkCoords []Area
		var eyesFound bool

		needLandmarks := (findLandmarks || fallbackCandidate) && scale > 50

		if needLandmarks {
			if findLandmarks {
				eyesCoords = make([]Area, 0, 2)
			}

			scaleF := float32(scale)
			leftCandidate := pigo.Puploc{
				Row:      face.Row - int(0.075*scaleF),
				Col:      face.Col - int(0.175*scaleF),
				Scale:    scaleF * 0.25,
				Perturbs: d.perturb,
			}

			leftEye := plc.RunDetector(leftCandidate, params.ImageParams, d.landmarkAngle, false)
			leftEyeFound := leftEye.Row > 0 && leftEye.Col > 0
			if leftEyeFound && findLandmarks {
				eyesCoords = append(eyesCoords, NewArea(
					"eye_l",
					leftEye.Row,
					leftEye.Col,
					int(leftEye.Scale),
				))
			}

			rightCandidate := pigo.Puploc{
				Row:      face.Row - int(0.075*scaleF),
				Col:      face.Col + int(0.185*scaleF),
				Scale:    scaleF * 0.25,
				Perturbs: d.perturb,
			}

			rightEye := plc.RunDetector(rightCandidate, params.ImageParams, d.landmarkAngle, false)
			rightEyeFound := rightEye.Row > 0 && rightEye.Col > 0
			if rightEyeFound && findLandmarks {
				eyesCoords = append(eyesCoords, NewArea(
					"eye_r",
					rightEye.Row,
					rightEye.Col,
					int(rightEye.Scale),
				))
			}

			if leftEyeFound && rightEyeFound {
				eyesFound = true

				if findLandmarks {
					landmarkCapacity := len(eyeCascades)*2 + len(mouthCascades) + 1
					landmarkCoords = make([]Area, 0, landmarkCapacity)

					for _, eye := range eyeCascades {
						for _, flpc := range flpcs[eye] {
							if flpc == nil {
								continue
							}

							flp := flpc.GetLandmarkPoint(leftEye, rightEye, params.ImageParams, d.perturb, false)
							if flp.Row > 0 && flp.Col > 0 {
								landmarkCoords = append(landmarkCoords, NewArea(
									eye,
									flp.Row,
									flp.Col,
									int(flp.Scale),
								))
							}

							flp = flpc.GetLandmarkPoint(leftEye, rightEye, params.ImageParams, d.perturb, true)
							if flp.Row > 0 && flp.Col > 0 {
								landmarkCoords = append(landmarkCoords, NewArea(
									eye+"_v",
									flp.Row,
									flp.Col,
									int(flp.Scale),
								))
							}
						}
					}

					for _, mouth := range mouthCascades {
						for _, flpc := range flpcs[mouth] {
							if flpc == nil {
								continue
							}

							flp := flpc.GetLandmarkPoint(leftEye, rightEye, params.ImageParams, d.perturb, false)
							if flp.Row > 0 && flp.Col > 0 {
								landmarkCoords = append(landmarkCoords, NewArea(
									"mouth_"+mouth,
									flp.Row,
									flp.Col,
									int(flp.Scale),
								))
							}
						}
					}

					if cascades := flpcs["lp84"]; len(cascades) > 0 {
						if flpc := cascades[0]; flpc != nil {
							flp := flpc.GetLandmarkPoint(leftEye, rightEye, params.ImageParams, d.perturb, true)
							if flp.Row > 0 && flp.Col > 0 {
								landmarkCoords = append(landmarkCoords, NewArea(
									"lp84",
									flp.Row,
									flp.Col,
									int(flp.Scale),
								))
							}
						}
					}
				}
			}
		}

		if eyesFound && fallbackCandidate && requiredScore > LandmarkQualityFloor {
			requiredScore = LandmarkQualityFloor
		}

		if score < requiredScore {
			continue
		}

		// Create face.
		f := Face{
			Rows:      params.ImageParams.Rows,
			Cols:      params.ImageParams.Cols,
			Score:     int(face.Q),
			Area:      faceCoord,
			Eyes:      eyesCoords,
			Landmarks: landmarkCoords,
		}

		// Does the face significantly overlap with previous results?
		if results.Contains(f) {
			continue
		}

		results.Append(f)
	}

	return results, nil
}

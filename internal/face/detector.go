package face

import (
	_ "embed"
	"fmt"
	_ "image/jpeg"
	"io"
	"os"
	"path/filepath"
	"runtime/debug"
	"sort"

	pigo "github.com/esimov/pigo/core"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/txt"
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

// Detector struct contains Pigo face detector general settings.
type Detector struct {
	minSize        int
	angle          float64
	shiftFactor    float64
	scaleFactor    float64
	iouThreshold   float64
	scoreThreshold float32
	perturb        int
}

// Detect runs the detection algorithm over the provided source image.
func Detect(fileName string, findLandmarks bool, minSize int) (faces Faces, err error) {
	defer func() {
		if r := recover(); r != nil {
			log.Errorf("faces: %s (panic)\nstack: %s", r, debug.Stack())
		}
	}()

	if minSize < 20 {
		minSize = 20
	}

	d := &Detector{
		minSize:        minSize,
		angle:          0.0,
		shiftFactor:    0.1,
		scaleFactor:    1.1,
		iouThreshold:   0.2,
		scoreThreshold: ScoreThreshold,
		perturb:        63,
	}

	if !fs.FileExists(fileName) {
		return faces, fmt.Errorf("faces: file '%s' not found", txt.Quote(filepath.Base(fileName)))
	}

	det, params, err := d.Detect(fileName)

	if err != nil {
		return faces, fmt.Errorf("faces: %v (detect faces)", err)
	}

	if det == nil {
		return faces, fmt.Errorf("faces: no result")
	}

	faces, err = d.Faces(det, params, findLandmarks)

	if err != nil {
		return faces, fmt.Errorf("faces: %s", err)
	}

	return faces, nil
}

// Detect runs the detection algorithm over the provided source image.
func (d *Detector) Detect(fileName string) (faces []pigo.Detection, params pigo.CascadeParams, err error) {
	var srcFile io.Reader

	file, err := os.Open(fileName)

	if err != nil {
		return faces, params, err
	}

	defer file.Close()

	srcFile = file

	src, err := pigo.DecodeImage(srcFile)

	if err != nil {
		return faces, params, err
	}

	pixels := pigo.RgbToGrayscale(src)
	cols, rows := src.Bounds().Max.X, src.Bounds().Max.Y

	var maxSize int

	if cols < 20 || rows < 20 || cols < d.minSize || rows < d.minSize {
		return faces, params, fmt.Errorf("image size %dx%d is too small", cols, rows)
	} else if cols < rows {
		maxSize = cols - 8
	} else {
		maxSize = rows - 8
	}

	imageParams := &pigo.ImageParams{
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
		ImageParams: *imageParams,
	}

	log.Debugf("faces: image size %dx%d, face size min %d, max %d", cols, rows, params.MinSize, params.MaxSize)

	// Run the classifier over the obtained leaf nodes and return the Face results.
	// The result contains quadruplets representing the row, column, scale and Face score.
	faces = classifier.RunCascade(params, d.angle)

	// Calculate the intersection over union (IoU) of two clusters.
	faces = classifier.ClusterDetections(faces, d.iouThreshold)

	return faces, params, nil
}

// Faces adds landmark coordinates to detected faces and returns the results.
func (d *Detector) Faces(det []pigo.Detection, params pigo.CascadeParams, findLandmarks bool) (results Faces, err error) {
	// Sort by size.
	sort.Slice(det, func(i, j int) bool {
		return det[i].Scale > det[j].Scale
	})

	for _, face := range det {
		// Small faces require higher quality.
		threshold := d.scoreThreshold

		if face.Scale < 30 {
			threshold += 11.5
		} else if face.Scale < 50 {
			threshold += 9.0
		} else if face.Scale < 80 {
			threshold += 6.5
		} else if face.Scale < 110 {
			threshold += 2.5
		}

		// Skip face if quality is too low.
		if face.Q < threshold {
			continue
		}

		var eyesCoords []Area
		var landmarkCoords []Area
		var puploc *pigo.Puploc

		faceCoord := NewArea(
			"face",
			face.Row,
			face.Col,
			face.Scale,
		)

		if face.Scale > 50 && findLandmarks {
			// Find left eye.
			puploc = &pigo.Puploc{
				Row:      face.Row - int(0.075*float32(face.Scale)),
				Col:      face.Col - int(0.175*float32(face.Scale)),
				Scale:    float32(face.Scale) * 0.25,
				Perturbs: d.perturb,
			}

			leftEye := plc.RunDetector(*puploc, params.ImageParams, d.angle, false)

			if leftEye.Row > 0 && leftEye.Col > 0 {
				eyesCoords = append(eyesCoords, NewArea(
					"eye_l",
					leftEye.Row,
					leftEye.Col,
					int(leftEye.Scale),
				))
			}

			// Find right eye.
			puploc = &pigo.Puploc{
				Row:      face.Row - int(0.075*float32(face.Scale)),
				Col:      face.Col + int(0.185*float32(face.Scale)),
				Scale:    float32(face.Scale) * 0.25,
				Perturbs: d.perturb,
			}

			rightEye := plc.RunDetector(*puploc, params.ImageParams, d.angle, false)

			if rightEye.Row > 0 && rightEye.Col > 0 {
				eyesCoords = append(eyesCoords, NewArea(
					"eye_r",
					rightEye.Row,
					rightEye.Col,
					int(rightEye.Scale),
				))
			}

			if leftEye != nil && rightEye != nil {
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
			}

			// Find mouth.
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

			flpc := flpcs["lp84"][0]

			if flpc != nil {
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

		f := Face{
			Rows:      params.ImageParams.Rows,
			Cols:      params.ImageParams.Cols,
			Score:     int(face.Q),
			Area:      faceCoord,
			Eyes:      eyesCoords,
			Landmarks: landmarkCoords,
		}

		if results.Contains(f) {
			// Ignore.
		} else {
			results.Append(f)
		}
	}

	return results, nil
}

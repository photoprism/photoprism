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
	maxSize        int
	angle          float64
	shiftFactor    float64
	scaleFactor    float64
	iouThreshold   float64
	scoreThreshold float32
	perturb        int
}

// Detect runs the detection algorithm over the provided source image.
func Detect(fileName string) (faces Faces, err error) {
	defer func() {
		if r := recover(); r != nil {
			log.Errorf("faces: %s (panic)\nstack: %s", r, debug.Stack())
		}
	}()

	fd := &Detector{
		minSize:        20,
		maxSize:        1000,
		angle:          0.0,
		shiftFactor:    0.1,
		scaleFactor:    1.1,
		iouThreshold:   0.2,
		scoreThreshold: 9.0,
		perturb:        63,
	}

	if !fs.FileExists(fileName) {
		return faces, fmt.Errorf("faces: file '%s' not found", txt.Quote(filepath.Base(fileName)))
	}

	det, params, err := fd.Detect(fileName)

	if err != nil {
		return faces, fmt.Errorf("faces: %v (detect faces)", err)
	}

	if det == nil {
		return faces, fmt.Errorf("faces: no result")
	}

	faces, err = fd.Faces(det, params)

	if err != nil {
		return faces, fmt.Errorf("faces: %s", err)
	}

	return faces, nil
}

// Detect runs the detection algorithm over the provided source image.
func (fd *Detector) Detect(fileName string) (faces []pigo.Detection, params pigo.CascadeParams, err error) {
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

	imageParams := &pigo.ImageParams{
		Pixels: pixels,
		Rows:   rows,
		Cols:   cols,
		Dim:    cols,
	}

	if rows > 800 || cols > 800 {
		fd.scoreThreshold += 9.0
	}

	params = pigo.CascadeParams{
		MinSize:     fd.minSize,
		MaxSize:     fd.maxSize,
		ShiftFactor: fd.shiftFactor,
		ScaleFactor: fd.scaleFactor,
		ImageParams: *imageParams,
	}

	// Run the classifier over the obtained leaf nodes and return the Face results.
	// The result contains quadruplets representing the row, column, scale and Face score.
	faces = classifier.RunCascade(params, fd.angle)

	// Calculate the intersection over union (IoU) of two clusters.
	faces = classifier.ClusterDetections(faces, fd.iouThreshold)

	return faces, params, nil
}

// Faces adds landmark coordinates to detected faces and returns the results.
func (fd *Detector) Faces(det []pigo.Detection, params pigo.CascadeParams) (results Faces, err error) {
	var maxQ float32

	// Sort by quality.
	sort.Slice(det, func(i, j int) bool {
		return det[i].Q > det[j].Q
	})

	for _, face := range det {
		var eyesCoords []Point
		var landmarkCoords []Point
		var puploc *pigo.Puploc

		if face.Q < fd.scoreThreshold {
			continue
		}

		if maxQ < face.Q {
			maxQ = face.Q
		} else if maxQ >= 20 && face.Q < 15 {
			continue
		}

		faceCoord := NewPoint(
			"face",
			face.Row,
			face.Col,
			face.Scale,
		)

		if face.Scale > 50 {
			// Find left eye.
			puploc = &pigo.Puploc{
				Row:      face.Row - int(0.075*float32(face.Scale)),
				Col:      face.Col - int(0.175*float32(face.Scale)),
				Scale:    float32(face.Scale) * 0.25,
				Perturbs: fd.perturb,
			}

			leftEye := plc.RunDetector(*puploc, params.ImageParams, fd.angle, false)

			if leftEye.Row > 0 && leftEye.Col > 0 {
				eyesCoords = append(eyesCoords, NewPoint(
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
				Perturbs: fd.perturb,
			}

			rightEye := plc.RunDetector(*puploc, params.ImageParams, fd.angle, false)

			if rightEye.Row > 0 && rightEye.Col > 0 {
				eyesCoords = append(eyesCoords, NewPoint(
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

						flp := flpc.GetLandmarkPoint(leftEye, rightEye, params.ImageParams, fd.perturb, false)
						if flp.Row > 0 && flp.Col > 0 {
							landmarkCoords = append(landmarkCoords, NewPoint(
								eye,
								flp.Row,
								flp.Col,
								int(flp.Scale),
							))
						}

						flp = flpc.GetLandmarkPoint(leftEye, rightEye, params.ImageParams, fd.perturb, true)
						if flp.Row > 0 && flp.Col > 0 {
							landmarkCoords = append(landmarkCoords, NewPoint(
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

					flp := flpc.GetLandmarkPoint(leftEye, rightEye, params.ImageParams, fd.perturb, false)
					if flp.Row > 0 && flp.Col > 0 {
						landmarkCoords = append(landmarkCoords, NewPoint(
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
				flp := flpc.GetLandmarkPoint(leftEye, rightEye, params.ImageParams, fd.perturb, true)
				if flp.Row > 0 && flp.Col > 0 {
					landmarkCoords = append(landmarkCoords, NewPoint(
						"lp84",
						flp.Row,
						flp.Col,
						int(flp.Scale),
					))
				}
			}
		}

		results = append(results, Face{
			Rows:      params.ImageParams.Rows,
			Cols:      params.ImageParams.Cols,
			Score:     int(face.Q),
			Face:      faceCoord,
			Eyes:      eyesCoords,
			Landmarks: landmarkCoords,
		})
	}

	return results, nil
}

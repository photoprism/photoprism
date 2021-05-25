package face

import (
	"embed"
	"fmt"
	_ "image/jpeg"
	"io"
	"os"
	"time"

	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/txt"

	pigo "github.com/esimov/pigo/core"
)

//go:embed cascade/lps/*
var efs embed.FS

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
		log.Errorf("face: %s", err)
	}

	pl := pigo.NewPuplocCascade()
	plc, err = pl.UnpackCascade(puplocFile)

	if err != nil {
		log.Errorf("face: %s", err)
	}

	flpcs, err = ReadCascadeDir(pl, "cascade/lps")

	if err != nil {
		log.Errorf("face: %s", err)
	}
}

var (
	eyeCascades   = []string{"lp46", "lp44", "lp42", "lp38", "lp312"}
	mouthCascades = []string{"lp93", "lp84", "lp82", "lp81"}
)

// Detector struct contains Pigo face detector general settings.
type Detector struct {
	minSize      int
	maxSize      int
	angle        float64
	shiftFactor  float64
	scaleFactor  float64
	iouThreshold float64
}

func DefaultDetector() *Detector {
	return &Detector{
		minSize:      20,
		maxSize:      1000,
		angle:        0.0,
		shiftFactor:  0.1,
		scaleFactor:  1.1,
		iouThreshold: 0.2,
	}
}

// Detect runs the detection algorithm over the provided source image.
func Detect(fileName string, fd *Detector) (det Faces, err error) {
	if !fs.FileExists(fileName) {
		return det, fmt.Errorf("face: file '%s' not found", fileName)
	}

	start := time.Now()

	log.Debugf("\nface: detecting faces in %s", txt.Quote(fileName))

	faces, params, err := fd.Detect(fileName)
	if err != nil {
		return det, fmt.Errorf("face: %v (detect faces)", err)
	}

	det, err = fd.Results(faces, params)

	if err != nil {
		return det, fmt.Errorf("face: %s (Faces)", err)
	}

	log.Debugf("\nface: %s done in \x1b[92m%.2fs\n", txt.Quote(fileName), time.Since(start).Seconds())

	return det, nil
}

// Detect runs the detection algorithm over the provided source image.
func (fd *Detector) Detect(fileName string) (faces []pigo.Detection, params pigo.CascadeParams, err error) {
	var srcFile io.Reader

	file, err := os.Open(fileName)

	if err != nil {
		return faces, params, err
	}

	defer func(file *os.File) {
		_ = file.Close()
	}(file)

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
func (fd *Detector) Results(faces []pigo.Detection, params pigo.CascadeParams) (Faces, error) {
	var (
		qThresh float32 = 5.0
		perturb         = 63
	)

	var (
		detections     Faces
		eyesCoords     []Point
		landmarkCoords []Point
		puploc         *pigo.Puploc
	)

	for _, face := range faces {
		if face.Q > qThresh {
			faceCoord := NewPoint(
				"face",
				face.Row-face.Scale/2,
				face.Col-face.Scale/2,
				face.Scale,
			)

			if face.Scale > 50 {
				// Find left eye.
				puploc = &pigo.Puploc{
					Row:      face.Row - int(0.075*float32(face.Scale)),
					Col:      face.Col - int(0.175*float32(face.Scale)),
					Scale:    float32(face.Scale) * 0.25,
					Perturbs: perturb,
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
					Perturbs: perturb,
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

				for _, eye := range eyeCascades {
					for _, flpc := range flpcs[eye] {
						flp := flpc.GetLandmarkPoint(leftEye, rightEye, params.ImageParams, perturb, false)
						if flp.Row > 0 && flp.Col > 0 {
							landmarkCoords = append(landmarkCoords, NewPoint(
								eye,
								flp.Row,
								flp.Col,
								int(flp.Scale),
							))
						}

						flp = flpc.GetLandmarkPoint(leftEye, rightEye, params.ImageParams, perturb, true)
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

				// Find mouth.
				for _, mouth := range mouthCascades {
					for _, flpc := range flpcs[mouth] {
						flp := flpc.GetLandmarkPoint(leftEye, rightEye, params.ImageParams, perturb, false)
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
				flp := flpcs["lp84"][0].GetLandmarkPoint(leftEye, rightEye, params.ImageParams, perturb, true)
				if flp.Row > 0 && flp.Col > 0 {
					landmarkCoords = append(landmarkCoords, NewPoint(
						"lp84",
						flp.Row,
						flp.Col,
						int(flp.Scale),
					))
				}
			}

			detections = append(detections, Face{
				Rows:      params.ImageParams.Rows,
				Cols:      params.ImageParams.Cols,
				Face:      faceCoord,
				Eyes:      eyesCoords,
				Landmarks: landmarkCoords,
			})
		}
	}

	return detections, nil
}

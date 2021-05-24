/*

Package face provides face landmark detection.

Copyright (c) 2018 - 2021 Michael Mayer <hello@photoprism.org>

    This program is free software: you can redistribute it and/or modify
    it under the terms of the GNU Affero General Public License as published
    by the Free Software Foundation, either version 3 of the License, or
    (at your option) any later version.

    This program is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU Affero General Public License for more details.

    You should have received a copy of the GNU Affero General Public License
    along with this program.  If not, see <https://www.gnu.org/licenses/>.

    PhotoPrism® is a registered trademark of Michael Mayer.  You may use it as required
    to describe our software, run your own server, for educational purposes, but not for
    offering commercial goods, products, or services without prior written permission.
    In other words, please ask.

Feel free to send an e-mail to hello@photoprism.org if you have questions,
want to support our work, or just want to say hello.

Additional information can be found in our Developer Guide:
https://docs.photoprism.org/developer-guide/

*/

package face

import (
	"embed"
	"fmt"
	_ "image/jpeg"
	"io"
	"os"
	"time"

	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/txt"

	pigo "github.com/esimov/pigo/core"
)

//go:embed cascade/lps/*
var efs embed.FS

var log = event.Log

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

// Point represents face landmark coordinates.
type Point struct {
	Id    string `json:"id,omitempty"`
	Row   int    `json:"x,omitempty"`
	Col   int    `json:"y,omitempty"`
	Scale int    `json:"size,omitempty"`
}

// Detect runs the detection algorithm over the provided source image.
func Detect(fileName string, fd *Detector) (det Results, err error) {
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
		return det, fmt.Errorf("face: %s (Results)", err)
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

	// Run the classifier over the obtained leaf nodes and return the Result results.
	// The result contains quadruplets representing the row, column, scale and Result score.
	faces = classifier.RunCascade(params, fd.angle)

	// Calculate the intersection over union (IoU) of two clusters.
	faces = classifier.ClusterDetections(faces, fd.iouThreshold)

	return faces, params, nil
}

// Results adds landmark coordinates to detected faces and returns the results.
func (fd *Detector) Results(faces []pigo.Detection, params pigo.CascadeParams) (Results, error) {
	var (
		qThresh float32 = 5.0
		perturb         = 63
	)

	var (
		detections     Results
		eyesCoords     []Point
		landmarkCoords []Point
		puploc         *pigo.Puploc
	)

	for _, face := range faces {
		if face.Q > qThresh {
			faceCoord := &Point{
				Col:   face.Row - face.Scale/2,
				Row:   face.Col - face.Scale/2,
				Scale: face.Scale,
			}

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
					eyesCoords = append(eyesCoords, Point{
						Id:    "eye_l",
						Col:   leftEye.Row,
						Row:   leftEye.Col,
						Scale: int(leftEye.Scale),
					})
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
					eyesCoords = append(eyesCoords, Point{
						Id:    "eye_r",
						Col:   rightEye.Row,
						Row:   rightEye.Col,
						Scale: int(rightEye.Scale),
					})
				}

				for _, eye := range eyeCascades {
					for _, flpc := range flpcs[eye] {
						flp := flpc.GetLandmarkPoint(leftEye, rightEye, params.ImageParams, perturb, false)
						if flp.Row > 0 && flp.Col > 0 {
							landmarkCoords = append(landmarkCoords, Point{
								Id:    eye,
								Col:   flp.Row,
								Row:   flp.Col,
								Scale: int(flp.Scale),
							})
						}

						flp = flpc.GetLandmarkPoint(leftEye, rightEye, params.ImageParams, perturb, true)
						if flp.Row > 0 && flp.Col > 0 {
							landmarkCoords = append(landmarkCoords, Point{
								Id:    eye,
								Col:   flp.Row,
								Row:   flp.Col,
								Scale: int(flp.Scale),
							})
						}
					}
				}

				// Find mouth.
				for _, mouth := range mouthCascades {
					for _, flpc := range flpcs[mouth] {
						flp := flpc.GetLandmarkPoint(leftEye, rightEye, params.ImageParams, perturb, false)
						if flp.Row > 0 && flp.Col > 0 {
							landmarkCoords = append(landmarkCoords, Point{
								Id:    mouth,
								Col:   flp.Row,
								Row:   flp.Col,
								Scale: int(flp.Scale),
							})
						}
					}
				}
				flp := flpcs["lp84"][0].GetLandmarkPoint(leftEye, rightEye, params.ImageParams, perturb, true)
				if flp.Row > 0 && flp.Col > 0 {
					landmarkCoords = append(landmarkCoords, Point{
						Id:    "lp84",
						Col:   flp.Row,
						Row:   flp.Col,
						Scale: int(flp.Scale),
					})
				}
			}

			detections = append(detections, Result{
				Rows:      params.ImageParams.Rows,
				Cols:      params.ImageParams.Cols,
				Face:      *faceCoord,
				Eyes:      eyesCoords,
				Landmarks: landmarkCoords,
			})
		}
	}

	return detections, nil
}

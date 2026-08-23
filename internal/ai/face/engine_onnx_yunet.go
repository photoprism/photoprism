package face

import (
	"fmt"
	"math"

	onnxruntime "github.com/yalue/onnxruntime_go"
)

// yunetStrides lists the feature map strides YuNet predicts at.
var yunetStrides = [3]int{8, 16, 32}

// yunetOutputs lists the output tensors per stride, in the order the decoder reads them.
var yunetOutputs = [4][3]string{
	{"cls_8", "cls_16", "cls_32"},
	{"obj_8", "obj_16", "obj_32"},
	{"bbox_8", "bbox_16", "bbox_32"},
	{"kps_8", "kps_16", "kps_32"},
}

// yunetFeatDim returns the feature map extent a stride produces for an input extent.
//
// The convolution chain halves with ceil semantics, so this is not integer division: a 720 px
// axis yields 23 cells at stride 32 rather than 22. It agrees with division whenever the input
// is a multiple of 32, which is the only geometry the graph accepts anyway, so it exists to keep
// the decoder correct if that ever stops being true.
func yunetFeatDim(size, stride int) int {
	for s := 1; s < stride; s *= 2 {
		size = (size + 1) / 2
	}

	return size
}

// parseYuNetDetections decodes YuNet output into bounding boxes and landmarks in the coordinate
// space of the source image.
//
// YuNet is anchor-free with one prior per cell, so a cell's own row and column are the reference
// point rather than a table of anchors, and its score is split across a classification and an
// objectness head that are combined geometrically. Its five keypoints already arrive in the order
// ArcFaceTemplate expects, which was established by fitting both orderings: as emitted they fit at
// a mean residual of 5.0, swapped at 22.1.
func (o *onnxEngine) parseYuNetDetections(values []onnxruntime.Value, detScale float32, origWidth, origHeight int) ([]onnxDetection, error) {
	index := make(map[string]int, len(o.outputNames))
	for i, name := range o.outputNames {
		index[name] = i
	}

	tensor := func(name string) ([]float32, error) {
		i, ok := index[name]
		if !ok || i >= len(values) {
			return nil, fmt.Errorf("faces: yunet output %s is missing", name)
		}

		t, ok := values[i].(*onnxruntime.Tensor[float32])
		if !ok {
			return nil, fmt.Errorf("faces: unexpected tensor type for yunet output %s", name)
		}

		return t.GetData(), nil
	}

	detections := make([]onnxDetection, 0, 32)

	for level, stride := range yunetStrides {
		cls, err := tensor(yunetOutputs[0][level])
		if err != nil {
			return nil, err
		}

		obj, err := tensor(yunetOutputs[1][level])
		if err != nil {
			return nil, err
		}

		bbox, err := tensor(yunetOutputs[2][level])
		if err != nil {
			return nil, err
		}

		kps, err := tensor(yunetOutputs[3][level])
		if err != nil {
			return nil, err
		}

		featWidth := yunetFeatDim(o.inputWidth, stride)
		featHeight := yunetFeatDim(o.inputHeight, stride)
		cells := featWidth * featHeight

		if len(cls) != cells || len(obj) != cells || len(bbox) != cells*4 || len(kps) != cells*NumLandmarks*2 {
			return nil, fmt.Errorf("faces: yunet stride %d expects %d cells for %dx%d, got cls=%d bbox=%d kps=%d",
				stride, cells, o.inputWidth, o.inputHeight, len(cls), len(bbox), len(kps))
		}

		s := float32(stride)

		for idx := range cells {
			// Both heads are probabilities, and the geometric mean is what OpenCV's reference
			// implementation scores with. Clamping first keeps a marginally out-of-range value
			// from producing a NaN under the square root.
			c := float64(clampFloat32(cls[idx], 0, 1))
			p := float64(clampFloat32(obj[idx], 0, 1))
			score := float32(math.Sqrt(c * p))

			if score < o.scoreThreshold {
				continue
			}

			col := float32(idx % featWidth)
			row := float32(idx / featWidth)

			cx := (col + bbox[idx*4]) * s
			cy := (row + bbox[idx*4+1]) * s
			w := float32(math.Exp(float64(bbox[idx*4+2]))) * s
			h := float32(math.Exp(float64(bbox[idx*4+3]))) * s

			det := onnxDetection{
				x1:    clampFloat32((cx-w/2)/detScale, 0, float32(origWidth)),
				y1:    clampFloat32((cy-h/2)/detScale, 0, float32(origHeight)),
				x2:    clampFloat32((cx+w/2)/detScale, 0, float32(origWidth)),
				y2:    clampFloat32((cy+h/2)/detScale, 0, float32(origHeight)),
				score: score,
			}

			if det.x2 <= det.x1 || det.y2 <= det.y1 {
				continue
			}

			// Keypoints stay unclamped for the reason the SCRFD path gives: the alignment fit
			// is least squares over all five, so snapping one that fell outside the frame
			// rotates and scales the whole crop.
			for p := range NumLandmarks {
				det.kps[p*2] = (kps[idx*NumLandmarks*2+p*2] + col) * s / detScale
				det.kps[p*2+1] = (kps[idx*NumLandmarks*2+p*2+1] + row) * s / detScale
			}

			det.hasKps = true

			detections = append(detections, det)
		}
	}

	return detections, nil
}

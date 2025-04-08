package tensorflow

import (
	"bytes"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"runtime/debug"

	tf "github.com/wamuir/graft/tensorflow"
	"github.com/wamuir/graft/tensorflow/op"

	"github.com/photoprism/photoprism/pkg/fs"
)

const (
	Mean  = float32(117)
	Scale = float32(1)
)

func ImageFromFile(fileName string, resolution int) (*tf.Tensor, error) {
	if img, err := OpenImage(fileName); err != nil {
		return nil, err
	} else {
		return Image(img, resolution)
	}
}

func OpenImage(fileName string) (image.Image, error) {
	f, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	img, _, err := image.Decode(f)

	return img, err
}

func ImageFromBytes(b []byte, resolution int) (*tf.Tensor, error) {
	img, _, imgErr := image.Decode(bytes.NewReader(b))

	if imgErr != nil {
		return nil, imgErr
	}

	return Image(img, resolution)
}

func Image(img image.Image, resolution int) (tfTensor *tf.Tensor, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("tensorflow: %s (panic)\nstack: %s", r, debug.Stack())
		}
	}()

	if resolution <= 0 {
		return tfTensor, fmt.Errorf("tensorflow: resolution must be larger 0")
	}

	var tfImage [1][][][3]float32

	for j := 0; j < resolution; j++ {
		tfImage[0] = append(tfImage[0], make([][3]float32, resolution))
	}

	for i := 0; i < resolution; i++ {
		for j := 0; j < resolution; j++ {
			r, g, b, _ := img.At(i, j).RGBA()
			tfImage[0][j][i][0] = convertValue(r, 127.5)
			tfImage[0][j][i][1] = convertValue(g, 127.5)
			tfImage[0][j][i][2] = convertValue(b, 127.5)
		}
	}

	return tf.NewTensor(tfImage)
}

// ImageTransform transforms the given image into a *tf.Tensor and returns it.
func ImageTransform(image []byte, imageFormat fs.Type, resolution int) (*tf.Tensor, error) {
	tensor, err := tf.NewTensor(string(image))
	if err != nil {
		return nil, err
	}

	graph, input, output, err := transformImageGraph(imageFormat, resolution)

	if err != nil {
		return nil, err
	}

	session, err := tf.NewSession(graph, nil)
	if err != nil {
		return nil, err
	}
	defer session.Close()

	normalized, err := session.Run(
		map[tf.Output]*tf.Tensor{input: tensor},
		[]tf.Output{output},
		nil)
	if err != nil {
		return nil, err
	}

	return normalized[0], nil
}

func transformImageGraph(imageFormat fs.Type, resolution int) (graph *tf.Graph, input, output tf.Output, err error) {
	s := op.NewScope()
	input = op.Placeholder(s, tf.String)

	// Assume the image is a JPEG, or a PNG if explicitly specified.
	var decodedImage tf.Output
	switch imageFormat {
	case fs.ImagePng:
		decodedImage = op.DecodePng(s, input, op.DecodePngChannels(3))
	default:
		decodedImage = op.DecodeJpeg(s, input, op.DecodeJpegChannels(3))
	}

	output = op.Div(s,
		op.Sub(s,
			op.ResizeBilinear(s,
				op.ExpandDims(s,
					op.Cast(s, decodedImage, tf.Float),
					op.Const(s.SubScope("make_batch"), int32(0))),
				op.Const(s.SubScope("size"), []int32{int32(resolution), int32(resolution)})),
			op.Const(s.SubScope("mean"), Mean)),
		op.Const(s.SubScope("scale"), Scale))

	graph, err = s.Finalize()

	return graph, input, output, err
}

func convertValue(value uint32, mean float32) float32 {
	if mean == 0 {
		mean = 127.5
	}

	return (float32(value>>8) - mean) / mean
}

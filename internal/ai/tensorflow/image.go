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

func ImageFromFile(fileName string, input *PhotoInput) (*tf.Tensor, error) {
	if img, err := OpenImage(fileName); err != nil {
		return nil, err
	} else {
		return Image(img, input, nil)
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

func ImageFromBytes(b []byte, input *PhotoInput, builder *ImageTensorBuilder) (*tf.Tensor, error) {
	img, _, imgErr := image.Decode(bytes.NewReader(b))

	if imgErr != nil {
		return nil, imgErr
	}

	return Image(img, input, builder)
}

func Image(img image.Image, input *PhotoInput, builder *ImageTensorBuilder) (tfTensor *tf.Tensor, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("tensorflow: %s (panic)\nstack: %s", r, debug.Stack())
		}
	}()

	if input.Resolution() <= 0 {
		return tfTensor, fmt.Errorf("tensorflow: resolution must be larger than 0")
	}

	if builder == nil {
		builder, err = NewImageTensorBuilder(input)
		if err != nil {
			return nil, err
		}
	}

	for i := 0; i < input.Resolution(); i++ {
		for j := 0; j < input.Resolution(); j++ {
			r, g, b, _ := img.At(i, j).RGBA()
			//Although RGB can be disordered, we assume the input intervals are
			//given in RGB order.
			builder.Set(i, j,
				convertValue(r, input.GetInterval(0)),
				convertValue(g, input.GetInterval(1)),
				convertValue(b, input.GetInterval(2)))
		}
	}

	return builder.BuildTensor()
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

func convertValue(value uint32, interval *Interval) float32 {
	var scale float32

	if interval.Mean != nil {
		scale = *interval.Mean
	} else {
		scale = interval.Size() / 255.0
	}
	offset := interval.Offset()

	return (float32(value>>8))*scale + offset
}

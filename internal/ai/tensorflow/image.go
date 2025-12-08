package tensorflow

import (
	"bytes"
	"fmt"
	"image"
	_ "image/jpeg" // register JPEG decoder
	_ "image/png"  // register PNG decoder
	"math"
	"os"
	"runtime/debug"

	tf "github.com/wamuir/graft/tensorflow"
	"github.com/wamuir/graft/tensorflow/op"

	"github.com/photoprism/photoprism/pkg/fs"
)

const (
	// Mean is the default mean pixel value used during normalization.
	Mean = float32(117)
	// Scale is the default scale applied during normalization.
	Scale = float32(1)
)

// ImageFromFile decodes an image from disk and converts it to a tensor for inference.
func ImageFromFile(fileName string, input *PhotoInput) (*tf.Tensor, error) {
	if img, err := OpenImage(fileName); err != nil {
		return nil, err
	} else {
		return Image(img, input, nil)
	}
}

// OpenImage opens an image file and decodes it using the registered decoders.
func OpenImage(fileName string) (image.Image, error) {
	f, err := os.Open(fileName) //nolint:gosec // fileName supplied by trusted caller; reading local images is expected
	if err != nil {
		return nil, err
	}
	defer f.Close()
	img, _, err := image.Decode(f)

	return img, err
}

// ImageFromBytes converts raw image bytes into a tensor using the provided input definition.
func ImageFromBytes(b []byte, input *PhotoInput, builder *ImageTensorBuilder) (*tf.Tensor, error) {
	img, _, imgErr := image.Decode(bytes.NewReader(b))

	if imgErr != nil {
		return nil, imgErr
	}

	return Image(img, input, builder)
}

// Image converts a decoded image into a tensor matching the provided input description.
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
			// Although RGB can be disordered, we assume the input intervals are
			// given in RGB order.
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

	if resolution <= 0 || resolution > math.MaxInt32 {
		return nil, input, output, fmt.Errorf("tensorflow: resolution %d is out of bounds", resolution)
	}

	// Assume the image is a JPEG, or a PNG if explicitly specified.
	var decodedImage tf.Output
	switch imageFormat {
	case fs.ImagePng:
		decodedImage = op.DecodePng(s, input, op.DecodePngChannels(3))
	default:
		decodedImage = op.DecodeJpeg(s, input, op.DecodeJpegChannels(3))
	}

	size := int32(resolution) //nolint:gosec // resolution is validated to be within int32 range above

	output = op.Div(s,
		op.Sub(s,
			op.ResizeBilinear(s,
				op.ExpandDims(s,
					op.Cast(s, decodedImage, tf.Float),
					op.Const(s.SubScope("make_batch"), int32(0))),
				op.Const(s.SubScope("size"), []int32{size, size})),
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

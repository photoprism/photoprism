package tensorflow

import (
	"errors"
	"fmt"

	tf "github.com/wamuir/graft/tensorflow"
)

// ImageTensorBuilder incrementally constructs an image tensor in BHWC or BCHW order.
type ImageTensorBuilder struct {
	data       []float32
	shape      []ShapeComponent
	resolution int
	rIndex     int
	gIndex     int
	bIndex     int
}

func shapeLen(c ShapeComponent, res int) int {
	switch c {
	case ShapeBatch:
		return 1
	case ShapeHeight, ShapeWidth:
		return res
	case ShapeColor:
		return 3
	default:
		return -1
	}
}

// NewImageTensorBuilder creates a builder for the given photo input definition.
func NewImageTensorBuilder(input *PhotoInput) (*ImageTensorBuilder, error) {

	if len(input.Shape) != 4 {
		return nil, fmt.Errorf("tensorflow: the shape length is %d and should be 4", len(input.Shape))
	}

	if input.Shape[0] != ShapeBatch {
		return nil, errors.New("tensorflow: the first shape component must be Batch")
	}

	if input.Shape[1] != ShapeColor && input.Shape[3] != ShapeColor {
		return nil, fmt.Errorf("tensorflow: unsupported shape %v", input.Shape)
	}

	totalSize := 1
	for i := range input.Shape {
		totalSize *= shapeLen(input.Shape[i], input.Resolution())
	}

	// Allocate just one big chunk
	flatBuffer := make([]float32, totalSize)

	rIndex, gIndex, bIndex := input.ColorChannelOrder.Indices()
	return &ImageTensorBuilder{
		data:       flatBuffer,
		shape:      input.Shape,
		resolution: input.Resolution(),
		rIndex:     rIndex,
		gIndex:     gIndex,
		bIndex:     bIndex,
	}, nil
}

// Set assigns the normalized RGB values for the pixel at (x,y).
func (t *ImageTensorBuilder) Set(x, y int, r, g, b float32) {
	t.data[t.flatIndex(x, y, t.rIndex)] = r
	t.data[t.flatIndex(x, y, t.gIndex)] = g
	t.data[t.flatIndex(x, y, t.bIndex)] = b
}

func (t *ImageTensorBuilder) flatIndex(x, y, c int) int {

	shapeVal := func(s ShapeComponent) int {
		switch s {
		case ShapeBatch:
			return 0
		case ShapeColor:
			return c
		case ShapeWidth:
			return x
		case ShapeHeight:
			return y
		default:
			return 0
		}
	}

	idx := 0
	for _, s := range t.shape {
		idx = idx*shapeLen(s, t.resolution) + shapeVal(s)
	}

	return idx
}

// BuildTensor materializes the underlying data into a TensorFlow tensor.
func (t *ImageTensorBuilder) BuildTensor() (*tf.Tensor, error) {

	arr := make([][][][]float32, shapeLen(t.shape[0], t.resolution))
	offset := 0
	for i := 0; i < shapeLen(t.shape[0], t.resolution); i++ {
		arr[i] = make([][][]float32, shapeLen(t.shape[1], t.resolution))
		for j := 0; j < shapeLen(t.shape[1], t.resolution); j++ {
			arr[i][j] = make([][]float32, shapeLen(t.shape[2], t.resolution))
			for k := 0; k < shapeLen(t.shape[2], t.resolution); k++ {
				arr[i][j][k] = t.data[offset : offset+shapeLen(t.shape[3], t.resolution)]
				offset += shapeLen(t.shape[3], t.resolution)
			}
		}
	}

	return tf.NewTensor(arr)
}

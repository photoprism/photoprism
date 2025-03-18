package vector

import (
	"fmt"
)

// Vector represents a set of floating-point values.
type Vector []float64

// Vectors represents a set of vectors.
type Vectors = []Vector

// NewVector creates a new vector from the given values.
func NewVector(values interface{}) (Vector, error) {
	switch v := values.(type) {
	case []uint8:
		return uint8ToVector(v), nil
	case []uint16:
		return uint16ToVector(v), nil
	case []uint32:
		return uint32ToVector(v), nil
	case []uint64:
		return uint64ToVector(v), nil
	case []int:
		return intToVector(v), nil
	case []int8:
		return int8ToVector(v), nil
	case []int16:
		return int16ToVector(v), nil
	case []int32:
		return int32ToVector(v), nil
	case []int64:
		return int64ToVector(v), nil
	case []float32:
		return float32ToVector(v), nil
	case []float64:
		return float64ToVector(v), nil
	case Vector:
		return v, nil
	default:
		return nil, fmt.Errorf("cannot create vector from type %T", values)
	}
}

// NullVector creates a new null vector with the given dimension.
func NullVector(dim int) Vector {
	return make(Vector, dim)
}

// uint8ToVector creates a new vector from a non-empty uint8 slice.
func uint8ToVector(values []uint8) Vector {
	v := make(Vector, len(values))

	for i := range values {
		v[i] = float64(values[i])
	}

	return v
}

// uint16ToVector creates a new vector from a non-empty uint16 slice.
func uint16ToVector(values []uint16) Vector {
	v := make(Vector, len(values))

	for i := range values {
		v[i] = float64(values[i])
	}

	return v
}

// uint32ToVector creates a new vector from a non-empty uint32 slice.
func uint32ToVector(values []uint32) Vector {
	v := make(Vector, len(values))

	for i := range values {
		v[i] = float64(values[i])
	}

	return v
}

// uint64ToVector creates a new vector from a non-empty uint64 slice.
func uint64ToVector(values []uint64) Vector {
	v := make(Vector, len(values))

	for i := range values {
		v[i] = float64(values[i])
	}

	return v
}

// intToVector creates a new vector from a non-empty int slice.
func intToVector(values []int) Vector {
	v := make(Vector, len(values))

	for i := range values {
		v[i] = float64(values[i])
	}

	return v
}

// int8ToVector creates a new vector from a non-empty int8 slice.
func int8ToVector(values []int8) Vector {
	v := make(Vector, len(values))

	for i := range values {
		v[i] = float64(values[i])
	}

	return v
}

// int16ToVector creates a new vector from a non-empty int16 slice.
func int16ToVector(values []int16) Vector {
	v := make(Vector, len(values))

	for i := range values {
		v[i] = float64(values[i])
	}

	return v
}

// int32ToVector creates a new vector from a non-empty int32 slice.
func int32ToVector(values []int32) Vector {
	v := make(Vector, len(values))

	for i := range values {
		v[i] = float64(values[i])
	}

	return v
}

// int64ToVector creates a new vector from a non-empty int64 slice.
func int64ToVector(values []int64) Vector {
	v := make(Vector, len(values))

	for i := range values {
		v[i] = float64(values[i])
	}

	return v
}

// float32ToVector creates a new vector from a non-empty float32 slice.
func float32ToVector(values []float32) Vector {
	v := make(Vector, len(values))

	for i := range values {
		v[i] = float64(values[i])
	}

	return v
}

// float64ToVector creates a new vector from a non-empty float64 slice.
func float64ToVector(values []float64) Vector {
	return Vector(values)
}

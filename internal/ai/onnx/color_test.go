package onnx

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

func TestColorOrder_Indices(t *testing.T) {
	t.Run("RGB", func(t *testing.T) {
		r, g, b := RGB.Indices()
		assert.Equal(t, 0, r)
		assert.Equal(t, 1, g)
		assert.Equal(t, 2, b)
	})
	t.Run("BGR", func(t *testing.T) {
		r, g, b := BGR.Indices()
		assert.Equal(t, 2, r)
		assert.Equal(t, 1, g)
		assert.Equal(t, 0, b)
	})
	t.Run("Permutation", func(t *testing.T) {
		r, g, b := ColorOrder("GBR").Indices()
		assert.Equal(t, 2, r)
		assert.Equal(t, 0, g)
		assert.Equal(t, 1, b)
	})
	t.Run("Undefined", func(t *testing.T) {
		r, g, b := OrderUndefined.Indices()
		assert.Equal(t, 0, r)
		assert.Equal(t, 1, g)
		assert.Equal(t, 2, b)
	})
	t.Run("Invalid", func(t *testing.T) {
		r, g, b := ColorOrder("RRG").Indices()
		assert.Equal(t, 0, r)
		assert.Equal(t, 1, g)
		assert.Equal(t, 2, b)
	})
}

func TestColorOrder_Valid(t *testing.T) {
	t.Run("Uppercase", func(t *testing.T) {
		assert.True(t, RGB.Valid())
		assert.True(t, BGR.Valid())
	})
	t.Run("Lowercase", func(t *testing.T) {
		assert.True(t, ColorOrder("bgr").Valid())
	})
	t.Run("Duplicate", func(t *testing.T) {
		assert.False(t, ColorOrder("RRG").Valid())
	})
	t.Run("TooShort", func(t *testing.T) {
		assert.False(t, ColorOrder("RG").Valid())
	})
	t.Run("Empty", func(t *testing.T) {
		assert.False(t, OrderUndefined.Valid())
	})
}

func TestColorOrder_String(t *testing.T) {
	t.Run("Lowercase", func(t *testing.T) {
		assert.Equal(t, "BGR", ColorOrder("bgr").String())
	})
	t.Run("Undefined", func(t *testing.T) {
		assert.Equal(t, "RGB", OrderUndefined.String())
	})
	t.Run("Invalid", func(t *testing.T) {
		assert.Equal(t, "RGB", ColorOrder("xyz").String())
	})
}

func TestParseColorOrder(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		order, err := ParseColorOrder(" bgr ")
		require.NoError(t, err)
		assert.Equal(t, BGR, order)
	})
	t.Run("Invalid", func(t *testing.T) {
		order, err := ParseColorOrder("RRG")
		require.Error(t, err)
		assert.Equal(t, OrderUndefined, order)
	})
	t.Run("Empty", func(t *testing.T) {
		_, err := ParseColorOrder("")
		require.Error(t, err)
	})
}

func TestColorOrder_Marshaling(t *testing.T) {
	// A string type round-trips through both encodings without custom marshalers, which
	// is why it is preferred over the numeric encoding the TensorFlow model info uses.
	t.Run("JSON", func(t *testing.T) {
		b, err := json.Marshal(BGR)
		require.NoError(t, err)
		assert.JSONEq(t, `"BGR"`, string(b))

		var order ColorOrder
		require.NoError(t, json.Unmarshal(b, &order))
		assert.Equal(t, BGR, order)
	})
	t.Run("YAML", func(t *testing.T) {
		b, err := yaml.Marshal(BGR)
		require.NoError(t, err)

		var order ColorOrder
		require.NoError(t, yaml.Unmarshal(b, &order))
		assert.Equal(t, BGR, order)
	})
}

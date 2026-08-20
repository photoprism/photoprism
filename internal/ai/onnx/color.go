package onnx

import (
	"fmt"
	"strings"
)

// ColorOrder names the order in which the red, green, and blue channels are written
// into the input tensor. It is a string so that it round-trips through YAML and JSON
// without custom marshalers and stays readable in a configuration file.
type ColorOrder string

const (
	// OrderUndefined leaves the channel order unspecified, which resolves to RGB.
	OrderUndefined ColorOrder = ""
	// RGB writes the channels in red, green, blue order.
	RGB ColorOrder = "RGB"
	// BGR writes the channels in blue, green, red order.
	BGR ColorOrder = "BGR"
)

// Indices returns the zero-based plane index of the red, green, and blue channels.
// Undefined and invalid orders resolve to RGB.
func (o ColorOrder) Indices() (r, g, b int) {
	r, g, b = 0, 1, 2

	if !o.Valid() {
		return r, g, b
	}

	for i, c := range strings.ToUpper(string(o)) {
		switch c {
		case 'R':
			r = i
		case 'G':
			g = i
		case 'B':
			b = i
		}
	}

	return r, g, b
}

// Valid reports whether the order names each of the three channels exactly once.
func (o ColorOrder) Valid() bool {
	s := strings.ToUpper(string(o))

	if len(s) != Channels {
		return false
	}

	return strings.Count(s, "R") == 1 && strings.Count(s, "G") == 1 && strings.Count(s, "B") == 1
}

// String returns the channel order in uppercase, or RGB when it is undefined or invalid.
func (o ColorOrder) String() string {
	if !o.Valid() {
		return string(RGB)
	}

	return strings.ToUpper(string(o))
}

// ParseColorOrder returns the channel order matching s, or an error when s does not
// name each of the three channels exactly once.
func ParseColorOrder(s string) (ColorOrder, error) {
	o := ColorOrder(strings.ToUpper(strings.TrimSpace(s)))

	if !o.Valid() {
		return OrderUndefined, fmt.Errorf("invalid color order %q", s)
	}

	return o, nil
}

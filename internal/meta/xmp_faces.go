package meta

import (
	"math"
	"strconv"
	"strings"

	"github.com/photoprism/photoprism/pkg/clean"
)

// Face is a named face region parsed from XMP metadata, expressed in
// displayed-orientation top-left normalized coordinates that match
// thumb/crop.Area: X/Y is the top-left corner and W/H the size, all in [0,1].
// Name is empty for regions that carry no assigned person name.
type Face struct {
	Name string
	X    float32
	Y    float32
	W    float32
	H    float32
}

// FaceOptions provides image properties needed to normalize XMP face regions.
type FaceOptions struct {
	Orientation int
	Width       int
	Height      int
}

// Valid reports whether the region has a positive, in-range rectangle.
func (f Face) Valid() bool {
	return f.W > 0 && f.H > 0 && f.X >= 0 && f.Y >= 0 && f.X+f.W <= 1.0001 && f.Y+f.H <= 1.0001
}

// parseFloat32 parses a plain decimal string into a float32; the bool reports
// whether parsing succeeded (distinguishing a real 0 from a parse failure).
func parseFloat32(s string) (float32, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, false
	}
	v, err := strconv.ParseFloat(s, 32)
	if err != nil {
		return 0, false
	}
	return float32(v), true
}

// clampUnit constrains a normalized coordinate to the [0,1] range.
func clampUnit(v float32) float32 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

// rotateRect maps a top-left normalized rectangle into the EXIF-displayed frame.
func rotateRect(x, y, w, h float32, orientation int) (float32, float32, float32, float32) {
	switch orientation {
	case 2: // Mirror horizontal.
		return 1 - x - w, y, w, h
	case 3: // 180°
		return 1 - x - w, 1 - y - h, w, h
	case 4: // Mirror vertical.
		return x, 1 - y - h, w, h
	case 5: // Mirror horizontal and rotate 270° clockwise.
		return y, x, h, w
	case 6: // 90° clockwise
		return 1 - y - h, x, h, w
	case 7: // Mirror horizontal and rotate 90° clockwise.
		return 1 - y - h, 1 - x - w, h, w
	case 8: // 270° clockwise (90° counter-clockwise)
		return y, 1 - x - w, h, w
	default: // 0, 1, and unknown orientations.
		return x, y, w, h
	}
}

// newFace builds a normalized, orientation-corrected Face and reports whether
// the resulting rectangle is valid. Shared by the MWG-RS and Microsoft paths.
func newFace(name string, x, y, w, h float32, orientation int) (Face, bool) {
	if w <= 0 || h <= 0 {
		return Face{}, false
	}

	x, y, w, h = rotateRect(x, y, w, h, orientation)

	left := clampUnit(x)
	top := clampUnit(y)
	right := clampUnit(x + w)
	bottom := clampUnit(y + h)

	f := Face{
		Name: clean.Name(name),
		X:    left,
		Y:    top,
		W:    right - left,
		H:    bottom - top,
	}

	return f, f.Valid()
}

// rotatedRegionSize returns the axis-aligned pixel bounds of a rotated region.
func rotatedRegionSize(w, h, rotation float32) (float32, float32) {
	if rotation == 0 {
		return w, h
	}

	sin, cos := math.Sincos(float64(rotation))

	return float32(math.Abs(float64(w)*cos) + math.Abs(float64(h)*sin)),
		float32(math.Abs(float64(w)*sin) + math.Abs(float64(h)*cos))
}

// normalizeRegionMWG converts an MWG-RS center-based rectangle or circle into
// a displayed-orientation normalized Face.
func normalizeRegionMWG(name string, cx, cy, w, h, diameter float32, unit string, rotation float32, appliedW, appliedH, orientation int) (Face, bool) {
	// Resolve pixel coordinates against the applied dimensions.
	if strings.EqualFold(unit, "pixel") {
		if appliedW <= 0 || appliedH <= 0 {
			return Face{}, false
		}
		cx /= float32(appliedW)
		cy /= float32(appliedH)

		if diameter > 0 {
			w = diameter / float32(appliedW)
			h = diameter / float32(appliedH)
		} else {
			w /= float32(appliedW)
			h /= float32(appliedH)
		}
	} else if diameter > 0 {
		if appliedW <= 0 || appliedH <= 0 {
			return Face{}, false
		}

		pixelDiameter := diameter * float32(min(appliedW, appliedH))
		w = pixelDiameter / float32(appliedW)
		h = pixelDiameter / float32(appliedH)
	}

	if cx < 0 || cx > 1 || cy < 0 || cy > 1 || w <= 0 || h <= 0 {
		return Face{}, false
	}

	if rotation != 0 {
		if appliedW <= 0 || appliedH <= 0 {
			return Face{}, false
		}

		pixelW, pixelH := rotatedRegionSize(w*float32(appliedW), h*float32(appliedH), rotation)
		w = pixelW / float32(appliedW)
		h = pixelH / float32(appliedH)
	}

	// MWG rectangles are center-based; convert to a top-left corner.
	x := cx - w/2
	y := cy - h/2

	return newFace(name, x, y, w, h, orientation)
}

// normalizeRegionMP converts a Microsoft MP:Rectangle ("x, y, w, h", normalized,
// top-left origin) into a displayed-orientation top-left Face. It returns false
// when the value cannot be parsed into four numbers or the rectangle is invalid.
func normalizeRegionMP(name, rectangle string, orientation int) (Face, bool) {
	parts := strings.Split(rectangle, ",")
	if len(parts) != 4 {
		return Face{}, false
	}

	vals := make([]float32, 4)
	for i, p := range parts {
		v, err := strconv.ParseFloat(strings.TrimSpace(p), 32)
		if err != nil {
			return Face{}, false
		}
		vals[i] = float32(v)
	}

	return newFace(name, vals[0], vals[1], vals[2], vals[3], orientation)
}

// DedupFaces collapses regions that describe the same person at the same place
// to a single entry, keeping the first occurrence. The key is the name plus the
// rectangle rounded to four decimals, so genuine duplicate region tags merge
// while two different people whose boxes happen to overlap are both preserved.
func DedupFaces(faces []Face) []Face {
	if len(faces) < 2 {
		return faces
	}

	seen := make(map[string]struct{}, len(faces))
	out := faces[:0:0]

	round := func(v float32) int { return int(math.Round(float64(v) * 10000)) }

	for _, f := range faces {
		key := strings.ToLower(f.Name) + "|" +
			strconv.Itoa(round(f.X)) + "|" + strconv.Itoa(round(f.Y)) + "|" +
			strconv.Itoa(round(f.W)) + "|" + strconv.Itoa(round(f.H))

		if _, ok := seen[key]; ok {
			continue
		}

		seen[key] = struct{}{}
		out = append(out, f)
	}

	return out
}

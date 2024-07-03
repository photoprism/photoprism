package colors

// ColorPerception provides information on how an image looks in terms of color and light.
type ColorPerception struct {
	Colors    Colors
	MainColor Color
	Luminance LightMap
	Chroma    Chroma
}

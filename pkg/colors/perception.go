package colors

// Information on how an image looks like in terms of colors and light.
type ColorPerception struct {
	Colors    Colors
	MainColor Color
	Luminance LightMap
	Chroma    Chroma
}

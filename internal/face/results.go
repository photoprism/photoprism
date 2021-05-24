package face

// Result holds the Result points of the various Result types
type Result struct {
	Rows      int     `json:"rows,omitempty"`
	Cols      int     `json:"cols,omitempty"`
	Face      Point   `json:"face,omitempty"`
	Eyes      []Point `json:"eyes,omitempty"`
	Landmarks []Point `json:"landmarks,omitempty"`
}

func (r *Result) Region() Region {
	return Region{
		RegionAppliedToDimensionsW:    r.Cols,
		RegionAppliedToDimensionsH:    r.Rows,
		RegionAppliedToDimensionsUnit: "pixel",
		RegionName:                    "",
		RegionType:                    "Face",
		RegionAreaX:                   float32(r.Face.Col) / float32(r.Cols),
		RegionAreaY:                   float32(r.Face.Row) / float32(r.Rows),
		RegionAreaW:                   float32(r.Face.Scale) / float32(r.Cols),
		RegionAreaH:                   float32(r.Face.Scale) / float32(r.Cols),
		RegionAreaUnit:                "normalized",
	}
}

// Results is a list of face detection results.
type Results []Result

func (r Results) Regions() (reg Regions) {
	for _, res := range r {
		reg = append(reg, res.Region())
	}

	return reg
}

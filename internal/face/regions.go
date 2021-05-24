package face

// Region represents XMP-compatible face region metadata.
type Region struct {
	RegionAppliedToDimensionsW    int
	RegionAppliedToDimensionsH    int
	RegionAppliedToDimensionsUnit string
	RegionName                    string
	RegionType                    string
	RegionAreaX                   float32
	RegionAreaY                   float32
	RegionAreaW                   float32
	RegionAreaH                   float32
	RegionAreaUnit                string
}

// Regions is a list of face region metadata.
type Regions []Region

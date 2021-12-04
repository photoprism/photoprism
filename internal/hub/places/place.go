package places

// Place represents a region identified by city, state, and country.
type Place struct {
	PlaceID     string `json:"id"`
	LocLabel    string `json:"label"`
	LocDistrict string `json:"district"`
	LocCity     string `json:"city"`
	LocState    string `json:"state"`
	LocCountry  string `json:"country"`
	LocKeywords string `json:"keywords"`
}

package places

// Place represents a region identified by city, state, and country.
type Place struct {
	PlaceID     string `json:"id"`
	LocLabel    string `json:"label"`
	LocDistrict string `json:"district,omitempty"`
	LocCity     string `json:"city,omitempty"`
	LocState    string `json:"state,omitempty"`
	LocCountry  string `json:"country"`
	LocKeywords string `json:"keywords,omitempty"`
}

package places

// Place represents a region identified by city, state, and country.
type Place struct {
	PlaceID     string `json:"id"`
	LocLabel    string `json:"label"`
	LocCity     string `json:"city"`
	LocDistrict string `json:"district"`
	LocState    string `json:"state"`
	LocCountry  string `json:"country"`
	LocKeywords string `json:"keywords"`
}

func NewPlace(id, label, city, district, state, country, keywords string) Place {
	result := Place{
		PlaceID:     id,
		LocLabel:    label,
		LocCity:     city,
		LocDistrict: district,
		LocState:    state,
		LocCountry:  country,
		LocKeywords: keywords,
	}

	return result
}

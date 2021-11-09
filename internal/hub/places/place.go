package places

// Place represents a region identified by city, state, and country.
type Place struct {
	PlaceID     string `json:"id"`
	LocLabel    string `json:"label"`
	LocCity     string `json:"city"`
	LocState    string `json:"state"`
	LocCountry  string `json:"country"`
	LocKeywords string `json:"keywords"`
}

func NewPlace(id, label, city, state, country, keywords string) Place {
	result := Place{
		PlaceID:     id,
		LocLabel:    label,
		LocCity:     city,
		LocState:    state,
		LocCountry:  country,
		LocKeywords: keywords,
	}

	return result
}

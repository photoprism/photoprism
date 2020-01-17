package places

// Place
type Place struct {
	PlaceID    string `json:"id"`
	LocLabel   string `json:"label"`
	LocCity    string `json:"city"`
	LocState   string `json:"state"`
	LocCountry string `json:"country"`
}

func NewPlace(id string, label string, city string, state string, country string) Place {
	result := Place{
		PlaceID:    id,
		LocLabel:   label,
		LocCity:    city,
		LocState:   state,
		LocCountry: country,
	}

	return result
}

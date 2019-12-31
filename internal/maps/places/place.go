package places

// Place
type Place struct {
	PlaceID      string `json:"id"`
	LocLabel   string `json:"label"`
	LocCity    string `json:"city"`
	LocState   string `json:"state"`
	LocCountry string `json:"country"`
}

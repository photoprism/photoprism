package form

// Link represents a sharing link form.
type NewLink struct {
	Password   string `json:"Password"`
	Expires    int    `json:"Expires"`
	CanComment bool   `json:"CanComment"`
	CanEdit    bool   `json:"CanEdit"`
}

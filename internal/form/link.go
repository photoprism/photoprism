package form

// Link represents a sharing link form.
type Link struct {
	Token      string `json:"Token"`
	Password   string `json:"Password"`
	Expires    int    `json:"Expires"`
	CanComment bool   `json:"CanComment"`
	CanEdit    bool   `json:"CanEdit"`
}

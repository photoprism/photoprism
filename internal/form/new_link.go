package form

// Link represents a sharing link form.
type NewLink struct {
	Password   string `json:"password"`
	Expires    int    `json:"expires"`
	CanComment bool   `json:"comment"`
	CanEdit    bool   `json:"edit"`
}

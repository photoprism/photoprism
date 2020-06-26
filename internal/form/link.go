package form

// Link represents a link sharing form.
type Link struct {
	Password     string `json:"Password"`
	ShareToken   string `json:"ShareToken"`
	ShareExpires int    `json:"ShareExpires"`
	MaxViews     uint   `json:"MaxViews"`
	CanComment   bool   `json:"CanComment"`
	CanEdit      bool   `json:"CanEdit"`
}
